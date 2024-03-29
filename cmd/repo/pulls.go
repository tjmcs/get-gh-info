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
	PullsCmd = &cobra.Command{
		Use:   "pulls",
		Short: "Gather PR-related data",
		Long:  "The subcommand used as the root for all queries for PR-related data",
	}
)

/*
 * Define a few types that we can use to define (ane extract data from) the body of the GraphQL
 * query that will be used to retrieve the list of open PRs in the named GitHub organization(s)
 */
type PullRequest struct {
	cmd.IssueOrPrBase
}
type PrSearchEdges []struct {
	Cursor githubv4.String
	Node   struct {
		PullRequest `graphql:"... on PullRequest"`
	}
}
type prSearchBody struct {
	IssueCount githubv4.Int
	Edges      PrSearchEdges
	PageInfo   cmd.PageInfo
}

/*
 * define a pair of structs that can be used to query GitHub for a list of all of the
 * open PRs in a given organization (by name) that match a given query; the first is
 * used to query for the first page of results and the second is used to query for
 * subsequent pages of results
 */
var FirstPrSearchQuery struct {
	Search struct {
		prSearchBody
	} `graphql:"search(first: $first, query: $query, type: $type)"`
}

var PrSearchQuery struct {
	Search struct {
		prSearchBody
	} `graphql:"search(first: $first, after: $after, query: $query, type: $type)"`
}

/*
 * define a few functions to get the values we'll need from the underling PullRequest
 */
func (p *PullRequest) IsClosed() bool {
	return p.Closed
}
func (p *PullRequest) GetCreatedAt() githubv4.DateTime {
	return p.CreatedAt
}
func (p *PullRequest) GetClosedAt() githubv4.DateTime {
	return p.ClosedAt
}
func (p *PullRequest) GetComments() cmd.Comments {
	return p.Comments
}

func init() {
	cmd.RepoCmd.AddCommand(PullsCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	PullsCmd.PersistentFlags().StringVarP(&cmd.LookbackTime, "lookback-time", "l", "", "'lookback' time window (eg. 10d, 3w, 2m, 1q, 1y)")
	PullsCmd.PersistentFlags().StringVarP(&cmd.ReferenceDate, "ref-date", "d", "", "reference date for time window (YYYY-MM-DD)")
	PullsCmd.PersistentFlags().BoolVarP(&cmd.CompleteWeeks, "complete-weeks", "w", false, "only output complete weeks (starting Monday)")
	PullsCmd.PersistentFlags().StringVarP(&cmd.CompTeam, "team", "t", "", "name of team to restrict repository list to")
	PullsCmd.PersistentFlags().StringVarP(&cmd.RepoMappingFile, "repo-mapping-file", "m", "", "name of the repository mapping file to use")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("lookbackTime", PullsCmd.PersistentFlags().Lookup("lookback-time"))
	viper.BindPFlag("referenceDate", PullsCmd.PersistentFlags().Lookup("ref-date"))
	viper.BindPFlag("completeWeeks", PullsCmd.PersistentFlags().Lookup("complete-weeks"))
	viper.BindPFlag("teamName", PullsCmd.PersistentFlags().Lookup("team"))
	viper.BindPFlag("repoMappingFile", PullsCmd.PersistentFlags().Lookup("repo-mapping-file"))
}
