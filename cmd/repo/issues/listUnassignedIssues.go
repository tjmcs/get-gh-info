/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package issues

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjmcs/get-gh-info/cmd"
	"github.com/tjmcs/get-gh-info/cmd/repo"
	"github.com/tjmcs/get-gh-info/utils"
)

// listUnassignedIssuesCmd represents the 'repo issues listUnassigned' command
var (
	listUnassignedIssuesCmd = &cobra.Command{
		Use:   "listUnassigned",
		Short: "List the unassigned and open issues in the named GitHub organization(s)",
		Long: `Constructs a list (sorted by age) of the of issues that are both open and
unassigned in the named GitHub organization and defined time window (skipping
any issues that include the 'backlog' label and only including PRs from the
repositories that are managed by the named team)`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(listUnassignedIssueCount())
		},
	}
)

func init() {
	repo.IssuesCmd.AddCommand(listUnassignedIssuesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
}

/*
 * define the function that is used to list the open issues in the named GitHub
 * organization(s) that were open during the defined timeframe and that had not
 * been assigned to anyone at that time; note that this function skips open issues
 * that include the 'backlog' label and only lists issues in repositories that are
 * managed by the named team(s)
 */
func listUnassignedIssueCount() []map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our queries for issues
	vars := map[string]interface{}{}
	vars["first"] = githubv4.Int(100)
	vars["type"] = githubv4.SearchTypeIssue
	vars["orderCommentsBy"] = githubv4.IssueCommentOrder{Field: "UPDATED_AT", Direction: "ASC"}
	// and initialize a map that will be used to store the list of unassigned issues
	// that we find
	unassignedPrList := []map[string]interface{}{}
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
		// define a couple of queries to run for each organization; the first is used to query
		// for open issues that were created before the end of our time window, the second is
		// used to query for closed issues that were created before the end time and closed after
		// the start time of our query window
		openQuery := githubv4.String(fmt.Sprintf("org:%s type:issue state:open -label:backlog created:<%s",
			orgName, endDateTimeStr))
		closedQuery := githubv4.String(fmt.Sprintf("org:%s type:issue state:closed -label:backlog created:<%s closed:>%s",
			orgName, endDateTimeStr, startDateTimeStr))
		queries := map[string]githubv4.String{
			"open":   openQuery,
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
			var edges repo.IssueSearchEdges
			var pageInfo cmd.PageInfo
			// loop over the pages of results until we've reached the end of the list of unassigned
			// issues for this organization
			for {
				// set the "after" field to our current "lastCursor" value
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
					// set firstPage to false so that we'll use the IssueSearchQuery struct
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
					// if the current repository is managed by the team we're interested in, then increment the
					// unassigned issue count for the current organization
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
						// if someone has been assigned to this issue, then skip it
						if len(issue.Assignees.Edges) > 0 {
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
						// determine the age of this issue (which will be the time from when the issue was
						// created to either the time it was closed if it's closed or to the the end
						// of our time window if it's still open)
						var prAge time.Duration
						if issue.Closed {
							prAge = issue.ClosedAt.Time.Sub(issue.CreatedAt.Time)
						} else {
							prAge = endDateTime.Sub(issue.CreatedAt.Time)
						}
						unassignedPrList = append(unassignedPrList, map[string]interface{}{
							"createdAt":       issue.CreatedAt.Time,
							"closed":          issue.Closed,
							"closedAt":        issue.ClosedAt.Time,
							"url":             issue.Url,
							"title":           issue.Title,
							"creator":         issue.Author.Login,
							"creatorIsMember": creatorIsMember,
							"company":         issue.Author.User.Company,
							"email":           issue.Author.User.Email,
							"assignees":       "",
							"age":             utils.JsonDuration{Duration: prAge},
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
	// print a message indicating the total number of unassigned issues found
	numUnassignedIssues := len(unassignedPrList)
	fmt.Fprintf(os.Stderr, "\nFound %d unassigned issues in repositories managed by the '%s' team between %s and %s\n", numUnassignedIssues,
		teamName, startDateStr, endDateStr)
	// If we have more than one unassigned issue, then sort the list of unassigned issues by the age of each issue
	if numUnassignedIssues > 1 {
		sort.Slice(unassignedPrList, func(i, j int) bool {
			return unassignedPrList[i]["age"].(utils.JsonDuration).Duration > unassignedPrList[j]["age"].(utils.JsonDuration).Duration
		})
	}
	// and return it
	return unassignedPrList
}
