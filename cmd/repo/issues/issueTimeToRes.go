/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package issues

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjmcs/get-gh-info/cmd"
	"github.com/tjmcs/get-gh-info/cmd/repo"
	"github.com/tjmcs/get-gh-info/utils"
)

// getTimeToResStatsCmd represents the 'repo issues timeToResolution' command
var (
	getTimeToResStatsCmd = &cobra.Command{
		Use:   "timeToResolution",
		Short: "Statistics for the 'time to resolution' of closed isues",
		Long: `Calculates the minimum, first quartile, median, average, third quartile,
and maximum 'time to resolution' for all closed issues in the named GitHub
organizations and in the defined time window (skipping any issues that include
the 'backlog' label and only counting issues in repositories that are managed
by the named team)`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(getTimeToResStats())
		},
	}
)

func init() {
	repo.IssuesCmd.AddCommand(getTimeToResStatsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
}

/*
 * define the function that is used to calculate the statistics associated with
 * the "time to first response" for any open issues in the named GitHub organization(s);
 * note that this function skips open issues that include the 'backlog' label and only
 * includes first response times for issues in repositories that are managed by the
 * named team(s)
 */
func getTimeToResStats() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our query for closed issues
	vars := map[string]interface{}{}
	vars["first"] = githubv4.Int(100)
	vars["type"] = githubv4.SearchTypeIssue
	vars["orderCommentsBy"] = githubv4.IssueCommentOrder{Field: "UPDATED_AT", Direction: "ASC"}
	// next, retrieve the list of repositories that are managed by the team we're looking for
	teamName, repositoryList := utils.GetTeamRepos()
	// should we filter out private repositories?
	excludePrivateRepos := viper.GetBool("excludePrivateRepos")
	// retrieve the start and end time for our query window
	startDateTime, endDateTime := utils.GetQueryTimeWindow()
	// save date and datetime strings for use in output (below)
	startDateStr := startDateTime.Format(cmd.YearMonthDayFormatStr)
	endDateStr := endDateTime.Format(cmd.YearMonthDayFormatStr)
	startDateTimeStr := startDateTime.Format(cmd.ISO8601_FormatStr)
	endDateTimeStr := endDateTime.Format(cmd.ISO8601_FormatStr)
	// and initialize a slice of durations that will be used to store the time to first
	// response values
	resolutionTimeList := []time.Duration{}
	// loop over the input organization names
	for _, orgName := range utils.GetOrgNameList() {
		// define the query to run for each organization; this query looks for closed issues
		// that were closed within the defined time window
		closedQuery := githubv4.String(fmt.Sprintf("org:%s type:issue state:closed -label:backlog closed:%s..%s", orgName,
			startDateTimeStr, endDateTimeStr))
		queries := map[string]githubv4.String{
			"closed": closedQuery,
		}
		// loop over the queries that we want to run for this organization, gathering
		// the results for each query
		for _, query := range queries {
			// add the query string to use with this query to the vars map
			vars["query"] = query
			// initialize the flag that we use to determine if we're trying to retrieve
			// the first page of results for this query (or not)
			firstPage := true
			// and a few other variables that we'll use to query the system for results
			var err error
			var edges repo.IssueSearchEdges
			var pageInfo cmd.PageInfo
			// loop over the pages of results from this query until we've reached the end
			// of the list of issues that matched
			for {
				// run our query and add the data we want from the query results to the
				// repositoryList map
				if firstPage {
					err = client.Query(context.Background(), &repo.FirstIssueSearchQuery, vars)
				} else {
					err = client.Query(context.Background(), &repo.IssueSearchQuery, vars)
				}
				if err != nil {
					// Handle error.
					fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
					os.Exit(1)
				}
				// grab out the list of edges and the page info from the results of our search
				// and loop over the edges
				if firstPage {
					edges = repo.FirstIssueSearchQuery.Search.Edges
					pageInfo = repo.FirstIssueSearchQuery.Search.PageInfo
					// set firstPage to false so that we'll use the repo.IssueSearchQuery struct
					// (and it's "after" value) for subsequent queries
					firstPage = false
					fmt.Fprintf(os.Stderr, ".")
				} else {
					edges = repo.IssueSearchQuery.Search.Edges
					pageInfo = repo.IssueSearchQuery.Search.PageInfo
					fmt.Fprintf(os.Stderr, ".")
				}
				for _, edge := range edges {
					// define a variable to that references the pull request itself
					issue := edge.Node.Issue
					// if the current repository is managed by the team we're interested in, search for the first
					// response from a member of the team and use the time of that response to calculate the time
					// to first response value for this issue
					if len(issue.Repository.Name) > 0 {
						orgAndRepoName := orgName + "/" + issue.Repository.Name
						idx := utils.FindIndexOf(orgAndRepoName, repositoryList)
						// if the current repository is not managed by the team we're interested in, skip it
						if idx < 0 {
							continue
						}
						// if the repository associated with this issue is private and we're excluding
						// private repositories or if it is archived, then skip it
						if (excludePrivateRepos && issue.Repository.IsPrivate) || issue.Repository.IsArchived {
							continue
						}
						// save the time when this issue was closed
						issueClosedAt := issue.ClosedAt
						// then save the current issue's creation time
						issueCreatedAt := issue.CreatedAt
						// if the is issue was created after the end of our time window, then skip it
						if endDateTime.Before(issueCreatedAt.Time) {
							continue
						}
						// and append the difference (the resolution time) to the list of resolution times
						resolutionTimeList = append(resolutionTimeList, issueClosedAt.Time.Sub(issueCreatedAt.Time))
					}
				}
				// if we've reached the end of the list of contributions, break out of the loop
				if !pageInfo.HasNextPage {
					break
				}
				// set the "after" field to the "EndCursor" from the pageInfo structure so
				// we will get the next page of results when we run the query again
				vars["after"] = pageInfo.EndCursor
			}
			// and unset the "after" key in the vars map so that we're ready
			// for the next query
			delete(vars, "after")
		} // end of loop over queries
	} // end of loop over organizations

	// calculate the stats for the slice of issue time to resolution values
	prAgeStats, numClosedIssues := utils.GetJsonDurationStats(resolutionTimeList)
	// print a message indicating how many open issues were found
	if numClosedIssues == 0 {
		fmt.Fprintf(os.Stderr, "\nWARN: No closed issues found for the specified organization(s)\n")
	} else {
		fmt.Fprintf(os.Stderr, "\nFound %d closed issues in repositories managed by the '%s' team between %s and %s\n", numClosedIssues,
			teamName, startDateStr, endDateStr)
	}
	// add return the results as a map
	return map[string]interface{}{"title": "Issue Time to Resolution", "start": startDateTimeStr,
		"end": endDateTimeStr, "seriesLength": numClosedIssues, "stats": prAgeStats}
}
