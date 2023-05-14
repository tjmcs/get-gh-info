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
	getFirstResponseTimeStatsCmd = &cobra.Command{
		Use:   "firstResponseTime",
		Short: "Calculates statistics for the time to first response",
		Long: `Calculates the minimum, first quartile, median, average, third quartile,
and maximum time to first response for all open issues in the named GitHub
(skilling any issues that include the 'backlog' label and only counting the
issues in repositories that are managed by the named team.`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(getFirstRespTimeStats())
		},
	}
)

func init() {
	repoCmd.AddCommand(getFirstResponseTimeStatsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
}

/*
 * Define a few types that we can use to define (ane extract data from) the body of the GraphQL
 * query that will be used to retrieve the list of open issues in the named GitHub organization(s)
 */
type respTimeSearchEdges []struct {
	Cursor githubv4.String
	Node   struct {
		Issue struct {
			CreatedAt  githubv4.DateTime
			UpdatedAt  githubv4.DateTime
			Title      string
			Url        string
			Repository struct {
				Name string
				Url  string
			}
			Comments struct {
				Nodes []struct {
					CreatedAt githubv4.DateTime
					UpdatedAt githubv4.DateTime
					Author    struct {
						Login string
					}
					Body string
				}
			} `graphql:"comments(first: 100, orderBy: {field: UPDATED_AT, direction: ASC})"`
		} `graphql:"... on Issue"`
	}
}
type respTimeSearchBody struct {
	IssueCount githubv4.Int
	Edges      respTimeSearchEdges
	PageInfo   PageInfo
}

/*
 * define a pair of structs that can be used to query GitHub for a list of all of the
 * open PRs in a given organization (by name) that match a given query; the first is
 * used to query for the first page of results and the second is used to query for
 * subsequent pages of results
 */
var firstInitRespTimeSearchQuery struct {
	Search struct {
		respTimeSearchBody
	} `graphql:"search(first: $first, query: $query, type: $type)"`
}

var initRespTimeSearchQuery struct {
	Search struct {
		respTimeSearchBody
	} `graphql:"search(first: $first, after: $after, query: $query, type: $type)"`
}

/*
 * define the function that is used to calculate the statistics associated with
 * the "time to first response" for any open issues in the named GitHub organization(s);
 * note that this function skips open issues that include the 'backlog' label and only
 * includes first response times for issues in repositories that are managed by the
 * named team(s)
 */
func getFirstRespTimeStats() map[string]utils.JsonDuration {
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
	// and initialize a list of durations that will be used to store the time to first
	// response values
	firstRespTimeList := []time.Duration{}
	// loop over the input organization names
	for _, orgName := range utils.GetOrgNameList() {
		// construct our query string and add it ot the vars map
		vars["query"] = githubv4.String(fmt.Sprintf("org:%s type:issue state:open -label:backlog", orgName))
		// of results for each organization (or not)
		firstPage := true
		// and a few other variables that we'll use to query the system for results
		var err error
		var edges respTimeSearchEdges
		var pageInfo PageInfo
		// loop over the pages of results until we've reached the end of the list of open
		// issues for this organization
		for {
			// set the "after" field to our current "lastCursor" value
			// run our query and add the data we want from the query results to the
			// repositoryList map
			if firstPage {
				err = client.Query(context.Background(), &firstInitRespTimeSearchQuery, vars)
			} else {
				err = client.Query(context.Background(), &initRespTimeSearchQuery, vars)
			}
			if err != nil {
				// Handle error.
				fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
				os.Exit(1)
			}
			// grab out the list of edges and the page info from the results of our search
			// and loop over the edges
			if firstPage {
				edges = firstInitRespTimeSearchQuery.Search.Edges
				pageInfo = firstInitRespTimeSearchQuery.Search.PageInfo
				// set firstPage to false so that we'll use the initRespTimeSearchQuery struct
				// (and it's "after" value) for subsequent queries
				firstPage = false
				fmt.Fprintf(os.Stderr, ".")
			} else {
				edges = initRespTimeSearchQuery.Search.Edges
				pageInfo = initRespTimeSearchQuery.Search.PageInfo
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
					// if no comments were found for this issue, then use the current time to determine the
					// time to first response
					if len(edge.Node.Issue.Comments.Nodes) == 0 {
						firstRespTimeList = append(firstRespTimeList, time.Now().Sub(issueCreatedAt.Time))
						continue
					}
					// if we got this far, then the current repository is managed by the team we're interested in,
					// so look for the first response from a member of the team
					for _, comment := range edge.Node.Issue.Comments.Nodes {
						if len(comment.Author.Login) > 0 {
							idx := utils.FindIndexOf(comment.Author.Login, teamMemberIds)
							if idx < 0 {
								continue
							}
							// if we got this far, then we've found the first comment made by a member of the team,
							// so calculate the time to first response
							commentCreatedAt := comment.CreatedAt.Time
							firstRespTimeList = append(firstRespTimeList, commentCreatedAt.Sub(issueCreatedAt.Time))
						}
					}
					// if we found a comment from a member of the team, calculate the time to first response
				}
			}
			// if we've reached the end of the list of contributions, break out of the loop
			if !pageInfo.HasNextPage {
				break
			}
			vars["after"] = pageInfo.EndCursor
		}
		// and reset a couple of things to prepare for the next organization
		delete(vars, "after")
	}
	// if we found no issues in our search, then exit with an error message
	if len(firstRespTimeList) == 0 {
		fmt.Fprintf(os.Stderr, "\nWARN: No open issues found for the specified organization(s)\n")
		zeroDuration := utils.JsonDuration{time.Duration(0)}
		return map[string]utils.JsonDuration{"minimum": zeroDuration, "firstQuartile": zeroDuration, "median": zeroDuration,
			"average": zeroDuration, "thirdQuartile": zeroDuration, "maximum": zeroDuration}
	}
	fmt.Fprintf(os.Stderr, "\nFound %d open issues\n", len(firstRespTimeList))
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
