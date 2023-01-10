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
	endDate      string
	monthsBack   int

	userCmd = &cobra.Command{
		Use:   "user",
		Short: "Gather user-related data",
		Long: `The subcommand used as the root for all commands that make
user-related queries`,
	}
)

func init() {
	rootCmd.AddCommand(userCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	userCmd.PersistentFlags().StringVarP(&userList, "user-list", "u", "", "list of users to gather contributions for")
	userCmd.PersistentFlags().StringVarP(&gitHubIdList, "github-id-list", "i", "", "list of GitHub IDs to gather contributions for")
	userCmd.PersistentFlags().IntVarP(&monthsBack, "months-back", "m", 6, "length of time to look back in months")
	userCmd.PersistentFlags().StringVarP(&endDate, "end-date", "d", "", "date to start looking back from (in YYYY-MM-DD format)")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("userList", userCmd.PersistentFlags().Lookup("user-list"))
	viper.BindPFlag("gitHubIdList", userCmd.PersistentFlags().Lookup("github-id-list"))
	viper.BindPFlag("monthsBack", userCmd.PersistentFlags().Lookup("months-back"))
	viper.BindPFlag("endDate", userCmd.PersistentFlags().Lookup("end-date"))
}
