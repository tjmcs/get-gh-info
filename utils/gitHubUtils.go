/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// define the struct that we'll use to determine the organization ID values
// that correspond to the input organization names
var OrgIdQuery struct {
	Organization struct {
		ID githubv4.ID
	} `graphql:"organization(login: $orgname)"`
}

/*
 * collect together the code used to setup any of the GraphQL queries that
 * we've defined here; first a pair function that we can use to get a new
 * (authenticated) GraphQL client
 */
func GetAuthenticatedClient() *githubv4.Client {
	// setup an authenticated HTTP client for use with the GitHub GraphQL API
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	// and return a new pointer to a GitHubv4 client that uses that
	// authenticated HTTP client
	return githubv4.NewClient(httpClient)
}

/*
 * used to get the list of organization names from the command-line or
 * from the configuration file (in that order)
 */
func GetOrgNameList() []string {
	// first get the list of organization names that we want to query
	// from either the command line or the configuration file (in that order)
	var orgNameList []string
	inputOrgList := viper.Get("orgList").(string)
	if inputOrgList != "" {
		orgNameList = strings.Split(inputOrgList, ",")
	} else {
		orgNameList = viper.GetStringSlice("orgs")
	}
	return orgNameList
}

/*
 * and a function that can be used to convert a list of organization
 * names to a list of organization ID values
 */
func GetOrgIdList(client *githubv4.Client) []githubv4.ID {
	// first get the list of organization names that we want to query
	// from either the command line or the configuration file (in that order)
	orgNameList := GetOrgNameList()
	orgIdList := []githubv4.ID{}
	// and then get the list of organization IDs that correspond to those
	// organization names
	for _, orgname := range orgNameList {
		vars := map[string]interface{}{
			"orgname": githubv4.String(orgname),
		}
		err := client.Query(context.Background(), &OrgIdQuery, vars)
		if err != nil {
			// Handle error.
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		orgIdList = append(orgIdList, OrgIdQuery.Organization.ID)
	}
	return orgIdList
}
