/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package user

import (
	"context"
	"fmt"
	"math"
	"os"

	mapset "github.com/deckarep/golang-set"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/tjmcs/get-gh-info/cmd"
	"github.com/tjmcs/get-gh-info/utils"
)

// contribSummaryCmd represents the 'contribSummary' command
var (
	contribSummaryCmd = &cobra.Command{
		Use:   "contribSummary",
		Short: "Generates a summary (including statistics) of contributions",
		Long: `Constructs a summary (including statistics) of all of the contributions
that each of the input users made to any repository to any of the repositories
in the named set of GitHub organizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(summaryOfContribs())
		},
	}
)

func init() {
	cmd.UserCmd.AddCommand(contribSummaryCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)

}

/*
 * define the function that is used to gather GitHub summary information
 * for the contrributions made by the named user(s) to the named org(s)
 */
func summaryOfContribs() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// and then get the list of organization IDs that we want to query
	orgIdList := utils.GetOrgIdList(client)
	// define the start and end time of our query window
	startDateTime, endDateTime := utils.GetQueryTimeWindow()
	// define a few lists that we'll use later on (to loop over the team members and
	// to skip some members when calculating statistics)
	userIdList := utils.GetUserIdList()
	_, teamList := utils.GetTeamMembers()
	// construct the list of GitHub IDs to gather information for
	mySet := mapset.NewSet()
	for _, user := range userIdList {
		mySet.Add(user)
	}
	// gitHubIdList := userIdList
	for _, member := range teamList {
		// gitHubIdList = append(gitHubIdList, member["githubid"])
		mySet.Add(member["githubid"])
	}
	// initialize the vars map that we'll use when making our query for a summary of contributions
	vars := map[string]interface{}{
		"from": startDateTime,
		"to":   endDateTime,
	}
	// and grab the GitHub IDs from that set as a slice
	gitHubIdList := mySet.ToSlice()
	// initialize a few variables
	var avgPullReqContribs, avgReposWithContribPullReqs,
		avgPullReqReviewContribs, avgReposWithContribPullReqReviews float64
	contribByUserSummary := map[string]interface{}{}
	// loop over the list of GitHub IDs
	for _, gitHubId := range gitHubIdList {
		// convert the input value to a string
		gitHubIdStr := gitHubId.(string)
		// initialize a few variables
		var totalPullReqContribs, totalReposWithContribPullReqs,
			totalPullReqReviewContribs, totalReposWithContribPullReqReviews int
		// set the login value for this query to the current user's GitHub ID
		vars["login"] = githubv4.String(gitHubIdStr)
		// loop over the list of organization IDs and gather contribution
		// information for this GitHub user for all of them
		for _, orgId := range orgIdList {
			// set the organization ID value for this query to the current
			// orgId value
			vars["organizationID"] = orgId
			// and run our query, returning the results in the ContribQuery struct
			err := client.Query(context.Background(), &ContribQuery, vars)
			if err != nil {
				// Handle error.
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			// extract the ContributionsCollection part of the result
			contributionsCollection := ContribQuery.User.ContributionsCollection
			// and use it to accumulate the results for this user to the repositories
			// in this organization
			totalPullReqContribs += contributionsCollection.TotalPullRequestContributions
			totalReposWithContribPullReqs += contributionsCollection.TotalRepositoriesWithContributedPullRequests
			totalPullReqReviewContribs += contributionsCollection.TotalPullRequestReviewContributions
			totalReposWithContribPullReqReviews += contributionsCollection.TotalRepositoriesWithContributedPullRequestReviews
		}
		// and add the contribution details for this user to the summary
		// for the entire team
		if utils.SliceContains(userIdList, gitHubIdStr) {
			contribByUserSummary[gitHubIdStr] = map[string]interface{}{
				"pullReqContribs":                 totalPullReqContribs,
				"reposWithPullReqContribs":        totalReposWithContribPullReqs,
				"pullReqReviewContribs":           totalPullReqReviewContribs,
				"reposWithPullReqReviewsContribs": totalReposWithContribPullReqReviews,
			}
		}
		// add current user contributions (weighted by the number of input GitHub users)
		// to determine the average for each metric for the team
		avgPullReqContribs += float64(totalPullReqContribs) / float64(len(gitHubIdList))
		avgReposWithContribPullReqs += float64(totalReposWithContribPullReqs) / float64(len(gitHubIdList))
		avgPullReqReviewContribs += float64(totalPullReqReviewContribs) / float64(len(gitHubIdList))
		avgReposWithContribPullReqReviews += float64(totalReposWithContribPullReqReviews) / float64(len(gitHubIdList))
	}

	// and add some summary statistics to the output map
	for _, gitHubId := range userIdList {
		userMap := contribByUserSummary[gitHubId].(map[string]interface{})
		if avgPullReqContribs != 0 {
			userMap["teamPcntPullReqContribs"] = math.Round(((float64(userMap["pullReqContribs"].(int)))/avgPullReqContribs)*10000) / 100
		}
		if avgReposWithContribPullReqs != 0 {
			userMap["teamPcntReposWithContribPullReqs"] = math.Round(((float64(userMap["reposWithPullReqContribs"].(int)))/avgReposWithContribPullReqs)*10000) / 100
		}
		if avgPullReqReviewContribs != 0 {
			userMap["teamPcntPullReqReviewContribs"] = math.Round(((float64(userMap["pullReqReviewContribs"].(int)))/avgPullReqReviewContribs)*10000) / 100
		}
		if avgReposWithContribPullReqReviews != 0 {
			userMap["teamPcntReposWithContribPullReqReviews"] = math.Round(((float64(userMap["reposWithPullReqReviewsContribs"].(int)))/avgReposWithContribPullReqReviews)*10000) / 100
		}
	}

	// and return the resulting map
	return contribByUserSummary
}
