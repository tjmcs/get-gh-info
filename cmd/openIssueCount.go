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
		Short: "Counts the number of open issues in the named GitHub organization(s)",
		Long: `Determines the number of open issues in the named GitHub organizations,
skipping any issues that include the 'backlog' label and only counting issues
in repositories that are managed by the named team.`,
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
	_, repositoryList := utils.GetTeamRepos()
	// loop over the input organization names
	for _, orgName := range utils.GetOrgNameList() {
		// initialize a counter for the number of open issues in the current organization
		orgOpenIssueCount := 0
		// construct our query string and add it ot the vars map
		vars["query"] = githubv4.String(fmt.Sprintf("org:%s type:issue state:open -label:backlog", orgName))
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
				fmt.Fprintf(os.Stderr, "Found first %d open issues (of %d in the %s org)\n", len(edges), firstIssueSearchQuery.Search.IssueCount, orgName)
			} else {
				edges = issueSearchQuery.Search.Edges
				pageInfo = issueSearchQuery.Search.PageInfo
				fmt.Fprintf(os.Stderr, "Found next %d open issues (of %d in the %s org))\n", len(edges), issueSearchQuery.Search.IssueCount, orgName)
			}
			for _, edge := range edges {
				// if the current repository is managed by the team we're interested in, then increment the
				// open issue count for the current organization
				if len(edge.Node.Issue.Repository.Name) > 0 {
					orgAndRepoName := orgName + "/" + edge.Node.Issue.Repository.Name
					idx := utils.FindIndexOf(orgAndRepoName, repositoryList)
					if idx > 0 {
						orgOpenIssueCount++
						openIssueCount++
					}
				}
			}
			// if we've reached the end of the list of contributions, break out of the loop
			if !pageInfo.HasNextPage {
				break
			}
			vars["after"] = pageInfo.EndCursor
		}
		// add the open issue count for the current organization to the openIssueCountMap
		openIssueCountMap[orgName] = orgOpenIssueCount
		// and reset a couple of things to prepare for the next organization
		delete(vars, "after")
	}
	// add the total open issue count to the openIssueCountMap
	openIssueCountMap["total"] = openIssueCount
	return openIssueCountMap
}
