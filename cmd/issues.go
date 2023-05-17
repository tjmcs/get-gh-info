/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	issuesCmd = &cobra.Command{
		Use:   "issues",
		Short: "Gather issue-related data",
		Long:  "The subcommand used as the root for all queries for issue-related data",
	}
)

func init() {
	repoCmd.AddCommand(issuesCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
}
