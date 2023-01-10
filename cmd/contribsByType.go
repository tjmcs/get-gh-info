/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// contribsByTypeCmd represents the 'contribsByType' command
var (
	contribsByTypeCmd = &cobra.Command{
		Use:   "contribsByType",
		Short: "Generates a list of any contributions made (by user)",
		Long: `Constructs a list (by user) of any contributions made
(commits and pull requests) by each of the input users against any
of the repositories in the named set of GitHub organizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			getContribsByType()
		},
	}
)

func init() {
	rootCmd.AddCommand(contribsByTypeCmd)

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
func getContribsByType() {
	// initialize the map used to track the contributions (grouped by type of contribution)
	contribsByUser := map[string]interface{}{}
	// first, fetch the list of PRs made by the named user(s) against repositories
	// under the named org(s)
	contribsByUser["pullRequests"] = fetchPrList()
	// then append onto that the list of PR reviews made by the named user(s) against
	// repositories under the named org(s)
	contribsByUser["pullRequestReviews"] = fetchPrReviews()
	// and dump out the results
	jsonStr, err := json.MarshalIndent(contribsByUser, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(-2)
	}
	fmt.Fprintf(os.Stdout, "%s\n", string(jsonStr))
}
