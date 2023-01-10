/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/tjmcs/get-gh-info/utils"
)

// prlistCmd represents the 'prlist' command
var prlistCmd = &cobra.Command{
	Use:   "prList",
	Short: "Generates a list the pull requests made",
	Long: `Constructs a list of all of the pull requests that each of the input users
made to any repository to any of the repositories in the named set of GitHub
organizations (including the title, status, url, and repository name) for each
pull request submitted by that user.`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.DumpMapAsJSON(prList())
	},
}

func init() {
	userCmd.AddCommand(prlistCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

/*
 * then a struct that can be used put together a list of all of the commits
 * included in a pull request
 */
type CommitEdges struct {
	Node struct {
		Commit struct {
			Author struct {
				User struct {
					Login string
				}
			}
			CommittedDate githubv4.DateTime
			Message       string
			PushedDate    githubv4.DateTime
			Repository    struct {
				Name string
				Url  string
			}
			Url string
		}
	}
}

/*
 * and a similar struct that can be used to put together a list of all of the
 * pull requests performed by a given user to any repository in a given
 * organization
 */
type PullRequestEdges struct {
	Cursor githubv4.String
	Node   struct {
		PullRequest struct {
			Author struct {
				Login string
			}
			Closed   bool
			ClosedAt githubv4.DateTime
			Commits  struct {
				Edges []CommitEdges
			} `graphql:"commits(first: 1)"`
			CreatedAt  githubv4.DateTime
			Merged     bool
			MergedAt   githubv4.DateTime
			Repository struct {
				Name string
				Url  string
			}
			Title string
			Url   string
		}
	}
}

/*
 * and a struct that can be used to put together a list of all of the pull
 * requests made by a given user to any repository in a given organization
 */
var PullRequestsMadeQuery struct {
	User struct {
		Login                   string
		ContributionsCollection struct {
			PullRequestContributions struct {
				Edges []PullRequestEdges
			} `graphql:"pullRequestContributions(first: $first, after: $after)"`
		} `graphql:"contributionsCollection(from: $from, to: $to, organizationID: $organizationID)"`
	} `graphql:"user(login: $login)"`
}

/*
 * define the function that is used to fetch the GitHub pull request information
 * for the pull requests made by the named user(s) against repositories under
 * the named org(s)
 */
func prList() map[string]interface{} {
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
	pullRequestsByUser := map[string]interface{}{}
	prsByRepo := map[string]interface{}{}
	// loop over the list of GitHub IDs
	for _, gitHubId := range gitHubIdList {
		// set the login value for this query to the current user's GitHub ID
		vars["login"] = githubv4.String(gitHubId)
		// and initialize a map to that will be used to hold the details for
		// all of the pull requests made by this user
		userPullRequests := []map[string]interface{}{}
		// and loop over the list of Org IDs
		for _, orgId := range orgIdList {
			// set the "organizationID" field and (re)set the "after" field its
			// initial value in the "vars" map
			vars["organizationID"] = orgId
			// define the variable used to track the cursor values as we go
			lastCursor := githubv4.String("")
			// then make requests for the pull requests made by this user to this
			// organization (and continue doing so until we reach the end of the
			// list of pull requests made by this user to this organization in the
			// specified time period)
			for {
				// set the "after" field to our current "lastCursof" value
				vars["after"] = lastCursor
				// run our query, returning the results in the PullRequestsMadeQuery struct
				err := client.Query(context.Background(), &PullRequestsMadeQuery, vars)
				if err != nil {
					// Handle error.
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				// grab out the list of edges from the pull request contributions
				// made and loop over them
				edges := PullRequestsMadeQuery.User.ContributionsCollection.PullRequestContributions.Edges
				// if nothing was returned, then we've found all of the contributions
				// from this user to this organization so break out of the loop
				if len(edges) == 0 {
					break
				}
				fmt.Fprintf(os.Stderr, "Found %d pull request contributions for user %s to org %s\n", len(edges), gitHubId, orgId)
				for _, edge := range edges {
					// save some typing later by grabbing the pull request associated with this edge
					pullReq := edge.Node.PullRequest
					// if the pull rquest was closed as merged, then add the details for this
					// edge to the list of commit contributions made by to the appropriate
					// repository
					if pullReq.Closed && pullReq.Merged {
						if _, ok := prsByRepo[pullReq.Repository.Url]; !ok {
							// if here, then we haven't seen this repository yet so create a new entry for it
							prsByRepo[pullReq.Repository.Url] = map[string]interface{}{
								"repositoryName":     pullReq.Repository.Name,
								"totalContributions": 1,
							}
						} else {
							// else just increment the number of contributions made to this repository
							repoContribsMap := prsByRepo[pullReq.Repository.Url].(map[string]interface{})
							if currentCount, ok := repoContribsMap["totalContributions"].(int); ok {
								repoContribsMap["totalContributions"] = currentCount + 1
							}
						}
					}
					// determine how long the pull request was open (or has been open if it's still open)
					// after it was created along with the time since the first commit was made
					daysOpen := 0.0
					daysSinceFirstCommit := 0.0
					firstCommitAt := pullReq.Commits.Edges[0].Node.Commit.CommittedDate.Time
					if pullReq.Closed && !pullReq.Merged {
						// pull request was closed but not merged
						daysOpen = math.Round(pullReq.ClosedAt.Sub(pullReq.CreatedAt.Time).Hours()/24.0*10000) / 10000
						daysSinceFirstCommit = math.Round(pullReq.ClosedAt.Sub(firstCommitAt).Hours()/24.0*10000) / 10000
					} else if pullReq.Merged {
						// pull request was merged
						daysOpen = math.Round(pullReq.MergedAt.Sub(pullReq.CreatedAt.Time).Hours()/24.0*10000) / 10000
						daysSinceFirstCommit = math.Round(pullReq.MergedAt.Sub(firstCommitAt).Hours()/24.0*10000) / 10000
					} else {
						// pull request is still open today (so used time elapsed since it was created)
						daysOpen = math.Round(time.Since(pullReq.CreatedAt.Time).Hours()/24.0*10000) / 10000
						daysSinceFirstCommit = math.Round(time.Since(firstCommitAt).Hours()/24.0*10000) / 10000
					}
					// add the details for this edge to the list of pull requests
					// made by this user
					userPullRequests = append(userPullRequests, map[string]interface{}{
						"author":         pullReq.Author.Login,
						"closed":         pullReq.Closed,
						"closedAt":       pullReq.ClosedAt,
						"createdAt":      pullReq.CreatedAt,
						"daysOpen":       daysOpen,
						"daysWorked":     math.Max(daysOpen, daysSinceFirstCommit),
						"firstCommitAt":  firstCommitAt,
						"merged":         pullReq.Merged,
						"mergedAt":       pullReq.MergedAt,
						"repositoryName": pullReq.Repository.Name,
						"title":          pullReq.Title,
						"url":            pullReq.Url,
					})
					// and save the cursor value for this edge for use later on
					lastCursor = edge.Cursor
				}
			}
		}
		// add pull requests for this user to the complete list of pull
		// requests by user
		if _, ok := pullRequestsByUser["ByUser"]; ok {
			if val, ok := pullRequestsByUser["ByUser"].([]map[string]interface{}); ok {
				pullRequestsByUser["ByUser"] = append(val, map[string]interface{}{
					gitHubId: userPullRequests,
				})
			}
		} else {
			pullRequestsByUser["ByUser"] = append([]map[string]interface{}{}, map[string]interface{}{
				gitHubId: userPullRequests,
			})
		}
	}
	// finally add an "AllUsers" entry to the list of contributions made by all users to each repository
	pullRequestsByUser["AllUsers"] = prsByRepo
	return pullRequestsByUser
}
