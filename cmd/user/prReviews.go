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

// prReviewsCmd represents the 'prReviews' command
var prReviewsCmd = &cobra.Command{
	Use:   "prReviews",
	Short: "Generates a list the pull request reviews made",
	Long: `Constructs a list of all of the pull request reviews that each of the input
users performed in any repository to any of the repositories in the named set
of GitHub organizations (including the title, status, url, and repository name)
for each pull request submitted by that user.`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.DumpMapAsJSON(prReviews())
	},
}

func init() {
	cmd.UserCmd.AddCommand(prReviewsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

/*
 * and a similar struct that can be used to put together a list of all of the
 * pull request reviews performed by a given user to any repository in a given
 * organization
 */
var pullRequestReviewsPerformedQuery struct {
	User struct {
		Login                   string
		ContributionsCollection struct {
			PullRequestReviewContributions struct {
				Edges []PullRequestEdges
			} `graphql:"pullRequestReviewContributions(first: $first, after: $after)"`
		} `graphql:"contributionsCollection(from: $from, to: $to, organizationID: $organizationID)"`
	} `graphql:"user(login: $login)"`
}

/*
 * define the function that is used to fetch GitHub pull request information
 * for the pull requests made by the named user(s) against repositories under
 * the named org(s)
 */
func prReviews() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// and then get the list of organization IDs that we want to query
	orgIdList := utils.GetOrgIdList(client)
	// define the start and end time of our query window
	startDateTime, endDateTime := utils.GetQueryTimeWindow()
	// define a list that we'll use later on (to loop over the team members and
	// to skip some members when calculating statistics)
	gitHubIdList := utils.GetUserIdList()
	// initialize the vars map that we'll use when making our query for PR review contributions
	vars := map[string]interface{}{
		"from":  startDateTime,
		"to":    endDateTime,
		"first": githubv4.Int(100),
	}
	pullRequestReviewsByUser := map[string]interface{}{}
	prReviewsByRepo := map[string]interface{}{}
	// loop over the list of GitHub IDs
	for _, gitHubId := range gitHubIdList {
		// set the login value for this query to the current user's GitHub ID
		vars["login"] = githubv4.String(gitHubId)
		// and initialize a map to that will be used to hold the details for
		// all of the pull request reviews made by this user
		userPullRequestReviews := []map[string]interface{}{}
		// and loop over the list of Org IDs
		for _, orgId := range orgIdList {
			// set the "organizationID" field and (re)set the "after" field its
			// initial value in the "vars" map
			vars["organizationID"] = orgId
			// define the variable used to track the cursor values as we go
			lastCursor := githubv4.String("")
			// then make requests for the pull request reviews made by this user
			// to this organization (and continue doing so until we reach the end
			// of the list of pull request reviews made by this user to this
			// organization in the specified time period)
			for {
				// set the "after" field to our current "lastCursof" value
				vars["after"] = lastCursor
				// run our query, returning the results in the PullRequestReviewsPerformedQuery struct
				err := client.Query(context.Background(), &pullRequestReviewsPerformedQuery, vars)
				if err != nil {
					// Handle error.
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				// grab out the list of edges from the pull request review
				// contributions made and loop over them
				edges := pullRequestReviewsPerformedQuery.User.ContributionsCollection.PullRequestReviewContributions.Edges
				// if nothing was returned, then we've found all of the contributions
				// from this user to this organization so break out of the loop
				if len(edges) == 0 {
					break
				}
				fmt.Fprintf(os.Stderr, "Found %d pull request review contributions for user %s to org %s\n", len(edges), gitHubId, orgId)
				for _, edge := range edges {
					// save some typing later by grabbing the pull request associated with this edge
					pullReq := edge.Node.PullRequest
					// if the pull rquest review is for a pull request that was closed as merged,
					// then add the details for this edge to the list of commit contributions
					// made by to the appropriate repository
					if pullReq.Closed && pullReq.Merged {
						if _, ok := prReviewsByRepo[pullReq.Repository.Url]; !ok {
							// if here, then we haven't seen this repository yet so create a new entry for it
							prReviewsByRepo[pullReq.Repository.Url] = map[string]interface{}{
								"repositoryName":     pullReq.Repository.Name,
								"totalContributions": 1,
							}
						} else {
							// else just increment the number of contributions made to this repository
							repoContribsMap := prReviewsByRepo[pullReq.Repository.Url].(map[string]interface{})
							if currentCount, ok := repoContribsMap["totalContributions"].(int); ok {
								repoContribsMap["totalContributions"] = currentCount + 1
							}
						}
					}
					// add the details for this edge to the list of pull request
					// reviews made by this user
					userPullRequestReviews = append(userPullRequestReviews, map[string]interface{}{
						"author":         pullReq.Author.Login,
						"closed":         pullReq.Closed,
						"merged":         pullReq.Merged,
						"repositoryName": pullReq.Repository.Name,
						"title":          pullReq.Title,
						"url":            pullReq.Url,
					})
					// and save the cursor value for this edge for use later on
					lastCursor = edge.Cursor
				}
			}
		}
		// add pull request reviews for this user to the complete list of pull
		// requests by user
		if _, ok := pullRequestReviewsByUser["ByUser"]; ok {
			if val, ok := pullRequestReviewsByUser["ByUser"].([]map[string]interface{}); ok {
				pullRequestReviewsByUser["ByUser"] = append(val, map[string]interface{}{
					gitHubId: userPullRequestReviews,
				})
			}
		} else {
			pullRequestReviewsByUser["ByUser"] = append([]map[string]interface{}{}, map[string]interface{}{
				gitHubId: userPullRequestReviews,
			})
		}
	}
	// finally add an "AllUsers" entry to the list of contributions made by all users to each repository
	pullRequestReviewsByUser["AllUsers"] = prReviewsByRepo
	return pullRequestReviewsByUser
}
