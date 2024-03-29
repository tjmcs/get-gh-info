/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package user

import (
	"context"
	"fmt"
	"os"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/tjmcs/get-gh-info/cmd"
	"github.com/tjmcs/get-gh-info/utils"
)

// contribsCmd represents the 'contribs' command
var (
	contribsCmd = &cobra.Command{
		Use:   "contribs",
		Short: "Generates a list of commits and PRs made",
		Long: `Constructs a list of any contributions made (commits and PRs) by each of
the input users against any of the repositories in the named set of GitHub
organizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(contribs())
		},
	}
)

func init() {
	cmd.UserCmd.AddCommand(contribsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

// define the struct that we'll use to determine the total contributions from
// each of the input usernames to each of the input organizations
var ContribQuery struct {
	User struct {
		Login                   string
		ContributionsCollection struct {
			TotalIssueContributions                            int
			TotalRepositoriesWithContributedIssues             int
			TotalCommitContributions                           int
			TotalRepositoriesWithContributedCommits            int
			TotalPullRequestContributions                      int
			TotalRepositoriesWithContributedPullRequests       int
			TotalPullRequestReviewContributions                int
			TotalRepositoriesWithContributedPullRequestReviews int
		} `graphql:"contributionsCollection(from: $from, to: $to, organizationID: $organizationID)"`
	} `graphql:"user(login: $login)"`
}

/*
 * define a few types that we'll be using in some of our query structs, first
 * a struct for the contributions made to each repository by a given user
 */
type contributionEdges struct {
	Cursor githubv4.String
	Node   struct {
		Repository struct {
			Name string
			Url  string
		}
		CommitCount int
		OccurredAt  githubv4.DateTime
	}
}

/*
 * and a struct that can be used to put together a list of all of the
 * contributions made by a given user to any repository in a given organization
 */
var contributionsMadeQuery struct {
	User struct {
		Login                   string
		ContributionsCollection struct {
			CommitContributionsByRepository []struct {
				Contributions struct {
					Edges []contributionEdges
				} `graphql:"contributions(first: $first, after: $after)"`
			} `graphql:"commitContributionsByRepository(maxRepositories: 100)"`
		} `graphql:"contributionsCollection(from: $from, to: $to, organizationID: $organizationID)"`
	} `graphql:"user(login: $login)"`
}

/*
 * define the function that is used to fetch the GitHub contribution information
 * for the contributions made by the named user(s) against repositories under
 * the named org(s)
 */
func contribs() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// and then get the list of organization IDs that we want to query
	orgIdList := utils.GetOrgIdList(client)
	// define the start and end time of our query window
	startDateTime, endDateTime := utils.GetQueryTimeWindow()
	// define a list that we'll use later on (to loop over the team members and
	// to skip some members when calculating statistics)
	gitHubIdList := utils.GetUserIdList()
	// initialize the vars map that we'll use when making our query for PR contributions
	vars := map[string]interface{}{
		"from":  startDateTime,
		"to":    endDateTime,
		"first": githubv4.Int(100),
	}
	contribsByUser := map[string]interface{}{}
	contribsByRepo := map[string]interface{}{}
	// loop over the list of GitHub IDs
	for _, gitHubId := range gitHubIdList {
		// set the login value for this query to the current user's GitHub ID
		vars["login"] = githubv4.String(gitHubId)
		// and initialize a map to that will be used to hold the details for
		// all of the contributions made by this user
		userCommitContribs := []map[string]interface{}{}
		// and loop over the list of Org IDs
		for _, orgId := range orgIdList {
			// set the "organizationID" field and (re)set the "after" field its
			// initial value in the "vars" map
			vars["organizationID"] = orgId
			// define the variable used to track the cursor values as we go
			lastCursor := githubv4.String("")
			// then make requests for the contributions made by this user to this
			// organization (and continue doing so until we reach the end of the
			// list of contributions made by this user to this organization in the
			// specified time period)
			for {
				// set the "after" field to our current "lastCursor" value
				vars["after"] = lastCursor
				// run our query, returning the results in the CommitContributionsMadeQuery struct
				err := client.Query(context.Background(), &contributionsMadeQuery, vars)
				if err != nil {
					// Handle error.
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				// grab out the list of edges from the pull request contributions
				// made and loop over them
				contribsByRepository := contributionsMadeQuery.User.ContributionsCollection.CommitContributionsByRepository
				if len(contribsByRepository) == 0 {
					break
				}
				// define a flag we can use to break out of the loop when we reach the end of the list of contributions
				endOfContributions := false
				for _, contribByRepository := range contribsByRepository {
					edges := contribByRepository.Contributions.Edges
					if len(edges) == 0 {
						endOfContributions = true
						break
					}
					for _, edge := range edges {
						// add the details for this edge to the list of commit contributions
						// made by to the appropriate repository
						if _, ok := contribsByRepo[edge.Node.Repository.Url]; !ok {
							// if here, then we haven't seen this repository yet so create a new entry for it
							contribsByRepo[edge.Node.Repository.Url] = map[string]interface{}{
								"repositoryName":     edge.Node.Repository.Name,
								"totalContributions": edge.Node.CommitCount,
							}
						} else {
							// else just increment the number of contributions made to this repository
							repoContribsMap := contribsByRepo[edge.Node.Repository.Url].(map[string]interface{})
							if currentCount, ok := repoContribsMap["totalContributions"].(int); ok {
								repoContribsMap["totalContributions"] = currentCount + edge.Node.CommitCount
							}
						}
						// and add the details for this edge to the list of commit contributions
						// made by this user (these edges are organized by date/repository pairs)
						userCommitContribs = append(userCommitContribs, map[string]interface{}{
							"repositoryName":   edge.Node.Repository.Name,
							"numContributions": edge.Node.CommitCount,
							"contributedAt":    edge.Node.OccurredAt,
						})
						// and save the cursor value for this edge for use later on
						lastCursor = edge.Cursor
					}
				}
				// if we've reached the end of the list of contributions, break out of the loop
				if endOfContributions {
					break
				}
			}
		}
		// add pull requests for this user to the complete list of pull
		// requests by user
		if _, ok := contribsByUser["ByUser"]; ok {
			if val, ok := contribsByUser["ByUser"].([]map[string]interface{}); ok {
				contribsByUser["ByUser"] = append(val, map[string]interface{}{
					gitHubId: userCommitContribs,
				})
			}
		} else {
			contribsByUser["ByUser"] = append([]map[string]interface{}{}, map[string]interface{}{
				gitHubId: userCommitContribs,
			})
		}
	}
	// finally add an "AllUsers" entry to the list of contributions made by all users to each repository
	contribsByUser["AllUsers"] = contribsByRepo

	// and return the resulting map
	return contribsByUser
}
