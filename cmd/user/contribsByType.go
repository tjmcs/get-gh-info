/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package user

import (
	"github.com/spf13/cobra"
	"github.com/tjmcs/get-gh-info/cmd"
	"github.com/tjmcs/get-gh-info/utils"
)

// contribsByTypeCmd represents the 'contribsByType' command
var (
	contribsByTypeCmd = &cobra.Command{
		Use:   "contribsByType",
		Short: "Generates a list of PRs and PR reviews made",
		Long: `Constructs a list of PRs and PR reviews made by each of the input users
against any of the repositories in the named set of GitHub organizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			contribsByType()
		},
	}
)

func init() {
	cmd.UserCmd.AddCommand(contribsByTypeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

/*
 * define the function that is used to print (as a JSON string) the information
 * for all of the pull request contributions (both pull requests, and pull request reviews)
 * made by the named user(s) against repositories under the named org(s)
 */
func contribsByType() {
	// initialize the map used to track the contributions (grouped by type of contribution)
	contribsByUser := map[string]interface{}{}
	// first, fetch the list of PRs made by the named user(s) against repositories
	// under the named org(s)
	contribsByUser["pullRequests"] = prList()
	// then append onto that the list of PR reviews made by the named user(s) against
	// repositories under the named org(s)
	contribsByUser["pullRequestReviews"] = prReviews()
	// and dump out the results
	utils.DumpMapAsJSON(contribsByUser)
}
