/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package repo

import (
	"github.com/spf13/cobra"
	"github.com/tjmcs/get-gh-info/cmd"
)

// rootCmd represents the base command when called without any subcommands
var (
	PullsCmd = &cobra.Command{
		Use:   "pulls",
		Short: "Gather PR-related data",
		Long:  "The subcommand used as the root for all queries for PR-related data",
	}
)

func init() {
	cmd.RepoCmd.AddCommand(PullsCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
}