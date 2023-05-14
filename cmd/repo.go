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
	repoCmd  = &cobra.Command{
		Use:   "repo",
		Short: "Gather repository-related data",
		Long: `The subcommand used as the root for all commands that make
repository-related queries`,
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
	rootCmd.PersistentFlags().StringVarP(&compTeam, "team", "t", "", "name of team to compare contributions against")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("teamName", rootCmd.PersistentFlags().Lookup("team"))
}
