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
	userList     string
	gitHubIdList string

	UserCmd = &cobra.Command{
		Use:   "user",
		Short: "Gather user-related data",
		Long:  "The subcommand used as the root for all queries for user-related data",
	}
)

func init() {
	RootCmd.AddCommand(UserCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	UserCmd.PersistentFlags().StringVarP(&userList, "user-list", "u", "", "list of users to gather contributions for")
	UserCmd.PersistentFlags().StringVarP(&gitHubIdList, "github-id-list", "i", "", "list of GitHub IDs to gather contributions for")
	UserCmd.PersistentFlags().StringVarP(&ReferenceDate, "ref-date", "d", "", "reference date for time window (YYYY-MM-DD)")
	UserCmd.PersistentFlags().StringVarP(&LookbackTime, "lookback-time", "l", "", "'lookback' time window (eg. 10d, 3w, 2m, 1q, 1y)")
	UserCmd.PersistentFlags().BoolVarP(&CompleteWeeks, "complete-weeks", "w", false, "only output complete weeks (starting Monday)")
	UserCmd.PersistentFlags().StringVarP(&CompTeam, "team", "t", "", "name of team to gather data for or compare against")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("userList", UserCmd.PersistentFlags().Lookup("user-list"))
	viper.BindPFlag("gitHubIdList", UserCmd.PersistentFlags().Lookup("github-id-list"))
	viper.BindPFlag("referenceDate", UserCmd.PersistentFlags().Lookup("ref-date"))
	viper.BindPFlag("lookbackTime", UserCmd.PersistentFlags().Lookup("lookback-time"))
	viper.BindPFlag("completeWeeks", UserCmd.PersistentFlags().Lookup("complete-weeks"))
	viper.BindPFlag("teamName", UserCmd.PersistentFlags().Lookup("team"))

}
