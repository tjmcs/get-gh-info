/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// define a couple of constant variables containing formating strings
	// (used to format dates for output and GraphQL queries)
	ISO8601_FormatStr     = "2006-01-02T15:04:05.999Z"
	YearMonthDayFormatStr = "2006-01-02"
)

// rootCmd represents the base command when called without any subcommands
var (
	// used locally to exclude private repositories from the output
	excludePrivate bool
	// used in some of the issues/pulls subcommands to track a flag indicating
	// that the output should be sorted by the time to first response
	SortByTimeToFirstResponse bool
	// used in some of the issues/pulls subcommands to track a flag indicating
	// that the output should be sorted by the time to last response (or staleness)
	SortByStaleness bool
	// and the repo command itself
	RepoCmd = &cobra.Command{
		Use:   "repo",
		Short: "Gather repository-related data",
		Long:  "The subcommand used as the root for all queries for repository-related data",
	}
)

/*
 * Define a few types that we can use to define (and extract data from) the body of the GraphQL
 * queries that will be used to retrieve the list of issues/PRs in the named GitHub organization(s)
 */
type Author struct {
	Login string
	User  struct {
		Email   string
		Company string
	} `graphql:"... on User"`
}

type Repository struct {
	Name       string
	Url        string
	IsPrivate  bool
	IsArchived bool
}

type Assignees struct {
	Edges []struct {
		Node struct {
			Login string
		}
	}
}

type Comments struct {
	Nodes []struct {
		CreatedAt githubv4.DateTime
		UpdatedAt githubv4.DateTime
		Author    struct {
			Login string
		}
		AuthorAssociation string
		Body              string
	}
}

type IssueOrPrBase struct {
	CreatedAt         githubv4.DateTime
	UpdatedAt         githubv4.DateTime
	Closed            bool
	ClosedAt          githubv4.DateTime
	Title             string
	Url               string
	Author            Author
	AuthorAssociation string
	Repository        Repository
	Assignees         Assignees `graphql:"assignees(first: 10)"`
	Comments          Comments  `graphql:"comments(first: 100, orderBy: $orderCommentsBy)"`
}

type PageInfo struct {
	EndCursor   githubv4.String
	HasNextPage bool
}

func init() {
	RootCmd.AddCommand(RepoCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	RepoCmd.PersistentFlags().BoolVarP(&excludePrivate, "exclude-private-repos", "e", false, "exclude private repositories from output")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("excludePrivateRepos", RepoCmd.PersistentFlags().Lookup("exclude-private-repos"))
}
