/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package pulls

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjmcs/get-gh-info/cmd"
	"github.com/tjmcs/get-gh-info/cmd/repo"
	"github.com/tjmcs/get-gh-info/utils"
)

// contribSummaryCmd represents the 'contribSummary' command
var (
	listOpenPrsCmd = &cobra.Command{
		Use:   "listOpen",
		Short: "List the open PRs in the named GitHub organization(s)",
		Long: `Constructs a list (sorted by age) of the of open PRs in the named
GitHub organization in the defined time window (skipping any PRs that include
the 'backlog' label and only including PRs from repositories that are managed
by the named team)`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(listOpenPrCount())
		},
	}
)

func init() {
	repo.PullsCmd.AddCommand(listOpenPrsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
}

/*
 * define the function that is used to count the number of open PRs in the
 * named GitHub organization(s); note that this function skips open PRs that
 * include the 'backlog' label and only counts PRs in repositories that are
 * managed by the named team(s)
 */
func listOpenPrCount() []map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our query for PR review contributions
	vars := map[string]interface{}{}
	vars["first"] = githubv4.Int(100)
	vars["type"] = githubv4.SearchTypeIssue
	vars["orderCommentsBy"] = githubv4.IssueCommentOrder{Field: "UPDATED_AT", Direction: "ASC"}
	// and initialize a map that will be used to store counts for each of the named organizations
	// and a total count
	openPrList := []map[string]interface{}{}
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
		// for open PRs that were created before the end of our time window, the second is used
		// to query for closed PRs that were created before the end time and closed after the
		// start time of our query window
		openQuery := githubv4.String(fmt.Sprintf("org:%s type:pr state:open -label:backlog created:<%s",
			orgName, endDateTimeStr))
		closedQuery := githubv4.String(fmt.Sprintf("org:%s type:pr state:closed -label:backlog created:<%s closed:>%s",
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
			var edges prSearchEdges
			var pageInfo cmd.PageInfo
			// loop over the pages of results until we've reached the end of the list of open
			// PRs for this organization
			for {
				// set the "after" field to our current "lastCursor" value
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
					// define a variable to that references the pull request itself
					pullRequest := edge.Node.PullRequest
					// if the current repository is managed by the team we're interested in, then increment the
					// open PR count for the current organization
					if len(pullRequest.Repository.Name) > 0 {
						orgAndRepoName := orgName + "/" + pullRequest.Repository.Name
						idx := utils.FindIndexOf(orgAndRepoName, repositoryList)
						// if the current repository is not managed by the team we're interested in, skip it
						if idx < 0 {
							continue
						}
						// if the repository associated with this PR is private and we're excluding
						// private repositories or if it is archived, then skip it
						if (excludePrivateRepos && pullRequest.Repository.IsPrivate) || pullRequest.Repository.IsArchived {
							continue
						}
						// if the is PR was created after the end of our time window, then skip it
						if endDateTime.Before(pullRequest.CreatedAt.Time) {
							continue
						}
						// determine if this issue was created by an internal or external user
						// (i.e., a member of the organization or not)
						creatorIsMember := false
						if pullRequest.AuthorAssociation == "OWNER" ||
							pullRequest.AuthorAssociation == "MEMBER" ||
							pullRequest.AuthorAssociation == "COLLABORATOR" {
							creatorIsMember = true
						}
						// get the list of assignees for this issue
						assigneeList := []string{}
						for _, assignee := range pullRequest.Assignees.Edges {
							assigneeList = append(assigneeList, assignee.Node.Login)
						}
						// determine the age of this PR (which will be the time from when the PR was
						// created to either the time it was closed if it's closed or to the the end
						// of our time window if it's still open
						var prAge time.Duration
						if pullRequest.Closed {
							prAge = pullRequest.ClosedAt.Time.Sub(pullRequest.CreatedAt.Time)
						} else {
							prAge = endDateTime.Sub(pullRequest.CreatedAt.Time)
						}
						openPrList = append(openPrList, map[string]interface{}{
							"createdAt":       pullRequest.CreatedAt.Time,
							"closed":          pullRequest.Closed,
							"closedAt":        pullRequest.ClosedAt.Time,
							"url":             pullRequest.Url,
							"title":           pullRequest.Title,
							"creator":         pullRequest.Author.Login,
							"creatorIsMember": creatorIsMember,
							"company":         pullRequest.Author.User.Company,
							"email":           pullRequest.Author.User.Email,
							"assignees":       strings.Join(assigneeList, ""),
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
	// print a message indicating the total number of open PRs found
	numOpenPrs := len(openPrList)
	fmt.Fprintf(os.Stderr, "\nFound %d open PRs in repositories managed by the '%s' team between %s and %s\n", numOpenPrs,
		teamName, startDateStr, endDateStr)
	// If we have more than one open PR, then sort the list of open PRs by the age of each PR
	if numOpenPrs > 1 {
		sort.Slice(openPrList, func(i, j int) bool {
			return openPrList[i]["age"].(utils.JsonDuration).Duration > openPrList[j]["age"].(utils.JsonDuration).Duration
		})
	}
	// and return it
	return openPrList
}
