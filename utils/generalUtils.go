/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
)

// and a method to see if a slice of strings contains a given string
func SliceContains(sl []string, name string) bool {
	for _, v := range sl {
		if v == name {
			return true
		}
	}
	return false
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
func GetUserIdList() []string {
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
		teamName, teamList = GetTeamList()
		for _, user := range userList {
			foundUser, memberID := findUserInTeam(teamList, user)
			// if a match was not found, check for a match in the default team
			if !foundUser {
				fmt.Fprintf(os.Stderr, "WARNING: user '%s' not found on the team '%s'; skipping\n", user, teamName)
				continue
			}
			// a match was found, add the user to the list of user IDs to query for
			userIdList = append(userIdList, memberID)
		}
	} else if idVal != "" {
		inputIdList := idVal.(string)
		// if so, split it to get a list of user IDs to retrieve GitHub IDs for (from the
		// config file)
		userIdList = strings.Split(inputIdList, ",")
	} else {
		// otherwise, get the list of user IDs from the team (as the default user list)
		_, teamList := GetTeamList()
		for _, member := range teamList {
			userIdList = append(userIdList, member["gitHubId"])
		}
	}
	// if neither flag was used or if an empty string was provided for either then it's an error
	if len(userIdList) == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: no matching users found on team '%s'\n", teamName)
		os.Exit(-1)
	}
	return userIdList
}

/*
 * a function that can be used to get the list of users on the team to compare
 * against
 */
func GetTeamList(inputTeamName ...string) (string, []map[string]string) {
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
func GetQueryTimeWindow() (githubv4.DateTime, githubv4.DateTime) {
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
