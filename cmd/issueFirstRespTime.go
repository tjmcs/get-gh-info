/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/tjmcs/get-gh-info/utils"
)

// contribSummaryCmd represents the 'contribSummary' command
var (
	getIssueFirstRespTimeStatsCmd = &cobra.Command{
		Use:   "issueFirstRespTime",
		Short: "Statistics for the 'time to first response' of open isues",
		Long: `Calculates the minimum, first quartile, median, average, third quartile,
and maximum 'time to first response' for all open issues in the named GitHub
organizations in the defined time window (skipping issues that include the
'backlog' label and only counting issues in repositories that are managed by
the named team)`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(getIssueFirstRespTimeStats())
		},
	}
)

func init() {
	repoCmd.AddCommand(getIssueFirstRespTimeStatsCmd)

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
func getIssueFirstRespTimeStats() map[string]utils.JsonDuration {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our query for PR review contributions
	vars := map[string]interface{}{}
	vars["first"] = githubv4.Int(100)
	vars["type"] = githubv4.SearchTypeIssue
	// next, retrieve the list of repositories that are managed by the team we're looking for
	teamName, repositoryList := utils.GetTeamRepos()
	// and the details for members of the corresponding team
	_, teamMemberMap := utils.GetTeamMembers(teamName)
	// and from that map, construct list of member logins for that team
	teamMemberIds := utils.GetTeamMemberIds(teamMemberMap)
	// define the start and end time of our query window
	startDateTime, endDateTime := utils.GetQueryTimeWindow()
	// and initialize a list of durations that will be used to store the time to first
	// response values
	firstRespTimeList := []time.Duration{}
	// loop over the input organization names
	for _, orgName := range utils.GetOrgNameList() {
		// define a couple of queries to run for each organization; the first is used to query
		// for open issues that were created before the end of our time window and the second is
		// used to query for closed issues that were both created before and closed after the end
		// of our query window
		openQuery := githubv4.String(fmt.Sprintf("org:%s type:issue state:open -label:backlog created:<%s", orgName, endDateTime.Format("2006-01-02")))
		closedQuery := githubv4.String(fmt.Sprintf("org:%s type:issue state:closed -label:backlog created:<%s closed:>%s", orgName, endDateTime.Format("2006-01-02"), endDateTime.Format("2006-01-02")))
		queries := map[string]githubv4.String{
			"open":   openQuery,
			"closed": closedQuery,
		}
		// loop over the queries that we want to run for this organization, gathering
		// the results for each query
		for queryType, query := range queries {
			// add the query string to use with this query to the vars map
			vars["query"] = query
			// initialize the flag that we use to determine if we're trying to retrieve
			// the first page of results for this query (or not)
			firstPage := true
			// and a few other variables that we'll use to query the system for results
			var err error
			var edges issueSearchEdges
			var pageInfo PageInfo
			// loop over the pages of results from this query until we've reached the end
			// of the list of issues that matched
			for {
				// run our query and add the data we want from the query results to the
				// repositoryList map
				if firstPage {
					err = client.Query(context.Background(), &firstIssueSearchQuery, vars)
				} else {
					err = client.Query(context.Background(), &issueSearchQuery, vars)
				}
				if err != nil {
					// Handle error.
					fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
					os.Exit(1)
				}
				// grab out the list of edges and the page info from the results of our search
				// and loop over the edges
				if firstPage {
					edges = firstIssueSearchQuery.Search.Edges
					pageInfo = firstIssueSearchQuery.Search.PageInfo
					// set firstPage to false so that we'll use the issueSearchQuery struct
					// (and it's "after" value) for subsequent queries
					firstPage = false
					fmt.Fprintf(os.Stderr, ".")
				} else {
					edges = issueSearchQuery.Search.Edges
					pageInfo = issueSearchQuery.Search.PageInfo
					fmt.Fprintf(os.Stderr, ".")
				}
				for _, edge := range edges {
					// if the current repository is managed by the team we're interested in, search for the first
					// response from a member of the team and use the time of that response to calculate the time
					// to first response value for this issue
					if len(edge.Node.Issue.Repository.Name) > 0 {
						orgAndRepoName := orgName + "/" + edge.Node.Issue.Repository.Name
						idx := utils.FindIndexOf(orgAndRepoName, repositoryList)
						// if the current repository is not managed by the team we're interested in, skip it
						if idx < 0 {
							continue
						}
						// save the current issue's creation time
						issueCreatedAt := edge.Node.Issue.CreatedAt
						// if this is a query for closed issues and the issue was closed before the start of
						// our query window, then skip it
						if queryType == "closed" {
							if edge.Node.Issue.ClosedAt.Before(startDateTime.Time) {
								continue
							}
						}
						// if no comments were found for this issue, then use the end of our query window
						// to determine the time to first response
						if len(edge.Node.Issue.Comments.Nodes) == 0 {
							firstRespTimeList = append(firstRespTimeList, endDateTime.Time.Sub(issueCreatedAt.Time))
							continue
						}
						// if we got this far, then the current repository is managed by the team we're interested in,
						// so look for the first response from a member of the team; first, initialize a variable to
						// hold the difference between the end of our query window and the creation time for this issue
						firstRespTime := endDateTime.Time.Sub(issueCreatedAt.Time)
						for _, comment := range edge.Node.Issue.Comments.Nodes {
							// if the comment has an author (it should)
							if len(comment.Author.Login) > 0 {
								// use that login to see if this comment was from a team member
								idx := utils.FindIndexOf(comment.Author.Login, teamMemberIds)
								// if the comment was not from a team member, skip it
								if idx < 0 {
									continue
								}
								// if the comment was created after the end of our query window, then we've
								// reached the end of the time where a user could have responded within our
								// time window, so just use the end of the query window to determine the time
								// to first response and break out of the loop
								if comment.CreatedAt.After(endDateTime.Time) {
									firstRespTime = endDateTime.Time.Sub(issueCreatedAt.Time)
									break
								}
							}
						}
						// and append this first response time to the list of first response times
						firstRespTimeList = append(firstRespTimeList, firstRespTime)
						// if we found a comment from a member of the team, calculate the time to first response
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

	// if we found no issues in our search, then exit with an error message
	if len(firstRespTimeList) == 0 {
		fmt.Fprintf(os.Stderr, "\nWARN: No open issues found for the specified organization(s)\n")
		zeroDuration := utils.JsonDuration{time.Duration(0)}
		return map[string]utils.JsonDuration{"minimum": zeroDuration, "firstQuartile": zeroDuration, "median": zeroDuration,
			"average": zeroDuration, "thirdQuartile": zeroDuration, "maximum": zeroDuration}
	}
	fmt.Fprintf(os.Stderr, "\nFound %d open issues in repositories managed by the '%s' team before %s\n", len(firstRespTimeList),
		teamName, endDateTime.Format("2006-01-02"))
	// now, sort the resulting list of durations from greatest to least
	sort.Slice(firstRespTimeList, func(i, j int) bool {
		return firstRespTimeList[i] > firstRespTimeList[j]
	})
	// from the sorted slice, find the minimum, first quartile, the median, average, third quartile,
	// and maximum values
	min := utils.JsonDuration{firstRespTimeList[len(firstRespTimeList)-1]}
	firstQuartile := utils.JsonDuration{firstRespTimeList[(len(firstRespTimeList)*3)/4]}
	median := utils.JsonDuration{firstRespTimeList[len(firstRespTimeList)/2]}
	avg := utils.JsonDuration{utils.GetAverageDuration(firstRespTimeList)}
	thirdQuartile := utils.JsonDuration{firstRespTimeList[len(firstRespTimeList)/4]}
	max := utils.JsonDuration{firstRespTimeList[0]}
	// add return the results as a map
	return map[string]utils.JsonDuration{"minimum": min, "firstQuartile": firstQuartile, "median": median,
		"average": avg, "thirdQuartile": thirdQuartile, "maximum": max}
}