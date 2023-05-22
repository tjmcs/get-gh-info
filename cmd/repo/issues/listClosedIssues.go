/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package issues

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjmcs/get-gh-info/cmd"
	"github.com/tjmcs/get-gh-info/cmd/repo"
	"github.com/tjmcs/get-gh-info/utils"
)

// listClosedIssuesCmd represents the 'repo issues listClosed' command
var (
	listClosedIssuesCmd = &cobra.Command{
		Use:   "listClosed",
		Short: "List the closed issues in the named GitHub organization(s)",
		Long: `Constructs a list (sorted by age) of the of the issues in the named
GitHub organization that were closed in the defined time window (skipping any issues
that include the 'backlog' label and only including issues from repositories that are
managed by the named team)`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(listClosedIssueCount())
		},
	}
)

func init() {
	repo.IssuesCmd.AddCommand(listClosedIssuesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
}

/*
 * define the function that is used to list the closed issues in the named GitHub
 * organization(s); note that this function skips closed issues that include the
 * 'backlog' label and only counts issues in repositories that are managed by the
 * named team(s)
 */
func listClosedIssueCount() []map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our query for closed issues
	vars := map[string]interface{}{}
	vars["first"] = githubv4.Int(100)
	vars["type"] = githubv4.SearchTypeIssue
	vars["orderCommentsBy"] = githubv4.IssueCommentOrder{Field: "UPDATED_AT", Direction: "ASC"}
	// and initialize a map that will be used to store the list of closed issues we find
	closedIssueList := []map[string]interface{}{}
	// next, retrieve the list of repositories that are managed by the team we're looking for
	teamName, repositoryList := utils.GetTeamRepos()
	// should we filter out private repositories?
	excludePrivateRepos := viper.GetBool("excludePrivateRepos")
	// retrieve the reference time for our query window
	startDateTime, endDateTime := utils.GetQueryTimeWindow()
	// save date and datetime strings for use in output (below)
	startDateStr := startDateTime.Format(cmd.YearMonthDayFormatStr)
	endDateStr := endDateTime.Format(cmd.YearMonthDayFormatStr)
	startDateTimeStr := startDateTime.Format(cmd.ISO8601_FormatStr)
	endDateTimeStr := endDateTime.Format(cmd.ISO8601_FormatStr)
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
			// of results for each organization (or not)
			firstPage := true
			// and a few other variables that we'll use to query the system for results
			var err error
			var edges issueSearchEdges
			var pageInfo cmd.PageInfo
			// loop over the pages of results until we've reached the end of the list of closed
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
					// define a variable to that references the issue itself
					issue := edge.Node.Issue
					// if the current repository is managed by the team we're interested in, then check to see
					// if we should add this issue to our list of closed issues
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
						// if the is issue was created after the end of our time window, then skip it
						if endDateTime.Before(issue.CreatedAt.Time) {
							continue
						}
						// determine if this issue was created by an internal or external user
						// (i.e., a member of the organization or not)
						creatorIsMember := false
						if issue.AuthorAssociation == "OWNER" ||
							issue.AuthorAssociation == "MEMBER" ||
							issue.AuthorAssociation == "COLLABORATOR" {
							creatorIsMember = true
						}
						// get the list of assignees for this issue
						assigneeList := []string{}
						for _, assignee := range issue.Assignees.Edges {
							assigneeList = append(assigneeList, assignee.Node.Login)
						}
						// determine the age of this issue (the time from when the issue was created to
						// the time it was closed)
						issueAge := issue.ClosedAt.Time.Sub(issue.CreatedAt.Time)
						closedIssueList = append(closedIssueList, map[string]interface{}{
							"createdAt":       issue.CreatedAt.Time,
							"closed":          issue.Closed,
							"closedAt":        issue.ClosedAt.Time,
							"url":             issue.Url,
							"title":           issue.Title,
							"creator":         issue.Author.Login,
							"creatorIsMember": creatorIsMember,
							"company":         issue.Author.User.Company,
							"email":           issue.Author.User.Email,
							"assignees":       strings.Join(assigneeList, ""),
							"age":             utils.JsonDuration{Duration: issueAge},
						})
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
	}
	// print a message indicating the total number of closed issues found
	numOpenIssues := len(closedIssueList)
	fmt.Fprintf(os.Stderr, "\nFound %d closed issues in repositories managed by the '%s' team between %s and %s\n", numOpenIssues,
		teamName, startDateStr, endDateStr)
	// If we have more than one closed issue, then sort the list of closed issues by the age of each issue
	if numOpenIssues > 1 {
		sort.Slice(closedIssueList, func(i, j int) bool {
			return closedIssueList[i]["age"].(utils.JsonDuration).Duration > closedIssueList[j]["age"].(utils.JsonDuration).Duration
		})
	}
	// and return it
	return closedIssueList
}
