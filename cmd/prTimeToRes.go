/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjmcs/get-gh-info/utils"
)

// contribSummaryCmd represents the 'contribSummary' command
var (
	getPrTimeToResStatsCmd = &cobra.Command{
		Use:   "timeToResolution",
		Short: "Statistics for the time to first response of open isues",
		Long: `Calculates the minimum, first quartile, median, average, third quartile,
and maximum 'time to resolution' for all closed PRs in the named GitHub
organizations and in the defined time window (skipping any PRs that include
the 'backlog' label and only counting PRs in repositories that are managed
by the named team)`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(getPrTimeToResStats())
		},
	}
)

func init() {
	pullsCmd.AddCommand(getPrTimeToResStatsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	getPrTimeToResStatsCmd.Flags().StringVarP(&duration, "lookback-time", "l", "", "'lookback' time window (eg. 10d, 3w, 2m, 1q, 1y)")

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("lookbackTime", getPrTimeToResStatsCmd.Flags().Lookup("lookback-time"))
}

/*
 * define the function that is used to calculate the statistics associated with
 * the "time to first response" for any open PRs in the named GitHub organization(s);
 * note that this function skips open PRs that include the 'backlog' label and only
 * includes first response times for PRs in repositories that are managed by the
 * named team(s)
 */
func getPrTimeToResStats() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our query for PR review contributions
	vars := map[string]interface{}{}
	vars["first"] = githubv4.Int(100)
	vars["type"] = githubv4.SearchTypeIssue
	vars["orderCommentsBy"] = githubv4.IssueCommentOrder{Field: "UPDATED_AT", Direction: "ASC"}
	// next, retrieve the list of repositories that are managed by the team we're looking for
	teamName, repositoryList := utils.GetTeamRepos()
	// define the start and end time of our query window
	startDateTime, endDateTime := utils.GetQueryTimeWindow()
	// save date strings for use in output (below)
	startDateTimeStr := startDateTime.Format("2006-01-02")
	endDateTimeStr := endDateTime.Format("2006-01-02")
	// and initialize a list of durations that will be used to store the time to first
	// response values
	resolutionTimeList := []time.Duration{}
	// loop over the input organization names
	for _, orgName := range utils.GetOrgNameList() {
		// define the query to run for each organization; the query searches for closed
		// PRs that were closed after the start of our time window
		closedQuery := githubv4.String(fmt.Sprintf("org:%s type:pr state:closed -label:backlog closed:%s..%s", orgName,
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
			var edges prSearchEdges
			var pageInfo PageInfo
			// loop over the pages of results from this query until we've reached the end
			// of the list of PRs that matched
			for {
				// run our query and add the data we want from the query results to the
				// repositoryList map
				if firstPage {
					err = client.Query(context.Background(), &firstPrSearchQuery, vars)
				} else {
					err = client.Query(context.Background(), &prSearchQuery, vars)
				}
				if err != nil {
					// Handle error.
					fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
					os.Exit(1)
				}
				// grab out the list of edges and the page info from the results of our search
				// and loop over the edges
				if firstPage {
					edges = firstPrSearchQuery.Search.Edges
					pageInfo = firstPrSearchQuery.Search.PageInfo
					// set firstPage to false so that we'll use the prSearchQuery struct
					// (and it's "after" value) for subsequent queries
					firstPage = false
					fmt.Fprintf(os.Stderr, ".")
				} else {
					edges = prSearchQuery.Search.Edges
					pageInfo = prSearchQuery.Search.PageInfo
					fmt.Fprintf(os.Stderr, ".")
				}
				for _, edge := range edges {
					// if the current repository is managed by the team we're interested in, search for the first
					// response from a member of the team and use the time of that response to calculate the time
					// to first response value for this PR
					if len(edge.Node.PullRequest.Repository.Name) > 0 {
						orgAndRepoName := orgName + "/" + edge.Node.PullRequest.Repository.Name
						idx := utils.FindIndexOf(orgAndRepoName, repositoryList)
						// if the current repository is not managed by the team we're interested in, skip it
						if idx < 0 {
							continue
						}
						// if the repository associated with this PR is private or archived, then skip it
						if edge.Node.PullRequest.Repository.IsPrivate || edge.Node.PullRequest.Repository.IsArchived {
							continue
						}
						// save the time when this issue was closed
						prClosedAt := edge.Node.PullRequest.ClosedAt
						// then save the current PR's creation time
						prCreatedAt := edge.Node.PullRequest.CreatedAt
						// and append the difference (the resolution time) to the list of resolution times
						resolutionTimeList = append(resolutionTimeList, prClosedAt.Time.Sub(prCreatedAt.Time))
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

	// calculate the stats for the slice of PR time to resolution values
	prAgeStats, numClosedPrs := utils.GetJsonDurationStats(resolutionTimeList)
	// print a message indicating how many open PRs were found
	if numClosedPrs == 0 {
		fmt.Fprintf(os.Stderr, "\nWARN: No closed PRs found for the specified organization(s)\n")
	} else {
		fmt.Fprintf(os.Stderr, "\nFound %d closed PRs in repositories managed by the '%s' team between %s and %s\n", numClosedPrs,
			teamName, startDateTimeStr, endDateTimeStr)
	}
	// add return the results as a map
	return map[string]interface{}{"title": "PR Time to Resolution", "start": startDateTimeStr,
		"end": endDateTimeStr, "seriesLength": numClosedPrs, "stats": prAgeStats}
}
