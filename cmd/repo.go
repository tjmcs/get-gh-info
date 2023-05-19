/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var (
	compTeam string
	// used to exclude private repositories from the output
	excludePrivate bool
	RepoCmd        = &cobra.Command{
		Use:   "repo",
		Short: "Gather repository-related data",
		Long:  "The subcommand used as the root for all queries for repository-related data",
	}
)

/*
 * PageInfo is a struct that contains the information needed to paginate through
 * a list of items returned from a GraphQL query.
 */
type PageInfo struct {
	EndCursor   githubv4.String
	HasNextPage bool
}

func init() {
	RootCmd.AddCommand(RepoCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	RepoCmd.PersistentFlags().StringVarP(&compTeam, "team", "t", "", "name of team to gather data for")
	RepoCmd.PersistentFlags().StringVarP(&referenceDate, "ref-date", "d", "", "reference date for time window (YYYY-MM-DD)")
	RepoCmd.PersistentFlags().BoolVarP(&completeWeeks, "complete-weeks", "w", false, "only output complete weeks (starting Monday)")
	RepoCmd.PersistentFlags().BoolVarP(&excludePrivate, "exclude-private-repos", "e", false, "exclude private repositories from output")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("teamName", RepoCmd.PersistentFlags().Lookup("team"))
	viper.BindPFlag("referenceDate", RepoCmd.PersistentFlags().Lookup("ref-date"))
	viper.BindPFlag("completeWeeks", RepoCmd.PersistentFlags().Lookup("complete-weeks"))
	viper.BindPFlag("excludePrivateRepos", RepoCmd.PersistentFlags().Lookup("exclude-private-repos"))
}
