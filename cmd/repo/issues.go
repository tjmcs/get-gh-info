/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package repo

import (
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjmcs/get-gh-info/cmd"
)

// rootCmd represents the base command when called without any subcommands
var (
	// RestrictToTeam is used in several subcommands to restrict the comments included
	// as feedback to only those that are made by immediate team members
	RestrictToTeam bool
	IssuesCmd      = &cobra.Command{
		Use:   "issues",
		Short: "Gather issue-related data",
		Long:  "The subcommand used as the root for all queries for issue-related data",
	}
)

/*
 * Define a few types that we can use to define (ane extract data from) the body of the GraphQL
 * query that will be used to retrieve the list of open issues in the named GitHub organization(s)
 */
type Issue struct {
	cmd.IssueOrPrBase
}
type IssueSearchEdges []struct {
	Cursor githubv4.String
	Node   struct {
		Issue `graphql:"... on Issue"`
	}
}
type issueSearchBody struct {
	IssueCount githubv4.Int
	Edges      IssueSearchEdges
	PageInfo   cmd.PageInfo
}

/*
 * define a pair of structs that can be used to query GitHub for a list of all of the
 * open PRs in a given organization (by name) that match a given query; the first is
 * used to query for the first page of results and the second is used to query for
 * subsequent pages of results
 */
var FirstIssueSearchQuery struct {
	Search struct {
		issueSearchBody
	} `graphql:"search(first: $first, query: $query, type: $type)"`
}

var IssueSearchQuery struct {
	Search struct {
		issueSearchBody
	} `graphql:"search(first: $first, after: $after, query: $query, type: $type)"`
}

/*
 * define a few functions to get the values we'll need from the underling Issue
 */
func (i *Issue) IsClosed() bool {
	return i.Closed
}
func (i *Issue) GetCreatedAt() githubv4.DateTime {
	return i.CreatedAt
}
func (i *Issue) GetClosedAt() githubv4.DateTime {
	return i.ClosedAt
}
func (i *Issue) GetComments() cmd.Comments {
	return i.Comments
}

func init() {
	cmd.RepoCmd.AddCommand(IssuesCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	IssuesCmd.PersistentFlags().StringVarP(&cmd.LookbackTime, "lookback-time", "l", "", "'lookback' time window (eg. 10d, 3w, 2m, 1q, 1y)")
	IssuesCmd.PersistentFlags().StringVarP(&cmd.ReferenceDate, "ref-date", "d", "", "reference date for time window (YYYY-MM-DD)")
	IssuesCmd.PersistentFlags().BoolVarP(&cmd.CompleteWeeks, "complete-weeks", "w", false, "only output complete weeks (starting Monday)")
	IssuesCmd.PersistentFlags().StringVarP(&cmd.CompTeam, "team", "t", "", "name of team to restrict repository list to")
	IssuesCmd.PersistentFlags().StringVarP(&cmd.RepoMappingFile, "repo-mapping-file", "m", "", "name of the repository mapping file to use")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("lookbackTime", IssuesCmd.PersistentFlags().Lookup("lookback-time"))
	viper.BindPFlag("referenceDate", IssuesCmd.PersistentFlags().Lookup("ref-date"))
	viper.BindPFlag("completeWeeks", IssuesCmd.PersistentFlags().Lookup("complete-weeks"))
	viper.BindPFlag("teamName", IssuesCmd.PersistentFlags().Lookup("team"))
	viper.BindPFlag("repoMappingFile", IssuesCmd.PersistentFlags().Lookup("repo-mapping-file"))
}
