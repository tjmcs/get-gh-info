/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

/*
 * determine if a given user is part of  the list of users that make up
 * the input team
 */
func findUserInTeam(teamList []map[string]string, user string) (bool, string) {
	for _, member := range teamList {
		if member["user"] == user {
			return true, member["githubid"]
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
		teamName, teamList = GetTeamMembers()
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
		_, teamList := GetTeamMembers()
		for _, member := range teamList {
			userIdList = append(userIdList, member["githubid"])
		}
	}
	// if neither flag was used or if an empty string was provided for either then it's an error
	if len(userIdList) == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: no matching users found on team '%s'\n", teamName)
		os.Exit(-2)
	}
	return userIdList
}

/*
 * a function that can be used to get the list of users on the team to compare
 * against
 */
func GetTeamMembers(inputTeamName ...string) (string, []map[string]string) {
	// first, get the name of the team to use for comparison (this value should
	// have been passed in on the command-line)
	var teamList []map[string]string
	teamName := ""
	if len(inputTeamName) > 1 {
		fmt.Fprintf(os.Stderr, "ERROR: only a single team name can be passed in; received %v\n", inputTeamName)
		os.Exit(-3)
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
			os.Exit(-4)
		}
	}
	// next, look for that team name under the 'teams' config value
	teamsMap := viper.Get("teams")
	if teamsMap == nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to find the required 'teams' map in the configuration file\n")
		os.Exit(-5)
	} else {
		// if found an entry by that name, then construct a new list of maps of strings
		// strings containing the members of that team
		teamMap := teamsMap.(map[string]interface{})[teamName]
		if teamMap == nil {
			fmt.Fprintf(os.Stderr, "ERROR: unrecognized team name '%s'\n", teamName)
			os.Exit(-6)
		}
		// construct the list of team members as a list of maps of strings to strings
		for _, member := range teamMap.([]interface{}) {
			memberStrMap := map[string]string{}
			for key, val := range member.(map[string]interface{}) {
				memberStrMap[key] = val.(string)
			}
			teamList = append(teamList, memberStrMap)
		}
	}
	return teamName, teamList
}

/*
 * get a list of the member GitHub IDs from the intput team members map
 */
func GetTeamMemberIds(teamMembers []map[string]string) []string {
	var memberLogins []string
	for _, member := range teamMembers {
		memberLogins = append(memberLogins, member["githubid"])
	}
	return memberLogins
}

/*
 * a function that can be used to extract all of the repositories for a given team
 * from the "team to repository map"; that file looks something like this:
 *
 *   - group: <team name>
 *     repositories:
 *       - <repo name>
 *       - <repo name>
 *       - <repo name>
 *       ...
 *     children:
 *       - group: <team name>
 *         repositories:
 *           - <repo name>
 *           - <repo name>
 *           - <repo name>
 *   - group: <team name>
 *     repositories:
 *       - <repo name>
 *       - <repo name>
 *       - <repo name>
 *       ...
 *
 * as such, we need to recursively walk the tree in order to find all of the repositories
 * that might be managed by a given team; what will be returned will be an array of maps,
 * where each map will contain the following keys:
 *   - "url": url associated with the repository
 *	 - "tags": a list of the tags for that repository (used to group repositories together)
 *
 */
func getTeamRepoMappingList(repoMapping []map[string]interface{}, teamName string) []map[string]interface{} {
	// initialize the list of repository mappings that will be returned
	repoMappingList := []map[string]interface{}{}
	// look for the team name in the list of groups at this level
	for _, group := range repoMapping {
		groupMap := group
		if groupMap["group"] == teamName {
			// if we found the team name, then return the list of all of the repository
			// mappings that fall under this part of the tree
			for _, repo := range groupMap["repositories"].([]interface{}) {
				repoMappingList = append(repoMappingList, convInterToInterMapToStringToInterMap(repo.(map[interface{}]interface{})))
			}
			// including the repository mappings for any children of this team
			if groupMap["children"] != nil {
				for _, childMap := range groupMap["children"].([]interface{}) {
					mapAsStringMap := convInterToInterMapToStringToInterMap(childMap.(map[interface{}]interface{}))
					for key, val := range mapAsStringMap {
						childTeam := ""
						if key == "group" {
							childTeam = val.(string)
						}
						// if there is no team name associated with this child, then skip it
						if childTeam == "" {
							continue
						}
						// and recursively call this function to get the list of repositories
						// mappings for this child group as well
						tmpListOfMaps := []map[string]interface{}{mapAsStringMap}
						subTeamRepoMapping := getTeamRepoMappingList(tmpListOfMaps, childTeam)
						repoMappingList = append(repoMappingList, subTeamRepoMapping...)
					}
				}
			}
			// and return the resulting list of repository mappings
			if len(repoMappingList) > 0 {
				return repoMappingList
			}
		} else if groupMap["children"] != nil {
			// if the team name for this group doesn't match, then look for it in the children of this group
			tmpChildren := []map[string]interface{}{}
			for _, childMap := range groupMap["children"].([]interface{}) {
				tmpChildren = append(tmpChildren, convInterToInterMapToStringToInterMap(childMap.(map[interface{}]interface{})))
			}
			teamRepoMappingList := getTeamRepoMappingList(tmpChildren, teamName)
			if len(teamRepoMappingList) > 0 {
				// if we found more entries, add them to our list
				return append(repoMappingList, teamRepoMappingList...)
			}
		}
	} // and if we get to here, then we didn't find a match in this group, so keep looking

	// finally, if we didn't find anything matching this team name anywhere in the tree
	// structure we just searched, then return nothing
	return nil
}

/*
 * a function that can be used to retrieve the list of repositories that are
 * owned by a given team
 */
func GetTeamRepos(inputTeamName ...string) (string, []string) {
	// first, get the name of the team we're looking for
	teamName := ""
	if len(inputTeamName) > 1 {
		fmt.Fprintf(os.Stderr, "ERROR: only a single team name can be passed in; received %v\n", inputTeamName)
		os.Exit(-3)
	} else if len(inputTeamName) == 1 {
		teamName = inputTeamName[0]
	} else {
		if val := viper.Get("teamName"); val != nil {
			teamName = val.(string)
		}
	}
	// if we didn't find a team, use the default team name from the configuration (if it exists)
	if teamName == "" {
		teamName = viper.GetString("default_team")
		if teamName == "" {
			fmt.Fprintf(os.Stderr, "ERROR: team name is a required argument; use the '--team, -t' flag or define a 'default_team' config value\n")
			os.Exit(-4)
		}
	}
	// next, retrieve the mapping of teams to repositories that was either
	// passed in on the command-line or read from the configuration file
	repoMappingFile := viper.Get("repoMappingFile")
	if repoMappingFile == nil || repoMappingFile == "" {
		// if the repo mapping wasn't passed in on the command-line, and wasn't
		// included in the configuration file, then use the default repo mapping
		// (if it exists)
		repoMappingFile = viper.GetString("default_repo_mapping")
		if repoMappingFile == "" {
			// if we still didn't find it, then exit with an error
			fmt.Fprintf(os.Stderr, "ERROR: unable to find the required 'repoMapping' filename\n")
			os.Exit(-7)
		}
	}
	// read the repo mapping file into a map of strings to interfaces
	teamToRepoMap := ReadYamlFile(repoMappingFile.(string))
	// and extract the list of repositories that are owned by that team from the map
	teamRepoMapping := getTeamRepoMappingList(teamToRepoMap, teamName)
	if teamRepoMapping == nil {
		fmt.Fprintf(os.Stderr, "ERROR: unrecognized team name '%s'; could not retrieve repository mappings\n", teamName)
		os.Exit(-8)
	}
	// flatten out the resulting mappings to get a list of repositories "managed" by this team
	// or one of its subteams
	teamRepos := []string{}
	for _, entry := range teamRepoMapping {
		splitString := strings.Split(entry["url"].(string), "/")
		teamRepos = append(teamRepos, strings.Join(splitString[len(splitString)-2:], "/"))
	}
	return teamName, teamRepos
}
