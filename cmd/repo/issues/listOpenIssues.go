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
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjmcs/get-gh-info/cmd"
	"github.com/tjmcs/get-gh-info/cmd/repo"
	"github.com/tjmcs/get-gh-info/utils"
)

// listOpenIssuesCmd represents the 'repo issues listOpen' command
var (
	listOpenIssuesCmd = &cobra.Command{
		Use:   "listOpen",
		Short: "List the open issues in the named GitHub organization(s)",
		Long: `Constructs a list (sorted by age) of the of open issues in the named
GitHub organization in the defined time window (skipping any issues that include
the 'backlog' label and only including issues from repositories that are managed
by the named team)`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(listOpenIssueCount())
		},
	}
)

func init() {
	repo.IssuesCmd.AddCommand(listOpenIssuesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	listOpenIssuesCmd.Flags().BoolVarP(&repo.RestrictToTeam, "restrict-to-team", "r", false, "only count comments from immediate team members")
	listOpenIssuesCmd.Flags().BoolVarP(&cmd.SortByTimeToFirstResponse, "by-first-response", "p", false, "sort by the time to first response")
	listOpenIssuesCmd.Flags().BoolVarP(&cmd.SortByStaleness, "by-staleness", "s", false, "sort by the time to last response (staleness)")
	// mark the two "sort by" flags as mutually exclusive
	listOpenIssuesCmd.MarkFlagsMutuallyExclusive("by-first-response", "by-staleness")
	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("restrictToTeam", listOpenIssuesCmd.Flags().Lookup("restrict-to-team"))
	viper.BindPFlag("byFirstReponse", listOpenIssuesCmd.Flags().Lookup("by-first-response"))
	viper.BindPFlag("byStaleness", listOpenIssuesCmd.Flags().Lookup("by-staleness"))
}

/*
 * define the function that is used to list the open issues in the named GitHub
 * organization(s) that were open during the defined timeframe; note that this
 * function skips open issues that include the 'backlog' label and only lists
 * issues in repositories that are managed by the named team(s)
 */
func listOpenIssueCount() []map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our queries for issues
	vars := map[string]interface{}{}
	vars["first"] = githubv4.Int(100)
	vars["type"] = githubv4.SearchTypeIssue
	// set the order in which we want to retrieve the comments for issues based
	// on the value 'byStaleness' flag (if we're going to sort by staleness, then
	// we want to sort in the comments descending order by update time, otherwise
	// we want to sort them in ascending order)
	sortByFirstResponse := viper.GetBool("byFirstReponse")
	sortByStaleness := viper.GetBool("byStaleness")
	if sortByStaleness {
		vars["orderCommentsBy"] = githubv4.IssueCommentOrder{Field: "UPDATED_AT", Direction: "DESC"}
	} else {
		vars["orderCommentsBy"] = githubv4.IssueCommentOrder{Field: "UPDATED_AT", Direction: "ASC"}
	}
	// and initialize a map that will be used to the list of issues that we find
	openIssueList := []map[string]interface{}{}
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
			// loop over the pages of results until we've reached the end of the list of open
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
					// open issue count for the current organization
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
						// determine the age of this issue (which will be the time from when the issue was
						// created to either the time it was closed if it's closed or to the the end
						// of our time window if it's still open
						var age time.Duration
						if issue.Closed {
							age = issue.ClosedAt.Time.Sub(issue.CreatedAt.Time)
						} else {
							age = endDateTime.Sub(issue.CreatedAt.Time)
						}
						// create a map to hold the data for this issue
						issueData := map[string]interface{}{
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
							"age":             utils.JsonDuration{Duration: age},
						}
						// finally, if a flag was set to sort the list of issues by the first response
						// time or staleness time, add that field to our output map
						if sortByFirstResponse {
							firstResponseTime := repo.GetFirstResponseTime(&issue, endDateTime, commentsFromTeamOnly, teamMemberIds)
							issueData["firstResponseTime"] = utils.JsonDuration{Duration: firstResponseTime}
						} else if sortByStaleness {
							stalenessTime := repo.GetLatestResponseTime(&issue, endDateTime, commentsFromTeamOnly, teamMemberIds)
							issueData["staleness"] = utils.JsonDuration{Duration: stalenessTime}
						}
						// and add the issue to the list of open issues
						openIssueList = append(openIssueList, issueData)
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
	// print a message indicating the total number of open issues found
	numOpenIssues := len(openIssueList)
	fmt.Fprintf(os.Stderr, "\nFound %d open issues in repositories managed by the '%s' team between %s and %s\n", numOpenIssues,
		teamName, startDateStr, endDateStr)
	// If we have more than one open issue, then sort the list of open issues by the age of each issue
	if numOpenIssues > 1 {
		var sortField string
		// use the flag that was set to determine which field to sort by (either age, first response time, or staleness)
		if sortByFirstResponse {
			sortField = "firstResponseTime"
		} else if sortByStaleness {
			sortField = "staleness"
		} else {
			sortField = "age"
		}
		// and sort the list of open issues by that field
		sort.Slice(openIssueList, func(i, j int) bool {
			return openIssueList[i][sortField].(utils.JsonDuration).Duration > openIssueList[j][sortField].(utils.JsonDuration).Duration
		})
	}
	// and return it
	return openIssueList
}
