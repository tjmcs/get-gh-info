/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjmcs/get-gh-info/utils"
)

// repoListCmd represents the 'repoList' command
var (
	searchPattern string
	matchCmd      = &cobra.Command{
		Use:   "match",
		Short: "Generates a list of repositories that match the search criteria",
		Long: `Constructs a list of all of the repositories who's name matches the search
criteria passed in by the user from the named set of GitHub organizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(repoList())
		},
	}
)

func init() {
	repoCmd.AddCommand(matchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	matchCmd.PersistentFlags().StringVarP(&searchPattern, "search-pattern", "p", "", "pattern to match against repository names")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("searchPattern", matchCmd.PersistentFlags().Lookup("searchPattern"))
}

/*
 * define a struct that can be used to put together a list of all of the
 * repositories in a given organization (by name) that match a given query
 */
var repositorySearchQuery struct {
	Search struct {
		RepositoryCount int
		Edges           []struct {
			Cursor githubv4.String
			Node   struct {
				Repository struct {
					Name       string
					IsArchived bool
					IsPrivate  bool
					Url        string
				} `graphql:"... on Repository"`
			}
		}
	} `graphql:"search(query: $query, type: $type, first: $first)"`
}

/*
 * define the function that is used to fetch a list of the Orb repositories
 * managed by the team under the named organizations
 */
func repoList() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our query for PR review contributions
	vars := map[string]interface{}{
		"type":  githubv4.SearchTypeRepository,
		"first": githubv4.Int(100),
	}
	// and initialize a map to that will be used to hold the details for
	// all of the pull request reviews made by this user
	repositoryList := map[string]interface{}{}

	// loop over the input organization names
	for _, orgName := range utils.GetOrgNameList() {
		// construct our query string and add it ot the vars map
		if searchPattern == "" {
			vars["query"] = githubv4.String(fmt.Sprintf("org:%s", orgName))
		} else {
			vars["query"] = githubv4.String(fmt.Sprintf("%s org:%s", searchPattern, orgName))
		}
		// run our query and add the data we want from the query results to the
		// repositoryList map
		err := client.Query(context.Background(), &repositorySearchQuery, vars)
		if err != nil {
			// Handle error.
			fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
			os.Exit(1)
		}
		// grab out the list of edges from the pull request review
		// contributions made and loop over them
		edges := repositorySearchQuery.Search.Edges
		fmt.Fprintf(os.Stderr, "Found %d repositories matching the pattern '%s' in '%s' orgization\n", len(edges), searchPattern, orgName)
		for _, edge := range edges {
			// if here, then we haven't seen this repository yet so create a new entry for it
			repositoryList[edge.Node.Repository.Name] = map[string]interface{}{
				"private":  edge.Node.Repository.IsPrivate,
				"archived": edge.Node.Repository.IsArchived,
				"url":      edge.Node.Repository.Url,
			}
		}
	}
	return repositoryList
}
