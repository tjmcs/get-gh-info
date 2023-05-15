/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/tjmcs/get-gh-info/utils"
)

// contribSummaryCmd represents the 'contribSummary' command
var (
	getOpenIssuesCmd = &cobra.Command{
		Use:   "countOpenIssues",
		Short: "Count of open issues in the named GitHub organization(s)",
		Long: `Determines the number of open issues in the named GitHub organizations
and in the defined time window (skipping any issues that include the
'backlog' label and only counting issues in repositories that are managed
by the named team)`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(getOpenIssueCount())
		},
	}
)

func init() {
	repoCmd.AddCommand(getOpenIssuesCmd)

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
type issueSearchEdges []struct {
	Cursor githubv4.String
	Node   struct {
		Issue struct {
			CreatedAt  githubv4.DateTime
			UpdatedAt  githubv4.DateTime
			Closed     bool
			ClosedAt   githubv4.DateTime
			Title      string
			Url        string
			Repository struct {
				Name string
				Url  string
			}
			Assignees struct {
				Edges []struct {
					Node struct {
						Login string
					}
				}
			} `graphql:"assignees(first: 10)"`
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
type issueSearchBody struct {
	IssueCount githubv4.Int
	Edges      issueSearchEdges
	PageInfo   PageInfo
}

/*
 * define a pair of structs that can be used to query GitHub for a list of all of the
 * open PRs in a given organization (by name) that match a given query; the first is
 * used to query for the first page of results and the second is used to query for
 * subsequent pages of results
 */
var firstIssueSearchQuery struct {
	Search struct {
		issueSearchBody
	} `graphql:"search(first: $first, query: $query, type: $type)"`
}

var issueSearchQuery struct {
	Search struct {
		issueSearchBody
	} `graphql:"search(first: $first, after: $after, query: $query, type: $type)"`
}

/*
 * define the function that is used to count the number of open issues in the
 * named GitHub organization(s); note that this function skips open issues that
 * include the 'backlog' label and only counts issues in repositories that are
 * managed by the named team(s)
 */
func getOpenIssueCount() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our query for PR review contributions
	vars := map[string]interface{}{}
	vars["first"] = githubv4.Int(100)
	vars["type"] = githubv4.SearchTypeIssue
	// and initialize a counter that will be used to count the number of open issues
	// in the named GitHub organization(s)
	openIssueCount := 0
	// and initialize a map that will be used to store counts for each of the named organizations
	// and a total count
	openIssueCountMap := map[string]interface{}{}
	// next, retrieve the list of repositories that are managed by the team we're looking for
	teamName, repositoryList := utils.GetTeamRepos()
	// define the start and end time of our query window
	startDateTime, endDateTime := utils.GetQueryTimeWindow()
	// loop over the input organization names
	for _, orgName := range utils.GetOrgNameList() {
		// initialize a counter for the number of open issues in the current organization
		orgOpenIssueCount := 0
		// define a couple of queries to run for each organization; the first is used to query
		// for open PRs that were created before the end of our time window and the second is
		// used to query for closed PRs that were both created before and closed after the end
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
			// of results for each organization (or not)
			firstPage := true
			// and a few other variables that we'll use to query the system for results
			var err error
			var edges issueSearchEdges
			var pageInfo PageInfo
			// loop over the pages of results until we've reached the end of the list of open
			// issues for this organization
			for {
				// set the "after" field to our current "lastCursor" value
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
					// if the current repository is managed by the team we're interested in, then increment the
					// open issue count for the current organization
					if len(edge.Node.Issue.Repository.Name) > 0 {
						orgAndRepoName := orgName + "/" + edge.Node.Issue.Repository.Name
						idx := utils.FindIndexOf(orgAndRepoName, repositoryList)
						// if the current repository is not managed by the team we're interested in, skip it
						if idx < 0 {
							continue
						}
						// save the current issue's creation time
						issueCreatedAt := edge.Node.Issue.CreatedAt
						// if the issue was created after the end of our query window, then skip it
						if issueCreatedAt.After(endDateTime.Time) {
							continue
						}
						// if this is a closed issue and it was closed before the start of our query window,
						// then skip it
						if queryType == "closed" {
							if edge.Node.Issue.ClosedAt.Before(startDateTime.Time) {
								continue
							}
						}
						orgOpenIssueCount++
						openIssueCount++
					}
				}
				// if we've reached the end of the list of contributions, break out of the loop
				if !pageInfo.HasNextPage {
					break
				}
				vars["after"] = pageInfo.EndCursor
			}
			// and unset the "after" key in the vars map so that we're ready
			// for the next query
			delete(vars, "after")
		} // end of loop over queries

		// add the open issue count for the current organization to the openIssueCountMap
		openIssueCountMap[orgName] = orgOpenIssueCount
	}
	// add the total open issue count to the openIssueCountMap
	openIssueCountMap["total"] = openIssueCount
	fmt.Fprintf(os.Stderr, "\nFound %d open issues in repositories managed by the '%s' team before %s\n", openIssueCount,
		teamName, endDateTime.Format("2006-01-02"))
	return openIssueCountMap
}
