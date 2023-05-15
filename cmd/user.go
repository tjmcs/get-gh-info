/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var (
	userList      string
	gitHubIdList  string
	referenceDate string
	duration      string
	completeWeeks bool

	userCmd = &cobra.Command{
		Use:   "user",
		Short: "Gather user-related data",
		Long:  "The subcommand used as the root for all queries for user-related data",
	}
)

func init() {
	rootCmd.AddCommand(userCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	userCmd.PersistentFlags().StringVarP(&userList, "user-list", "u", "", "list of users to gather contributions for")
	userCmd.PersistentFlags().StringVarP(&gitHubIdList, "github-id-list", "i", "", "list of GitHub IDs to gather contributions for")
	userCmd.PersistentFlags().StringVarP(&duration, "lookback-time", "l", "", "'lookback' time window (eg. 10d, 3w, 2m, 1q, 1y)")
	userCmd.PersistentFlags().StringVarP(&referenceDate, "ref-date", "d", "", "reference date for time window (YYYY-MM-DD)")
	userCmd.PersistentFlags().BoolVarP(&completeWeeks, "complete-weeks", "w", false, "only output complete weeks (starting Monday)")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("userList", userCmd.PersistentFlags().Lookup("user-list"))
	viper.BindPFlag("gitHubIdList", userCmd.PersistentFlags().Lookup("github-id-list"))
	viper.BindPFlag("lookbackTime", userCmd.PersistentFlags().Lookup("lookback-time"))
	viper.BindPFlag("referenceDate", userCmd.PersistentFlags().Lookup("ref-date"))
	viper.BindPFlag("completeWeeks", userCmd.PersistentFlags().Lookup("complete-weeks"))
}
