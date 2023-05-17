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
	// restrictToTeam is used in several subcommands to restrict the comments included
	// as feedback to only those that are made by immediate team members
	restrictToTeam bool
	repoCmd        = &cobra.Command{
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
	rootCmd.AddCommand(repoCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	rootCmd.PersistentFlags().StringVarP(&compTeam, "team", "t", "", "name of team to gather data for")
	rootCmd.PersistentFlags().StringVarP(&referenceDate, "ref-date", "d", "", "reference date for time window (YYYY-MM-DD)")
	rootCmd.PersistentFlags().BoolVarP(&completeWeeks, "complete-weeks", "w", false, "only output complete weeks (starting Monday)")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("teamName", rootCmd.PersistentFlags().Lookup("team"))
	viper.BindPFlag("referenceDate", rootCmd.PersistentFlags().Lookup("ref-date"))
	viper.BindPFlag("completeWeeks", rootCmd.PersistentFlags().Lookup("complete-weeks"))
}
