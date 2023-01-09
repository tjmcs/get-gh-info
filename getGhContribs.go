package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

/*
 * define a struct that can be used to put together a list of all of the
 * repositories in a given organization (by name) that match a given query
 */
var RepositorySearchQuery struct {
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

// define the struct that we'll use to determine the organization ID values
// that correspond to the input organization names
var OrgIdQuery struct {
	Organization struct {
		ID githubv4.ID
	} `graphql:"organization(login: $orgname)"`
}

// define the struct that we'll use to determine the total contributions from
// each of the input usernames to each of the input organizations
var ContribQuery struct {
	User struct {
		Login                   string
		ContributionsCollection struct {
			TotalIssueContributions                            int
			TotalRepositoriesWithContributedIssues             int
			TotalCommitContributions                           int
			TotalRepositoriesWithContributedCommits            int
			TotalPullRequestContributions                      int
			TotalRepositoriesWithContributedPullRequests       int
			TotalPullRequestReviewContributions                int
			TotalRepositoriesWithContributedPullRequestReviews int
		} `graphql:"contributionsCollection(from: $from, to: $to, organizationID: $organizationID)"`
	} `graphql:"user(login: $login)"`
}

/*
 * define a few types that we'll be using in some of our query structs, first
 * a struct for the contributions made to each repository by a given user
 */
type ContributionEdges struct {
	Cursor githubv4.String
	Node   struct {
		Repository struct {
			Name string
			Url  string
		}
		CommitCount int
		OccurredAt  githubv4.DateTime
	}
}

/*
 * and a struct that can be used to put together a list of all of the
 * contributions made by a given user to any repository in a given organization
 */
var ContributionsMadeQuery struct {
	User struct {
		Login                   string
		ContributionsCollection struct {
			CommitContributionsByRepository []struct {
				Contributions struct {
					Edges []ContributionEdges
				} `graphql:"contributions(first: $first, after: $after)"`
			} `graphql:"commitContributionsByRepository(maxRepositories: 100)"`
		} `graphql:"contributionsCollection(from: $from, to: $to, organizationID: $organizationID)"`
	} `graphql:"user(login: $login)"`
}

/*
 * then a struct that can be used put together a list of all of the commits
 * included in a pull request
 */
type CommitEdges struct {
	Node struct {
		Commit struct {
			Author struct {
				User struct {
					Login string
				}
			}
			CommittedDate githubv4.DateTime
			Message       string
			PushedDate    githubv4.DateTime
			Repository    struct {
				Name string
				Url  string
			}
			Url string
		}
	}
}

/*
 * and a similar struct that can be used to put together a list of all of the
 * pull requests performed by a given user to any repository in a given
 * organization
 */
type PullRequestEdges struct {
	Cursor githubv4.String
	Node   struct {
		PullRequest struct {
			Author struct {
				Login string
			}
			Closed   bool
			ClosedAt githubv4.DateTime
			Commits  struct {
				Edges []CommitEdges
			} `graphql:"commits(first: 1)"`
			CreatedAt  githubv4.DateTime
			Merged     bool
			MergedAt   githubv4.DateTime
			Repository struct {
				Name string
				Url  string
			}
			Title string
			Url   string
		}
	}
}

/*
 * and a struct that can be used to put together a list of all of the pull
 * requests made by a given user to any repository in a given organization
 */
var PullRequestsMadeQuery struct {
	User struct {
		Login                   string
		ContributionsCollection struct {
			PullRequestContributions struct {
				Edges []PullRequestEdges
			} `graphql:"pullRequestContributions(first: $first, after: $after)"`
		} `graphql:"contributionsCollection(from: $from, to: $to, organizationID: $organizationID)"`
	} `graphql:"user(login: $login)"`
}

/*
 * and a similar struct that can be used to put together a list of all of the
 * pull request reviews performed by a given user to any repository in a given
 * organization
 */
var PullRequestReviewsPerformedQuery struct {
	User struct {
		Login                   string
		ContributionsCollection struct {
			PullRequestReviewContributions struct {
				Edges []PullRequestEdges
			} `graphql:"pullRequestReviewContributions(first: $first, after: $after)"`
		} `graphql:"contributionsCollection(from: $from, to: $to, organizationID: $organizationID)"`
	} `graphql:"user(login: $login)"`
}

// and a method to see if a slice of strings contains a given string
func sliceContains(sl []string, name string) bool {
	for _, v := range sl {
		if v == name {
			return true
		}
	}
	return false
}

/*
 * collect together the code used to setup any of the GraphQL queries that
 * we've defined here; first a pair function that we can use to get a new
 * (authenticated) GraphQL client
 */
func getAuthenticatedClient() *githubv4.Client {
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
 * and a function that can be used to convert a list of organization
 * names to a list of organization ID values
 */
func getOrgIdList(client *githubv4.Client) []githubv4.ID {
	// first get the list of organization names that we want to query
	// from either the command line or the configuration file (in that order)
	var orgNameList []string
	inputOrgList := viper.Get("orgList").(string)
	if inputOrgList != "" {
		orgNameList = strings.Split(inputOrgList, ",")
	} else {
		orgNameList = viper.GetStringSlice("orgs")
	}
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

/*
 * determine if a given user is part of  the list of users that make up
 * the input team
 */
func findUserInTeam(teamList []map[string]string, user string) (bool, string) {
	for _, member := range teamList {
		if member["user"] == user {
			return true, member["gitHubId"]
		}
	}
	return false, ""
}

/*
 * define a function that can be used to get the list of users to query for;
 * this function uses either a list of user names or a list of GitHub IDs that
 * were passed on on the command line (using either the '-u, --user-list' flag
 * or the '-i, --github-id-list' flags, respectively)
 *
 * NOTE: it is an error to pass in both a user list and a GitHub ID list, and
 *   it is an error if none of the users or GitHub IDs that were passed in on
 *   the command-line are found in the specified team (however, if some of the
 *   users do match, then the missing users will be skipped and the program
 *   will continue, returning only results for the users that *were* found)
 */
func getUserIdList() []string {
	var userIdList []string
	var teamName string
	var teamList []map[string]string
	// check to see if a list of users was passed in on the command-line
	// (either as a list of user names or a list of user IDs)
	userVal := viper.Get("userList")
	idVal := viper.Get("gitHubIdList")
	if userVal != "" && idVal != "" {
		// if both flags were used, it's an error (we don't know which we should use)
		fmt.Fprintln(os.Stderr, "ERROR: both --userList and --githubIdList were used; only one of these flags can be used at a time")
		os.Exit(-1)
	} else if userVal != "" {
		inputUserList := userVal.(string)
		// if so, split it to get a list of users to retrieve GitHub IDs for (from the
		// config file)
		var userList []string
		if inputUserList != "" {
			userList = strings.Split(inputUserList, ",")
		} else {
			// otherwise, get the list of user IDs from the configuration file
			userList = viper.GetStringSlice("users")
		}
		// retrieve the details for the input team (or the default team if a team
		// was not specified on the comamand line)
		teamName, teamList = getTeamList()
		_, defaultTeamList := getTeamList(viper.GetString("default_team"))
		for _, user := range userList {
			foundUser, memberID := findUserInTeam(teamList, user)
			// if a match was not found, check for a match in the default team
			if !foundUser {
				foundUser, memberID = findUserInTeam(defaultTeamList, user)
				// if a match was still not found in the default team, print a warning and continue
				if !foundUser {
					fmt.Fprintf(os.Stderr, "WARNING: user '%s' not found on the team '%s'; skipping\n", user, teamName)
				}
			}
			// if a match was found, add the user to the list of user IDs to query for
			if foundUser {
				userIdList = append(userIdList, memberID)
				break
			}
		}
	} else if idVal != "" {
		inputIdList := idVal.(string)
		// if so, split it to get a list of user IDs to retrieve GitHub IDs for (from the
		// config file)
		userIdList = strings.Split(inputIdList, ",")
	} else {
		// otherwise, get the list of user IDs from the team (as the default user list)
		_, teamList := getTeamList()
		for _, member := range teamList {
			userIdList = append(userIdList, member["gitHubId"])
		}
	}
	// if neither flag was used or if an empty string was provided for either then it's an error
	if len(userIdList) == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: no matching users found on team '%s' or the default team\n", teamName)
		os.Exit(-1)
	}
	return userIdList
}

/*
 * a function that can be used to get the list of users on the team to compare
 * against
 */
func getTeamList(inputTeamName ...string) (string, []map[string]string) {
	// first, get the name of the team to use for comparison (this value should
	// have been passed in on the command-line)
	var teamList []map[string]string
	teamName := ""
	if len(inputTeamName) > 1 {
		fmt.Fprintf(os.Stderr, "ERROR: only a single team name can be passed in; received %v\n", inputTeamName)
		os.Exit(-1)
	} else if len(inputTeamName) == 1 {
		teamName = inputTeamName[0]
	} else {
		if val := viper.Get("teamName"); val != nil {
			teamName = val.(string)
		}
	}
	if teamName == "" {
		teamName = viper.GetString("default_team")
		if teamName == "" {
			fmt.Fprintf(os.Stderr, "ERROR: team name is a required argument; use the '--team, -t' flag or define a 'default_team' config value\n")
			os.Exit(-1)
		}
	}
	// next, look for that team name under the 'teams' config value
	teamsMap := viper.Get("teams")
	if teamsMap == nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to find the required 'teams' map in the configuration file\n")
		os.Exit(-1)
	} else {
		// if found an entry by that name, then construct a new list of maps of strings
		// strings containing the members of that team
		teamMap := teamsMap.(map[string]interface{})[teamName]
		if teamMap == nil {
			fmt.Fprintf(os.Stderr, "ERROR: unrecognized team name '%s'\n", teamName)
			os.Exit(-1)
		}
		// construct the list of team members as a list of maps of strings to strings
		for _, member := range teamMap.([]interface{}) {
			memberStrMap := map[string]string{}
			for key, val := range member.(map[interface{}]interface{}) {
				memberStrMap[key.(string)] = val.(string)
			}
			teamList = append(teamList, memberStrMap)
		}
	}
	return teamName, teamList
}

/*
 * a function that can be used to get a time window to use for our queries
 */
func getQueryTimeWindow() (githubv4.DateTime, githubv4.DateTime) {
	// first, get the date to start looking back from (this value should
	// have been passed in on the command-line, but defaults to the empty string)
	var endDateTime time.Time
	endDate := viper.Get("endDate").(string)
	if endDate != "" {
		dateTime, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: unable to parse end date '%s'; expected format is '2006-01-02'\n", endDate)
			os.Exit(-1)
		}
		endDateTime = dateTime
	} else {
		// If here, then no end-date was specified, so choose a default value of
		// of the current day at midnight (UTC) and make that the ending date time
		// for our query
		endDateTime = time.Now().UTC().Truncate(time.Hour * 24)
	}
	// then, look back six months from that date time to get the starting date
	// time to define the start of our query window
	monthsBack := -viper.Get("monthsBack").(int)
	startDateTime := endDateTime.AddDate(0, monthsBack, 0)
	// and return the results
	return githubv4.DateTime{startDateTime}, githubv4.DateTime{endDateTime}
}

/*
 * define the function that is used to gather GitHub summary information
 * for the contrributions made by the named user(s) to the named org(s)
 */
func getContributionSummary() {
	// first, get a new GitHub GraphQL API client
	client := getAuthenticatedClient()
	// and then get the list of organization IDs that we want to query
	orgIdList := getOrgIdList(client)
	// define the start and end time of our query window
	startDateTime, endDateTime := getQueryTimeWindow()
	// define a few lists that we'll use later on (to loop over the team members and
	// to skip some members when calculating statistics)
	userIdList := getUserIdList()
	_, teamList := getTeamList()
	// construct the list of GitHub IDs to gather information for
	mySet := mapset.NewSet()
	for _, user := range userIdList {
		mySet.Add(user)
	}
	// gitHubIdList := userIdList
	for _, member := range teamList {
		// gitHubIdList = append(gitHubIdList, member["gitHubId"])
		mySet.Add(member["gitHubId"])
	}
	// initialize the vars map that we'll use when making our query for a summary of contributions
	vars := map[string]interface{}{
		"from": startDateTime,
		"to":   endDateTime,
	}
	// and grab the GitHub IDs from that set as a slice
	gitHubIdList := mySet.ToSlice()
	// initialize a few variables
	var avgPullReqContribs, avgReposWithContribPullReqs,
		avgPullReqReviewContribs, avgReposWithContribPullReqReviews float64
	contribByUserSummary := map[string]interface{}{}
	// loop over the list of GitHub IDs
	for _, gitHubId := range gitHubIdList {
		// convert the input value to a string
		gitHubIdStr := gitHubId.(string)
		// initialize a few variables
		var totalPullReqContribs, totalReposWithContribPullReqs,
			totalPullReqReviewContribs, totalReposWithContribPullReqReviews int
		// set the login value for this query to the current user's GitHub ID
		vars["login"] = githubv4.String(gitHubIdStr)
		// loop over the list of organization IDs and gather contribution
		// information for this GitHub user for all of them
		for _, orgId := range orgIdList {
			// set the organization ID value for this query to the current
			// orgId value
			vars["organizationID"] = orgId
			// and run our query, returning the results in the ContribQuery struct
			err := client.Query(context.Background(), &ContribQuery, vars)
			if err != nil {
				// Handle error.
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			// extract the ContributionsCollection part of the result
			contributionsCollection := ContribQuery.User.ContributionsCollection
			// and use it to accumulate the results for this user to the repositories
			// in this organization
			totalPullReqContribs += contributionsCollection.TotalPullRequestContributions
			totalReposWithContribPullReqs += contributionsCollection.TotalRepositoriesWithContributedPullRequests
			totalPullReqReviewContribs += contributionsCollection.TotalPullRequestReviewContributions
			totalReposWithContribPullReqReviews += contributionsCollection.TotalRepositoriesWithContributedPullRequestReviews
		}
		// and add the contribution details for this user to the summary
		// for the entire team
		if sliceContains(userIdList, gitHubIdStr) {
			contribByUserSummary[gitHubIdStr] = map[string]interface{}{
				"pullReqContribs":                totalPullReqContribs,
				"reposWithContribPullReqs":       totalReposWithContribPullReqs,
				"pullReqReviewContribs":          totalPullReqReviewContribs,
				"reposWithContribPullReqReviews": totalReposWithContribPullReqReviews,
			}
		}
		// add current user contributions (weighted by the number of input GitHub users)
		// to determine the average for each metric for the team
		avgPullReqContribs += float64(totalPullReqContribs) / float64(len(gitHubIdList))
		avgReposWithContribPullReqs += float64(totalReposWithContribPullReqs) / float64(len(gitHubIdList))
		avgPullReqReviewContribs += float64(totalPullReqReviewContribs) / float64(len(gitHubIdList))
		avgReposWithContribPullReqReviews += float64(totalReposWithContribPullReqReviews) / float64(len(gitHubIdList))
	}

	// and add some summary statistics to the output map
	for _, gitHubId := range userIdList {
		userMap := contribByUserSummary[gitHubId].(map[string]interface{})
		userMap["teamPcntPullReqContribs"] = math.Round(((float64(userMap["pullReqContribs"].(int)))/avgPullReqContribs-1)*10000) / 100
		userMap["teamPcntReposWithContribPullReqs"] = math.Round(((float64(userMap["reposWithContribPullReqs"].(int)))/avgReposWithContribPullReqs-1)*10000) / 100
		userMap["teamPcntPullReqReviewContribs"] = math.Round(((float64(userMap["pullReqReviewContribs"].(int)))/avgPullReqReviewContribs-1)*10000) / 100
		userMap["teamPcntReposWithContribPullReqReviews"] = math.Round(((float64(userMap["reposWithContribPullReqReviews"].(int)))/avgReposWithContribPullReqReviews-1)*10000) / 100
	}

	// finally, dump out the contributions as a JSON array of dictionary values
	jsonStr, err := json.MarshalIndent(contribByUserSummary, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(-2)
	}
	fmt.Fprintf(os.Stdout, "%s\n", string(jsonStr))
}

/*
 * define the function that is used to fetch the GitHub contribution information
 * for the contributions made by the named user(s) against repositories under
 * the named org(s)
 */
func fetchContributionList() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := getAuthenticatedClient()
	// and then get the list of organization IDs that we want to query
	orgIdList := getOrgIdList(client)
	// define the start and end time of our query window
	startDateTime, endDateTime := getQueryTimeWindow()
	// define a list that we'll use later on (to loop over the team members and
	// to skip some members when calculating statistics)
	gitHubIdList := getUserIdList()
	// initialize the vars map that we'll use when making our query for PR contributions
	vars := map[string]interface{}{
		"from":  startDateTime,
		"to":    endDateTime,
		"first": githubv4.Int(100),
	}
	contribsByUser := map[string]interface{}{}
	contribsByRepo := map[string]interface{}{}
	// loop over the list of GitHub IDs
	for _, gitHubId := range gitHubIdList {
		// set the login value for this query to the current user's GitHub ID
		vars["login"] = githubv4.String(gitHubId)
		// and initialize a map to that will be used to hold the details for
		// all of the contributions made by this user
		userCommitContribs := []map[string]interface{}{}
		// and loop over the list of Org IDs
		for _, orgId := range orgIdList {
			// set the "organizationID" field and (re)set the "after" field its
			// initial value in the "vars" map
			vars["organizationID"] = orgId
			// define the variable used to track the cursor values as we go
			lastCursor := githubv4.String("")
			// then make requests for the contributions made by this user to this
			// organization (and continue doing so until we reach the end of the
			// list of contributions made by this user to this organization in the
			// specified time period)
			for {
				// set the "after" field to our current "lastCursof" value
				vars["after"] = lastCursor
				// run our query, returning the results in the CommitContributionsMadeQuery struct
				err := client.Query(context.Background(), &ContributionsMadeQuery, vars)
				if err != nil {
					// Handle error.
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				// grab out the list of edges from the pull request contributions
				// made and loop over them
				contribsByRepository := ContributionsMadeQuery.User.ContributionsCollection.CommitContributionsByRepository
				if len(contribsByRepository) == 0 {
					break
				}
				// define a flag we can use to break out of the loop when we reach the end of the list of contributions
				endOfContributions := false
				for _, contribByRepository := range contribsByRepository {
					edges := contribByRepository.Contributions.Edges
					if len(edges) == 0 {
						endOfContributions = true
						break
					}
					for _, edge := range edges {
						// add the details for this edge to the list of commit contributions
						// made by to the appropriate repository
						if _, ok := contribsByRepo[edge.Node.Repository.Url]; !ok {
							// if here, then we haven't seen this repository yet so create a new entry for it
							contribsByRepo[edge.Node.Repository.Url] = map[string]interface{}{
								"repositoryName":     edge.Node.Repository.Name,
								"totalContributions": edge.Node.CommitCount,
							}
						} else {
							// else just increment the number of contributions made to this repository
							repoContribsMap := contribsByRepo[edge.Node.Repository.Url].(map[string]interface{})
							if currentCount, ok := repoContribsMap["totalContributions"].(int); ok {
								repoContribsMap["totalContributions"] = currentCount + edge.Node.CommitCount
							}
						}
						// and add the details for this edge to the list of commit contributions
						// made by this user (these edges are organized by date/repository pairs)
						userCommitContribs = append(userCommitContribs, map[string]interface{}{
							"repositoryName":   edge.Node.Repository.Name,
							"numContributions": edge.Node.CommitCount,
							"contributedAt":    edge.Node.OccurredAt,
						})
						// and save the cursor value for this edge for use later on
						lastCursor = edge.Cursor
					}
				}
				// if we've reached the end of the list of contributions, break out of the loop
				if endOfContributions {
					break
				}
			}
		}
		// add pull requests for this user to the complete list of pull
		// requests by user
		if _, ok := contribsByUser["ByUser"]; ok {
			if val, ok := contribsByUser["ByUser"].([]map[string]interface{}); ok {
				contribsByUser["ByUser"] = append(val, map[string]interface{}{
					gitHubId: userCommitContribs,
				})
			}
		} else {
			contribsByUser["ByUser"] = append([]map[string]interface{}{}, map[string]interface{}{
				gitHubId: userCommitContribs,
			})
		}
	}
	// finally add an "AllUsers" entry to the list of contributions made by all users to each repository
	contribsByUser["AllUsers"] = contribsByRepo

	// and return the resulting list
	return contribsByUser
}

/*
 * define the function that is used to print (as a JSON string) the GitHub
 * contribution information for the contributions made by the named user(s) against
 * repositories under the named org(s)
 */
func getContributionList() {
	// fetch the list of PRs made by the named user(s) against repositories
	// under the named org(s)
	pullRequestsByUser := fetchContributionList()
	// and dump out the results
	jsonStr, err := json.MarshalIndent(pullRequestsByUser, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(-2)
	}
	fmt.Fprintf(os.Stdout, "%s\n", string(jsonStr))
}

/*
 * define the function that is used to fetch the GitHub pull request information
 * for the pull requests made by the named user(s) against repositories under
 * the named org(s)
 */
func fetchPullRequestList() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := getAuthenticatedClient()
	// and then get the list of organization IDs that we want to query
	orgIdList := getOrgIdList(client)
	// define the start and end time of our query window
	startDateTime, endDateTime := getQueryTimeWindow()
	// define a list that we'll use later on (to loop over the team members and
	// to skip some members when calculating statistics)
	gitHubIdList := getUserIdList()
	// initialize the vars map that we'll use when making our query for PR contributions
	vars := map[string]interface{}{
		"from":  startDateTime,
		"to":    endDateTime,
		"first": githubv4.Int(100),
	}
	pullRequestsByUser := map[string]interface{}{}
	prsByRepo := map[string]interface{}{}
	// loop over the list of GitHub IDs
	for _, gitHubId := range gitHubIdList {
		// set the login value for this query to the current user's GitHub ID
		vars["login"] = githubv4.String(gitHubId)
		// and initialize a map to that will be used to hold the details for
		// all of the pull requests made by this user
		userPullRequests := []map[string]interface{}{}
		// and loop over the list of Org IDs
		for _, orgId := range orgIdList {
			// set the "organizationID" field and (re)set the "after" field its
			// initial value in the "vars" map
			vars["organizationID"] = orgId
			// define the variable used to track the cursor values as we go
			lastCursor := githubv4.String("")
			// then make requests for the pull requests made by this user to this
			// organization (and continue doing so until we reach the end of the
			// list of pull requests made by this user to this organization in the
			// specified time period)
			for {
				// set the "after" field to our current "lastCursof" value
				vars["after"] = lastCursor
				// run our query, returning the results in the PullRequestsMadeQuery struct
				err := client.Query(context.Background(), &PullRequestsMadeQuery, vars)
				if err != nil {
					// Handle error.
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				// grab out the list of edges from the pull request contributions
				// made and loop over them
				edges := PullRequestsMadeQuery.User.ContributionsCollection.PullRequestContributions.Edges
				// if nothing was returned, then we've found all of the contributions
				// from this user to this organization so break out of the loop
				if len(edges) == 0 {
					break
				}
				fmt.Fprintf(os.Stderr, "Found %d pull request contributions for user %s to org %s\n", len(edges), gitHubId, orgId)
				for _, edge := range edges {
					// save some typing later by grabbing the pull request associated with this edge
					pullReq := edge.Node.PullRequest
					// if the pull rquest was closed as merged, then add the details for this
					// edge to the list of commit contributions made by to the appropriate
					// repository
					if pullReq.Closed && pullReq.Merged {
						if _, ok := prsByRepo[pullReq.Repository.Url]; !ok {
							// if here, then we haven't seen this repository yet so create a new entry for it
							prsByRepo[pullReq.Repository.Url] = map[string]interface{}{
								"repositoryName":     pullReq.Repository.Name,
								"totalContributions": 1,
							}
						} else {
							// else just increment the number of contributions made to this repository
							repoContribsMap := prsByRepo[pullReq.Repository.Url].(map[string]interface{})
							if currentCount, ok := repoContribsMap["totalContributions"].(int); ok {
								repoContribsMap["totalContributions"] = currentCount + 1
							}
						}
					}
					// determine how long the pull request was open (or has been open if it's still open)
					// after it was created along with the time since the first commit was made
					daysOpen := 0.0
					daysSinceFirstCommit := 0.0
					firstCommitAt := pullReq.Commits.Edges[0].Node.Commit.CommittedDate.Time
					if pullReq.Closed && !pullReq.Merged {
						// pull request was closed but not merged
						daysOpen = math.Round(pullReq.ClosedAt.Sub(pullReq.CreatedAt.Time).Hours()/24.0*10000) / 10000
						daysSinceFirstCommit = math.Round(pullReq.ClosedAt.Sub(firstCommitAt).Hours()/24.0*10000) / 10000
					} else if pullReq.Merged {
						// pull request was merged
						daysOpen = math.Round(pullReq.MergedAt.Sub(pullReq.CreatedAt.Time).Hours()/24.0*10000) / 10000
						daysSinceFirstCommit = math.Round(pullReq.MergedAt.Sub(firstCommitAt).Hours()/24.0*10000) / 10000
					} else {
						// pull request is still open today (so used time elapsed since it was created)
						daysOpen = math.Round(time.Since(pullReq.CreatedAt.Time).Hours()/24.0*10000) / 10000
						daysSinceFirstCommit = math.Round(time.Since(firstCommitAt).Hours()/24.0*10000) / 10000
					}
					// add the details for this edge to the list of pull requests
					// made by this user
					userPullRequests = append(userPullRequests, map[string]interface{}{
						"author":         pullReq.Author.Login,
						"closed":         pullReq.Closed,
						"closedAt":       pullReq.ClosedAt,
						"createdAt":      pullReq.CreatedAt,
						"daysOpen":       daysOpen,
						"daysWorked":     math.Max(daysOpen, daysSinceFirstCommit),
						"firstCommitAt":  firstCommitAt,
						"merged":         pullReq.Merged,
						"mergedAt":       pullReq.MergedAt,
						"repositoryName": pullReq.Repository.Name,
						"title":          pullReq.Title,
						"url":            pullReq.Url,
					})
					// and save the cursor value for this edge for use later on
					lastCursor = edge.Cursor
				}
			}
		}
		// add pull requests for this user to the complete list of pull
		// requests by user
		if _, ok := pullRequestsByUser["ByUser"]; ok {
			if val, ok := pullRequestsByUser["ByUser"].([]map[string]interface{}); ok {
				pullRequestsByUser["ByUser"] = append(val, map[string]interface{}{
					gitHubId: userPullRequests,
				})
			}
		} else {
			pullRequestsByUser["ByUser"] = append([]map[string]interface{}{}, map[string]interface{}{
				gitHubId: userPullRequests,
			})
		}
	}
	// finally add an "AllUsers" entry to the list of contributions made by all users to each repository
	pullRequestsByUser["AllUsers"] = prsByRepo
	return pullRequestsByUser
}

/*
 * define the function that is used to print (as a JSON string) the GitHub pull
 * request information for the pull requests made by the named user(s) against
 * repositories under the named org(s)
 */
func getPullRequestList() {
	// fetch the list of PRs made by the named user(s) against repositories
	// under the named org(s)
	pullRequestsByUser := fetchPullRequestList()
	// and dump out the results
	jsonStr, err := json.MarshalIndent(pullRequestsByUser, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(-2)
	}
	fmt.Fprintf(os.Stdout, "%s\n", string(jsonStr))
}

/*
 * define the function that is used to fetch GitHub pull request information
 * for the pull requests made by the named user(s) against repositories under
 * the named org(s)
 */
func fetchPullRequestReviewList() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := getAuthenticatedClient()
	// and then get the list of organization IDs that we want to query
	orgIdList := getOrgIdList(client)
	// define the start and end time of our query window
	startDateTime, endDateTime := getQueryTimeWindow()
	// define a list that we'll use later on (to loop over the team members and
	// to skip some members when calculating statistics)
	gitHubIdList := getUserIdList()
	// initialize the vars map that we'll use when making our query for PR review contributions
	vars := map[string]interface{}{
		"from":  startDateTime,
		"to":    endDateTime,
		"first": githubv4.Int(100),
	}
	pullRequestReviewsByUser := map[string]interface{}{}
	prReviewsByRepo := map[string]interface{}{}
	// loop over the list of GitHub IDs
	for _, gitHubId := range gitHubIdList {
		// set the login value for this query to the current user's GitHub ID
		vars["login"] = githubv4.String(gitHubId)
		// and initialize a map to that will be used to hold the details for
		// all of the pull request reviews made by this user
		userPullRequestReviews := []map[string]interface{}{}
		// and loop over the list of Org IDs
		for _, orgId := range orgIdList {
			// set the "organizationID" field and (re)set the "after" field its
			// initial value in the "vars" map
			vars["organizationID"] = orgId
			// define the variable used to track the cursor values as we go
			lastCursor := githubv4.String("")
			// then make requests for the pull request reviews made by this user
			// to this organization (and continue doing so until we reach the end
			// of the list of pull request reviews made by this user to this
			// organization in the specified time period)
			for {
				// set the "after" field to our current "lastCursof" value
				vars["after"] = lastCursor
				// run our query, returning the results in the PullRequestReviewsPerformedQuery struct
				err := client.Query(context.Background(), &PullRequestReviewsPerformedQuery, vars)
				if err != nil {
					// Handle error.
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				// grab out the list of edges from the pull request review
				// contributions made and loop over them
				edges := PullRequestReviewsPerformedQuery.User.ContributionsCollection.PullRequestReviewContributions.Edges
				// if nothing was returned, then we've found all of the contributions
				// from this user to this organization so break out of the loop
				if len(edges) == 0 {
					break
				}
				fmt.Fprintf(os.Stderr, "Found %d pull request review contributions for user %s to org %s\n", len(edges), gitHubId, orgId)
				for _, edge := range edges {
					// save some typing later by grabbing the pull request associated with this edge
					pullReq := edge.Node.PullRequest
					// if the pull rquest review is for a pull request that was closed as merged,
					// then add the details for this edge to the list of commit contributions
					// made by to the appropriate repository
					if pullReq.Closed && pullReq.Merged {
						if _, ok := prReviewsByRepo[pullReq.Repository.Url]; !ok {
							// if here, then we haven't seen this repository yet so create a new entry for it
							prReviewsByRepo[pullReq.Repository.Url] = map[string]interface{}{
								"repositoryName":     pullReq.Repository.Name,
								"totalContributions": 1,
							}
						} else {
							// else just increment the number of contributions made to this repository
							repoContribsMap := prReviewsByRepo[pullReq.Repository.Url].(map[string]interface{})
							if currentCount, ok := repoContribsMap["totalContributions"].(int); ok {
								repoContribsMap["totalContributions"] = currentCount + 1
							}
						}
					}
					// add the details for this edge to the list of pull request
					// reviews made by this user
					userPullRequestReviews = append(userPullRequestReviews, map[string]interface{}{
						"author":         pullReq.Author.Login,
						"closed":         pullReq.Closed,
						"merged":         pullReq.Merged,
						"repositoryName": pullReq.Repository.Name,
						"title":          pullReq.Title,
						"url":            pullReq.Url,
					})
					// and save the cursor value for this edge for use later on
					lastCursor = edge.Cursor
				}
			}
		}
		// add pull request reviews for this user to the complete list of pull
		// requests by user
		if _, ok := pullRequestReviewsByUser["ByUser"]; ok {
			if val, ok := pullRequestReviewsByUser["ByUser"].([]map[string]interface{}); ok {
				pullRequestReviewsByUser["ByUser"] = append(val, map[string]interface{}{
					gitHubId: userPullRequestReviews,
				})
			}
		} else {
			pullRequestReviewsByUser["ByUser"] = append([]map[string]interface{}{}, map[string]interface{}{
				gitHubId: userPullRequestReviews,
			})
		}
	}
	// finally add an "AllUsers" entry to the list of contributions made by all users to each repository
	pullRequestReviewsByUser["AllUsers"] = prReviewsByRepo
	return pullRequestReviewsByUser
}

/*
 * define the function that is used to print (as a JSON string) the GitHub pull
 * request information for the pull requests made by the named user(s) against
 * repositories under the named org(s)
 */
func getPullRequestReviewList() {
	// fetch the list of PR reviews made by the named user(s) against repositories
	// under the named org(s)
	pullRequestReviewsByUser := fetchPullRequestReviewList()
	// and dump out the results
	jsonStr, err := json.MarshalIndent(pullRequestReviewsByUser, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(-2)
	}
	fmt.Fprintf(os.Stdout, "%s\n", string(jsonStr))
}

/*
 * define the function that is used to print (as a JSON string) the information
 * for all of the pull request contributions (both pull requests, and pull request reviews)
 * made by the named user(s) against repositories under the named org(s)
 */
func getAllContribsByTypeList() {
	// initialize the map used to track the contributions (grouped by type of contribution)
	contribsByUser := map[string]interface{}{}
	// first, fetch the list of PRs made by the named user(s) against repositories
	// under the named org(s)
	contribsByUser["pullRequests"] = fetchPullRequestList()
	// then append onto that the list of PR reviews made by the named user(s) against
	// repositories under the named org(s)
	contribsByUser["pullRequestReviews"] = fetchPullRequestReviewList()
	// and dump out the results
	jsonStr, err := json.MarshalIndent(contribsByUser, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(-2)
	}
	fmt.Fprintf(os.Stdout, "%s\n", string(jsonStr))
}

/*
	query MyQuery {
	search(query: "-orb in:name org:circleci-public", type: REPOSITORY, first: 100) {
		nodes {
			... on Repository {
				name
				isPrivate
				isArchived
			}
		}
		repositoryCount
	}
	}
*/

/*
 * define the function that is used to fetch a list of the Orb repositories
 * managed by the team under the named organizations
 */
func fetchOrbRepositoryList() map[string]interface{} {
	// first, get a new GitHub GraphQL API client
	client := getAuthenticatedClient()
	// initialize the vars map that we'll use when making our query for PR review contributions
	vars := map[string]interface{}{
		"query": githubv4.String("-orb in:name org:CircleCI-Public"),
		"type":  githubv4.SearchTypeRepository,
		"first": githubv4.Int(100),
	}
	// and initialize a map to that will be used to hold the details for
	// all of the pull request reviews made by this user
	orbRepositoryList := map[string]interface{}{}
	// run our query, returning the results in the PullRequestReviewsPerformedQuery struct
	err := client.Query(context.Background(), &RepositorySearchQuery, vars)
	if err != nil {
		// Handle error.
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err)
		os.Exit(1)
	}
	// grab out the list of edges from the pull request review
	// contributions made and loop over them
	edges := RepositorySearchQuery.Search.Edges
	fmt.Fprintf(os.Stderr, "Found %d orb repositories in CircleCI-Public Org\n", len(edges))
	for _, edge := range edges {
		// if here, then we haven't seen this repository yet so create a new entry for it
		orbRepositoryList[edge.Node.Repository.Name] = map[string]interface{}{
			"private":  edge.Node.Repository.IsPrivate,
			"archived": edge.Node.Repository.IsArchived,
			"url":      edge.Node.Repository.Url,
		}
	}
	return orbRepositoryList
}

/*
 * define the function that is used to print (as a JSON string) the information
 * for all of the pull request contributions (both pull requests, and pull request reviews)
 * made by the named user(s) against repositories under the named org(s)
 */
func getOrbRepositoryList() {
	// get the list of orb repositories
	orbRepoList := fetchOrbRepositoryList()
	// and dump out the results
	jsonStr, err := json.MarshalIndent(orbRepoList, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(-2)
	}
	fmt.Fprintf(os.Stdout, "%s\n", string(jsonStr))
}

// setup the root command, a couple of subcommands and the variables that we set using the
// flags in those commands
var (
	// Used for flags.
	outputFile   string
	userList     string
	gitHubIdList string
	compTeam     string
	orgList      string
	endDate      string
	monthsBack   int

	rootCmd = &cobra.Command{
		Use:   "getGhContributions",
		Short: "Gets contribution information for the named GitHub users",
		Long: `Gathers contribution information for a named set of GitHub users
(provided on the command-line or in an associated configuration file) to
any repositories under the named set of GitHub organizations.`,
	}
)

var (
	orbRepositoryList = &cobra.Command{
		Use:   "orbRepositoryList",
		Short: "Generates a list of Orb repository names (with additional metadata)",
		Long: `Constructs a list of the current Orb repositories that are in the
CircleCI-Public GitHub organization.`,
		Run: func(cmd *cobra.Command, args []string) {
			getOrbRepositoryList()
		},
	}
)

var (
	summaryOfContribs = &cobra.Command{
		Use:   "summaryOfContribs",
		Short: "Generates a summary (including statistics) of contributions (by user)",
		Long: `Constructs a summary (by user) of all of the contributions that each
of the input users made to any repository to any of the repositories in
the named set of GitHub organizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			getContributionSummary()
		},
	}
)

var (
	pullRequestList = &cobra.Command{
		Use:   "pullRequestList",
		Short: "Generates a list the pull requests made (by user)",
		Long: `Constructs a list (by user) of all of the pull requests that each
of the input users made to any repository to any of the repositories in
the named set of GitHub organizations (including the title, status, url,
and repository name) for each pull request submitted by that user.`,
		Run: func(cmd *cobra.Command, args []string) {
			getPullRequestList()
		},
	}
)

var (
	pullRequestReviewList = &cobra.Command{
		Use:   "pullRequestReviewList",
		Short: "Generates a list the pull request reviews made (by user)",
		Long: `Constructs a list (by user) of all of the pull request reviews
that each of the input users performed in any repository to any of the
repositories in the named set of GitHub organizations (including the
title, status, url, and repository name) for each pull request submitted
by that user.`,
		Run: func(cmd *cobra.Command, args []string) {
			getPullRequestReviewList()
		},
	}
)

var (
	contributionList = &cobra.Command{
		Use:   "contributionList",
		Short: "Generates a list of any contributions made (by user)",
		Long: `Constructs a list (by user) of any contributions made
(commits and pull requests) by each of the input users against any
of the repositories in the named set of GitHub organizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			getContributionList()
		},
	}
)

var (
	allContribsByTypeList = &cobra.Command{
		Use:   "allContribsByTypeList",
		Short: "Generates a list of all contributions made (by type and user)",
		Long: `Constructs a list (by contribution type and user) of all of
the contributions made (pull requests and pull request reviews) by each of
the input users performed against any of the repositories in the named set
of GitHub organizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			getAllContribsByTypeList()
		},
	}
)

// used to pull in a configuration from the appropriate file (if it exists)
func initConfig() {
	// set the (YAML) configuration file name
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	// set the locations to look for that configuration file in
	viper.AddConfigPath("/etc/appname/")
	viper.AddConfigPath("$HOME/.appname")
	viper.AddConfigPath(".")
	// read the configuration file
	err := viper.ReadInConfig()
	// if there was an error reading the configuration file, handle it
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	} else {
		fmt.Fprintf(os.Stderr, "Using configuration file: %s\n", viper.ConfigFileUsed())
	}
}

func main() {
	// initialize the configuration
	cobra.OnInitialize(initConfig)

	// and a set of optional global flags (the flags that are optional for all of our subcommands)
	rootCmd.PersistentFlags().StringVarP(&outputFile, "file", "f", "", "file/stream to output data to (defaults to standard output)")
	rootCmd.PersistentFlags().StringVarP(&userList, "user-list", "u", "", "list of users to gather contributions for")
	rootCmd.PersistentFlags().StringVarP(&gitHubIdList, "github-id-list", "i", "", "list of GitHub IDs to gather contributions for")
	rootCmd.PersistentFlags().StringVarP(&orgList, "org-list", "o", "", "list of orgs to gather contributions to")
	rootCmd.PersistentFlags().IntVarP(&monthsBack, "months-back", "m", 6, "length of time to look back (in months; defaults to 6)")
	rootCmd.PersistentFlags().StringVarP(&endDate, "end-date", "d", "", "date to start looking back from (in YYYY-MM-DD format)")

	// this flag is only used for the summaryOfContribs command
	summaryOfContribs.PersistentFlags().StringVarP(&compTeam, "team", "t", "", "name of team to compare contributions against")

	// and add our subcommands to the root command
	rootCmd.AddCommand(pullRequestList)
	rootCmd.AddCommand(pullRequestReviewList)
	rootCmd.AddCommand(summaryOfContribs)
	rootCmd.AddCommand(contributionList)
	rootCmd.AddCommand(allContribsByTypeList)
	rootCmd.AddCommand(orbRepositoryList)

	// bind the flags defined above to viper (so that we can use viper to retrieve the values)
	viper.BindPFlag("outputFile", rootCmd.PersistentFlags().Lookup("file"))
	viper.BindPFlag("userList", rootCmd.PersistentFlags().Lookup("user-list"))
	viper.BindPFlag("gitHubIdList", rootCmd.PersistentFlags().Lookup("github-id-list"))
	viper.BindPFlag("teamName", summaryOfContribs.PersistentFlags().Lookup("team"))
	viper.BindPFlag("orgList", rootCmd.PersistentFlags().Lookup("org-list"))
	viper.BindPFlag("monthsBack", rootCmd.PersistentFlags().Lookup("months-back"))
	viper.BindPFlag("endDate", rootCmd.PersistentFlags().Lookup("end-date"))

	// finally, execute the root command to trigger the underlying process for this app
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
