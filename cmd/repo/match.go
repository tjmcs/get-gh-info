/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package repo

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/gobwas/glob"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjmcs/get-gh-info/cmd"
	"github.com/tjmcs/get-gh-info/utils"
)

// repoListCmd represents the 'repoList' command
var (
	includeArchivedRepos bool
	searchPattern        string
	globStyleMatch       bool
	matchCmd             = &cobra.Command{
		Use:   "match",
		Short: "Show list of repositories that match the search criteria",
		Long: `Constructs a list of all of the repositories in the named (set of) GitHub
organization(s) that have a name matching the define search pattern
passed in by the user.`,
		Run: func(cmd *cobra.Command, args []string) {
			utils.DumpMapAsJSON(repoList())
		},
	}
)

func init() {
	cmd.RepoCmd.AddCommand(matchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	matchCmd.Flags().StringVarP(&searchPattern, "search-pattern", "p", "", "pattern to match against repository names")
	matchCmd.Flags().BoolVarP(&includeArchivedRepos, "include-archived-repos", "i", false, "include archived repositories in output")
	matchCmd.Flags().BoolVarP(&globStyleMatch, "glob-style-pattern", "g", false, "interpret the pattern as a glob-style pattern")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("searchPattern", matchCmd.Flags().Lookup("searchPattern"))
	viper.BindPFlag("includeArchivedRepos", matchCmd.Flags().Lookup("include-archived-repos"))
	viper.BindPFlag("globStylePattern", matchCmd.Flags().Lookup("glob-style-pattern"))
}

/*
 * define a struct that can be used to put together a list of all of the
 * repositories in a given organization (by name) that match a given query
 */
type repositorySearchEdges []struct {
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
type repositorySearchBody struct {
	RepositoryCount int
	Edges           repositorySearchEdges
	PageInfo        cmd.PageInfo
}

var firstRepositorySearchQuery struct {
	Search struct {
		repositorySearchBody
	} `graphql:"search(query: $query, type: $type, first: $first)"`
}

var repositorySearchQuery struct {
	Search struct {
		repositorySearchBody
	} `graphql:"search(query: $query, type: $type, after: $after, first: $first)"`
}

/*
 * define the function that is used to fetch a list of the Orb repositories
 * managed by the team under the named organizations
 */
func repoList() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := utils.GetAuthenticatedClient()
	// initialize the vars map that we'll use when making our query for matching repositories
	vars := map[string]interface{}{
		"type":  githubv4.SearchTypeRepository,
		"first": githubv4.Int(100),
	}
	// and initialize a map to that will be used to hold the details for
	// all of the pull request reviews made by this user
	repositoryList := map[string]interface{}{}
	// and grab a couple of flag values from viper
	excludePrivateRepos := viper.GetBool("excludePrivateRepos")
	includeArchivedRepos := viper.GetBool("includeArchivedRepos")
	// loop over the input organization names
	for _, orgName := range utils.GetOrgNameList() {
		// construct our query string and add it ot the vars map
		vars["query"] = githubv4.String(fmt.Sprintf("org:%s", orgName))
		// iniitialize a few variables that we'll use for pattern matching
		// (depending on the flags set on the command-line)
		var searchRE *regexp.Regexp
		var err error
		var globPattern glob.Glob
		searchPatternType := ""
		// if the user specified a search pattern
		if searchPattern != "" {
			// and if they also specified that they wanted to use a glob-style pattern
			// to match repository names, then compile the pattern as a glob-style
			// pattern
			if viper.GetBool("globStylePattern") {
				globPattern, err = glob.Compile(searchPattern)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
					os.Exit(1)
				}
				searchPatternType = "glob"
			} else {
				// otherwise compile the searchPattern as a regular expression
				searchRE, err = regexp.Compile(searchPattern)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
					os.Exit(1)
				}
				searchPatternType = "regexp"
			}
		}
		// loop over the pages of results from this query until we've reached the end
		// of the list of PRs that matched
		for {
			// initialize a few variables that we'll use to parse the query results
			var edges repositorySearchEdges
			var pageInfo cmd.PageInfo
			// run our query and add the data we want from the query results to the
			// repositoryList map
			if vars["after"] == nil {
				err = client.Query(context.Background(), &firstRepositorySearchQuery, vars)
			} else {
				err = client.Query(context.Background(), &repositorySearchQuery, vars)
			}
			if err != nil {
				// Handle error.
				fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
				os.Exit(1)
			}
			// grab out the list of edges and the page info from the results of our search
			// and loop over the edges
			if vars["after"] == nil {
				edges = firstRepositorySearchQuery.Search.Edges
				pageInfo = firstRepositorySearchQuery.Search.PageInfo
				fmt.Fprintf(os.Stderr, ".")
			} else {
				edges = repositorySearchQuery.Search.Edges
				pageInfo = repositorySearchQuery.Search.PageInfo
				fmt.Fprintf(os.Stderr, ".")
			}
			// if we didn't get any edges back, then we've reached the end of the results
			// so break out of the loop
			if len(edges) == 0 {
				break
			}
			// fmt.Fprintf(os.Stderr, "Found %d repositories in '%s' orgization\n", len(edges), orgName)
			for _, edge := range edges {
				// if flag was not set to include archived repositories and the repository
				// is archived, then skip it
				if !includeArchivedRepos && edge.Node.Repository.IsArchived {
					continue
				}
				// if flag was set to exclude private repositories and the repository
				// is private, then skip it
				if excludePrivateRepos && edge.Node.Repository.IsPrivate {
					continue
				}
				// check to see if the repository name matches the search pattern (if one was specified)
				// based on the type of pattern matching we're doing (if no search pattern was specified,
				// then neither of these will be tried and the repository will be included in the list)
				if searchPatternType == "regexp" && !searchRE.MatchString(edge.Node.Repository.Name) {
					// if not, then skip this repository
					continue
				} else if searchPatternType == "glob" && !globPattern.Match(edge.Node.Repository.Name) {
					// if not, then skip this repository
					continue
				}
				// if here, then we haven't seen this repository yet so create a new entry for it
				repositoryList[edge.Node.Repository.Name] = map[string]interface{}{
					"private":  edge.Node.Repository.IsPrivate,
					"archived": edge.Node.Repository.IsArchived,
					"url":      edge.Node.Repository.Url,
				}
			}
			// if we've reached the end of the list of repositories, break out of the loop
			if !pageInfo.HasNextPage {
				break
			}
			// set the "after" field to the "EndCursor" from the pageInfo structure so
			// we will get the next page of results when we run the query again
			vars["after"] = pageInfo.EndCursor
		}
		// and unset the "after" key in the vars map so that we're ready
		// to query for restuls from the next organization
		delete(vars, "after")
	}
	fmt.Fprintf(os.Stderr, "\n")
	return repositoryList
}
