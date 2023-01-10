/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	repoCmd = &cobra.Command{
		Use:   "repo",
		Short: "Gather repository-related data",
		Long: `The subcommand used as the root for all commands that make
repository-related queries`,
	}
)

func init() {
	rootCmd.AddCommand(repoCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
