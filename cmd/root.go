/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var (
	// define variables used in subcommands to setup flags asociated with the time window
	// for our queries
	LookbackTime  string
	ReferenceDate string
	CompleteWeeks bool
	// and a couple of others that are used in various subcommands
	CompTeam       string
	ExcludePrivate bool
	// and some other global variables that are used locally to setup persistent flags
	cfgFile    string
	outputFile string
	orgList    string

	RootCmd = &cobra.Command{
		Use:   "getGhInfo",
		Short: "Gathers information from GitHub using the GitHub GraphQL API",
		Long: `Gathers the requested information from GitHub using the GitHub GraphQL API
(where the input parameters for the query to run are provided either on the
command-line or in an associated configuration file) and outputs the results`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(2)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "configuration file to use")
	RootCmd.PersistentFlags().StringVarP(&outputFile, "file", "f", "", "file/stream for output (defaults to stdout)")
	RootCmd.PersistentFlags().StringVarP(&orgList, "org-list", "o", "", "list of orgs to gather information from")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("outputFile", RootCmd.PersistentFlags().Lookup("file"))
	viper.BindPFlag("orgList", RootCmd.PersistentFlags().Lookup("org-list"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// if a configuration file was passed in, use it
		viper.SetConfigFile(cfgFile)
		viper.SetConfigType("yaml")
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		// Add a config with name "getGhInfo" (with or without extension)
		// in the '$HOME/.config' directory to the search path
		viper.SetConfigName("getGhInfo")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(fmt.Sprintf("%s/.config", home))
		// and add a config with the name "config" (with or without extension)
		// in the current working directory to the search path (note that files
		// added later take precedence over those added earlier)
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
	}

	// read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		os.Exit(3)
	}
}
