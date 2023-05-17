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
	getPrStalenessStatsCmd = &cobra.Command{
		Use:   "staleness",
		Short: "Statistics for the time since the last response for open PRs",
		Long: `Calculates the minimum, first quartile, median, average, third quartile,
third quartile and maximum 'time since the last response' or 'staleness'
for all closed PRs in the named GitHub organizations and in the defined
time window (skipping any PRs that include the 'backlog' label and only
counting PRs in repositories that are managed by the named team)`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(getPrStalenessStats())
		},
	}
)

func init() {
	pullsCmd.AddCommand(getPrStalenessStatsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	getPrStalenessStatsCmd.Flags().BoolVarP(&restrictToTeam, "restrict-to-team", "r", false, "only count comments from immediate team members")

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("restrictToTeam", getPrStalenessStatsCmd.Flags().Lookup("restrict-to-team"))
}

/*
 * define the function that is used to calculate the statistics associated with
 * the "staleness time" for any open PRs in the named GitHub organization(s);
 * note that this function skips open PRs that include the 'backlog' label and only
 * includes first response times for PRs in repositories that are managed by the
 * named team(s)
 */
func getPrStalenessStats() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our query for PR review contributions
	vars := map[string]interface{}{}
	vars["first"] = githubv4.Int(100)
	vars["type"] = githubv4.SearchTypeIssue
	vars["orderCommentsBy"] = githubv4.IssueCommentOrder{Field: "UPDATED_AT", Direction: "DESC"}
	// next, retrieve the list of repositories that are managed by the team we're looking for
	teamName, repositoryList := utils.GetTeamRepos()
	// should we filter out private repositories?
	excludePrivateRepos := viper.GetBool("excludePrivateRepos")
	// should we only count comments from immediate team members?
	commentsFromTeamOnly := viper.GetBool("restrictToTeam")
	teamMemberIds := []string{}
	if !commentsFromTeamOnly {
		// and the details for members of the corresponding team
		_, teamMemberMap := utils.GetTeamMembers(teamName)
		// and from that map, construct list of member logins for that team
		teamMemberIds = utils.GetTeamMemberIds(teamMemberMap)
	}
	// retrieve the reference time for our query window
	refDateTime, _ := utils.GetQueryTimeWindow()
	// save date strings for use in output (below)
	refDateTimeStr := refDateTime.Format("2006-01-02")
	// and initialize a list of durations that will be used to store the time to first
	// response values
	stalenessTimeList := []time.Duration{}
	// loop over the input organization names
	for _, orgName := range utils.GetOrgNameList() {
		// define a couple of queries to run for each organization; the first is used to query
		// for open PRs that were created before the end of our time window and the second is
		// used to query for closed PRs that were both created before and closed after the end
		// of our query window
		openQuery := githubv4.String(fmt.Sprintf("org:%s type:pr state:open -label:backlog created:<%s",
			orgName, refDateTimeStr))
		closedQuery := githubv4.String(fmt.Sprintf("org:%s type:pr state:closed -label:backlog created:<%s closed:>%s",
			orgName, refDateTimeStr, refDateTimeStr))
		queries := map[string]githubv4.String{
			"open":   openQuery,
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
						// if the repository associated with this issue is private and we're excluding
						// private repositories or if it is archived, then skip it
						if (excludePrivateRepos && edge.Node.PullRequest.Repository.IsPrivate) || edge.Node.PullRequest.Repository.IsArchived {
							continue
						}
						// save the current PR's creation time
						prCreatedAt := edge.Node.PullRequest.CreatedAt
						// if we got this far, then the current repository is managed by the team we're interested
						// in, so initialize a variable to store the default staleness time for this PR (here, the
						// default is the time between the PR's creation and the end of our query window)
						stalenessTime := refDateTime.Time.Sub(prCreatedAt.Time)
						// if no comments were found for this PR, then use the default staleness time
						if len(edge.Node.PullRequest.Comments.Nodes) == 0 {
							stalenessTimeList = append(stalenessTimeList, stalenessTime)
							continue
						}
						// loop over the comments for this PR, looking for the first comment from a team member
						for _, comment := range edge.Node.PullRequest.Comments.Nodes {
							// if this comment was created after the reference time, then skip it
							if comment.CreatedAt.After(refDateTime.Time) {
								continue
							}
							// if the comment has an author (it should)
							if len(comment.Author.Login) > 0 {
								// if the flag to only count comments from the immediate team was
								// set, then only count comments from immediate team members
								if commentsFromTeamOnly {
									// if here, looking only for comments only from immediate team members,
									// so if this comment is not from an immediate team member skip it
									idx := utils.FindIndexOf(comment.Author.Login, teamMemberIds)
									if idx < 0 {
										continue
									}
								} else {
									// otherwise (by default), we're looking for comments from anyone who is
									// an owner of this repository, a member of the organization that owns this
									// repository, or collaborator on this repository; if that's not the case
									// for this comment, then skip it
									if comment.AuthorAssociation != "OWNER" &&
										comment.AuthorAssociation != "MEMBER" &&
										comment.AuthorAssociation != "COLLABORATOR" {
										continue
									}
								}
								// if get here, then we've found a comment from a member of the team,
								// so use the time the comment was created to calculate a staleness
								// value for this issue
								if edge.Node.PullRequest.Closed {
									// if the PR is closed, then use the time the PR was closed
									// to determine hte staleness time
									stalenessTime = edge.Node.PullRequest.ClosedAt.Time.Sub(comment.CreatedAt.Time)
								} else {
									// otherwise use the reference time
									stalenessTime = refDateTime.Sub(comment.CreatedAt.Time)
								}
								break
							}
						}
						// and append this first response time to the list of first response times
						stalenessTimeList = append(stalenessTimeList, stalenessTime)
						// if we found a comment from a member of the team, calculate the staleness time
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

	// calculate the stats for the slice of PR staleness values
	prStalenessTimeStats, numOpenPrs := utils.GetJsonDurationStats(stalenessTimeList)
	// print a message indicating how many open PRs were found
	if numOpenPrs == 0 {
		fmt.Fprintf(os.Stderr, "\nWARN: No open PRs found for the specified organization(s)\n")
	} else {
		fmt.Fprintf(os.Stderr, "\nFound %d open PRs in repositories managed by the '%s' team on %s\n", numOpenPrs,
			teamName, refDateTimeStr)
	}
	// add return the results as a map
	return map[string]interface{}{"title": "Open PR Staleness Time", "refDate": refDateTimeStr,
		"seriesLength": numOpenPrs, "stats": prStalenessTimeStats}
}
