# Tracking GitHub Contributions

This directory contains a Go program that's used to gather either information regarding contributions made by an input list of team members to the repositories contained in an input list of GitHub organizations or information about the issues and pull requests in the repositories themselves. The program gathers this data via GitHub's GraphQL API, then outputs the data to a named output file in JSON format (or to the standard output stream if an output file isn't specified).

Users can either pass the input arguments for the program in on the command line using a set of defined command line flags or retrieve them from the set of "default values" defined in an associated configuration file. By default, the program looks for these default values in the  `~/.config/getGhInfo` file, but if that file doesn't exist, then it uses the values defined in the `config.yml` file in this repository instead. Of course, If the user wants to use a different name for their configuration file, then a command-line flag also exists for that purpose.

## Usage

The easiest way to see the options supported the command-line interface (or CLI) for this app is to take advantage of the help support that's built into the app. Here, for example, is the general (top-level) help available through the CLI:

```bash
Gathers the requested information from GitHub using the GitHub GraphQL API
(where the input parameters for the query to run are provided either on the
command-line or in an associated configuration file) and outputs the results

Usage:
  getGhInfo [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  repo        Gather repository-related data
  user        Gather user-related data

Flags:
  -c, --config string     configuration file to use
  -f, --file string       file/stream for output (defaults to stdout)
  -h, --help              help for getGhInfo
  -o, --org-list string   list of orgs to gather information from

Use "getGhInfo [command] --help" for more information about a command.
```

As you can see in that output, there are two main commands available:

1. the `user` command, which gathers information related to user contributions to the repositories in the specified GitHub organization (or organizations), and
2. the `repo` command, which gathers information related to the issues and pull requests in those same repositories

The following sections provide more detailed examples of how to use both of these commands.

### Obtaining information about user contributions

The `user` command retrieves information about the contributions that a given user or set of users made to the repositories in a defined GitHub organization (or a defined set of GitHub organizations). The help output for that command shows the sorts of options that are available to you as a user:

```bash
The subcommand used as the root for all queries for user-related data

Usage:
  getGhInfo user [command]

Available Commands:
  contribSummary Generates a summary (including statistics) of contributions
  contribs       Generates a list of commits and PRs made
  contribsByType Generates a list of PRs and PR reviews made
  prList         Generates a list the pull requests made
  prReviews      Generates a list the pull request reviews made

Flags:
  -w, --complete-weeks          only output complete weeks (starting Monday)
  -i, --github-id-list string   list of GitHub IDs to gather contributions for
  -h, --help                    help for user
  -l, --lookback-time string    'lookback' time window (eg. 10d, 3w, 2m, 1q, 1y)
  -d, --ref-date string         reference date for time window (YYYY-MM-DD)
  -t, --team string             name of team to gather data for or compare against
  -u, --user-list string        list of users to gather contributions for

Global Flags:
  -c, --config string     configuration file to use
  -f, --file string       file/stream for output (defaults to stdout)
  -o, --org-list string   list of orgs to gather information from

Use "getGhInfo user [command] --help" for more information about a command.
```

The following list describes the sub-commands supported by the `user` command, including a brief description of each sub-command's output:

* **The `contribSummary` sub-command** - generates a summary of the contributions made by each user in the input list of users to repositories in the named GitHub organizations, including the number of pull requests, pull request reviews, number of repositories that they have contributed pull requests to, and number of repositories that they have contributed pull request reviews to. In addition, the app adds values to the summary that show (as a percentage) how the values for each user in the input user list compare with the average for all users in the input team.
* **The `contribs` sub-command** - generates a list of the total number of commits made by all users in the defined list of users to repositories in the defined GitHub organizations, along with a detailed list (broken out by user) of the number of commits that each user in the defined list of users made to those same repositories (for historical reasons the API call used organizes the data for each repository by the date on which a user made these commits, with separate entries for each date/repository combination).
* **The `contribsByType` sub-command** - generates a list of the of the total number of pull requests and pull request reviews made by all users in the defined list of users to repositories in the defined GitHub organizations, along with a detailed list (broken out by user) of the details for the pull requests and pull reviews that each user in the list user made to those same repositories.
* **The `prList` sub-command** - generates a list of the of the total number of pull requests made by all users in the defined list of users to repositories in the defined GitHub organizations, along with a detailed list (broken out by user) of the details for the pull requests that each user in the list user made to those same repositories.
* **The `prReviews` sub-command** - generates a list of the of the total number of pull request reviews made by all users in the defined list of users to repositories in the defined GitHub organizations, along with a detailed list (broken out by user) of the details for the pull request reviews that each user in the list user made to those same repositories.

#### Flags used to control output

The preceding  `user` sub-commands all support a common set of command-line flags, providing users with the ability to control the queries made against the GitHub GraphQL interface and, as a result, the output from this app.

##### The `-t, --team` flag

With this flag, you can define the name of the team that you want to retrieve data for or compare user contributions against (depending on the sub-command). For most of the `user` sub-commands this flag can replace use of the `-u, --user-list` or `-i, --github-id-list` flags when defining the list of users (providing a shorthand method of asking for information about the contributions for all of the users in the named team). However, when you use this flag with either the `-u, --user-list` flag or the `-i, --github-id-list` flag, the app skips any users in the list passed in using either of those two flags that aren't members of the team defined using this flag, so you can use this flag to ensure that the data from this app only includes data for users in that team. For the ``contribSummary`` sub-command, specifically, you use this flag to define the team that you wish to use for comparison when calculating the summary statistics for each user. If this flag isn't included on the command line, then the app uses the value of the `default_team` defined in the configuration file instead.

##### The `-o, --org-list` flag

You can use this flag to provide a comma-separated list of GitHub organizations (by name, not ID) that the user wishes to query for contributions. The app looks for contributions to any repositories in the list of organizations provided using this flag, and the output generated includes all contributions (by type based on the sub-command used) made to any repository in this list of organizations. If this flag isn't used to specify the list of organizations that the user wants to query, then the app uses the `orgs` parameter from the configuration file to set the default list of organizations to query for.

##### The `-u, --user-list` flag

You can use this flag to provide a comma-separated list of users (by name, not GitHub ID) that the user wishes to query for contributions from. For this to work correctly, the associated configuration file must already contain a mapping between the user names passed in using this flag and their associated GitHub user IDs. If a user list isn't specified, either using this flag or the alternate `-i, --github-id-list` flag (see below for more details), then the app uses the list of users in the team defined on the command-line (or the default team if a team wasn't specified) to construct the list of users to query for. As an aside, if you use this flag to pass in a list of users by name, then those users must be a part of the team used for comparison; if a named user isn't a member of that team, then the app skips that user (and with an associated warning printed to the standard error stream for the app noting that fact).

##### The `-i, --github-id-list` flag

You can use this flag to provide a comma-separated list of users (by GitHub ID) that the user wishes to query for contributions from. The app provides this flag as an alternative to using the `-u, --user-list` flag to accomplish this same task, so if you include both of these flags on the command-line then the app exits with an error. If a user list isn't specified, either using this flag or the alternate `-u, --user-list` flag (see below for more details), then the app uses the list of users in the team defined on the command-line (or the default team if a team wasn't specified) to construct the list of users to query for. Note that there is no check to ensure that users passed in via GitHub ID values using this flag are actually members of the underlying team, so if you are looking for information about contributions from non-team members to repositories in the named organizations, this is the flag that you should use to make that query.

##### The `-c, --config` flag

You can use this flag to specify the configuration file used to obtain things like the default team name, default list of organizations to query for, the list of team names, and the mappings of those team names to team members. By default the app uses either the  `~/.config/getGhInfo.yaml` file (if that file exists) or the `config.yml` file included at the top-level of this repository (if it doesn't), but some users might find it more useful to create their own configuration file outside of this repository (rather than modifying the default file included in the repository), and this flag is one way that the user can do so (and indicate to the app that they want to use their own configuration file instead of the default). Note that the default configuration file in this repository is easily overridden simply by creating an alternate `~/.config/getGhInfo.yaml` file containing their own definitions for the default team name, list of organizations, team names, and mapping of team names to user names and GitHub ID values. If the file passed in using this flag exists and is readable by the user, then it's used instead of either the  `~/.config/getGhInfo.yaml` file or the default file that's defined in this repository. If it doesn't exist or it's not readable, then the app exits with an error.

##### The `-f, --file` flag

You can use this flag to direct the output of the app to the named file. If this flag isn't used, then the app directs its output to its standard output stream by default. Error messages are all directed to standard error, so it should be relatively easy to separate any errors or warnings that occur from the results of the queries run, even when this flag isn't set.

In addition to these flags, there are also a set of three flags shown in the preceding help output that you can use to define the time window over which you would like to look for contributions to repositories in the defined GitHub organizations by any of the defined set of users.

### Obtaining information about GitHub repositories

The `repo` command retrieves information (lists and summary statistics) about the issues and pull requests in the repositories in a defined GitHub organization (or a defined set of GitHub organizations). The help output for that command shows the sorts of options that are available to you as a user:

```bash
The subcommand used as the root for all queries for repository-related data

Usage:
  getGhInfo repo [command]

Available Commands:
  issues      Gather issue-related data
  match       Show list of repositories that match the search criteria
  pulls       Gather PR-related data

Flags:
  -h, --help   help for repo

Global Flags:
  -c, --config string     configuration file to use
  -f, --file string       file/stream for output (defaults to stdout)
  -o, --org-list string   list of orgs to gather information from

Use "getGhInfo repo [command] --help" for more information about a command.
```

As you can clearly see, there are three sub-commands that supported by the `repo` command and this list provides a brief description of the output of each of those sub-commands:

* **match** - you can use this sub-command to generate a list of all of the repositories in the named GitHub organization (or list of GitHub organizations) that match a given pattern. By default the app assumes that the pattern passed in as a regular expression and it searches for repositories with names that match that pattern in the named organizations.
* **issues** - you can use this sub-command to gather statistics, report counts, or return lists of issues associated with the repositories that are "owned" by a given team (from the teams defined in the configuration file used with this app). There are a number of different sub-commands, and each of those sub-commands reports back different information about the issues associated with those repositories. Since these sub-commands are common between the `pulls` sub-command (described below) and this sub-command, a separate section of this document describes each of these sub-commands (below).
* **pulls** - this sub-command is use to gather statistics, report counts, or return lists of pull requests associated with the repositories that are "owned" by a given team (from the teams defined in the configuration file used with this app). There are a number of different sub-commands, and each of those sub-commands reports back different information about the issues associated with those repositories. Since these sub-commands are common between the `issues` sub-command (described previously) and this sub-command, a separate section of this document describes each of these sub-commands (below).

#### Flags used to control output

The  `user` sub-commands shown previously all support a common set of flags that's used on the command line to control the queries made against the GitHub GraphQL interface and, as a result, the output from this app.

##### The `-o, --org-list` flag

You can use this flag to provide a comma-separated list of GitHub organizations (by name, not ID) that the user wishes to query for contributions. The app looks for contributions to any repositories in the list of organizations provided using this flag, and the output generated includes all contributions (by type based on the sub-command used) made to any repository in this list of organizations. If this flag isn't used to specify the list of organizations that the user wants to query, then the app uses the `orgs` parameter from the associated configuration file to set the default list of organizations to query for.

##### The `-c, --config` flag

You can use this flag to specify the configuration file used to obtain things like the default team name, default list of organizations to query for, the list of team names, and the mappings of those team names to team members. By default the app uses either the e `~/.config/getGhInfo.yaml` file (if that file exists) or the `config.yml` file included at the top-level of this repository (if it doesn't), but some users might find it more useful to create their own configuration file outside of this repository (rather than modifying the default file included in the repository), and this flag is one way that the user can do so (and indicate to the app that they want to use their own configuration file instead of the default). Note that the default configuration file in this repository is easily overridden simply by creating an alternate `~/.config/getGhInfo.yaml` file containing their own definitions for the default team name, list of organizations, team names, and mapping of team names to user names and GitHub ID values. As mentioned previously, if this file exists then it's used instead of the default file that's defined in this repository.

##### The `-f, --file` flag

You can use this flag to direct the output of the app to the named file. If this flag isn't used, then the app directs its output to its standard output stream by default. Error messages are all directed to standard error, so it should be relatively easy to separate any errors or warnings that occur from the results of the queries run, even when this flag isn't set.

In addition to these flags, there are also a set of three flags in the help output shown previously that you can use to define the time window over which you would like to look for contributions to repositories in the defined GitHub organizations by any of the defined set of users.

### The `match` repository sub-command

The first sub-command for the `repo` command is one that's used to gather information about the repositories that are available in the named list of GitHub organizations. Note that while the usage for this sub-command is currently quite simple, there might be plans in the future to extend it's capabilities to support additional features:
```bash
$ go run . repo match --help
Constructs a list of all of the repositories in the named (set of) GitHub
organization(s) that have a name matching the define search pattern
passed in by the user.

Usage:
  getGhInfo repo match [flags]

Flags:
  -g, --glob-style-pattern       interpret the pattern as a glob-style pattern
  -h, --help                     help for match
  -i, --include-archived-repos   include archived repositories in output
  -p, --search-pattern string    pattern to match against repository names

Global Flags:
  -c, --config string           configuration file to use
  -e, --exclude-private-repos   exclude private repositories from output
  -f, --file string             file/stream for output (defaults to stdout)
  -o, --org-list string         list of orgs to gather information from
```

As you can see, there are a few flags that's set to control the output of this command, specifically:

##### The `-e, --exclude-private-repos` flag

You can use this flag to exclude private repositories from the list of matching repositories that's returned by this sub-command. By default, this flag isn't set, so unless it's set the output of this sub-command includes data from both public and private repositories.

##### The `--include-archived-repos` flag

You can use this flag to include information about archived repositories from the output of this sub-command. By default, this flag isn't set, so unless it's set information the output of this sub-command won't include data from any archived repositories.

##### The `-p, --search-pattern` flag

You can use this flag to define the search pattern (as a regular expression) that the repository name must match to include information from that repository in the output of this sub-command. If you don't pass in a pattern to match using this flag, this sub-command returns information for **all** repositories in the specified GitHub organization (or organizations). Otherwise, if you do set the pattern using this flag, the sub-command returns information **only** for repositories whose names match the pattern.

##### The `-g, --glob-style-pattern` flag

If you include this flag on the command line, the app attempts to match the repositories found against the previously described search pattern, interpreting that search pattern as a glob-syle pattern instead of the default behavior of interpreting that search pattern as a regular expression. In some situations, this might make defining a pattern to match against simpler.

### The `issues` and `pulls` repository sub-commands

These two `repo` sub-commands share a common structure in terms of the sub-commands that they each support. The only real difference is in the name of the sub-command itself (`issues` vs. `pulls`) and the type of data that each return as a result. Obviously, the first returns data and statistics related to the issues associated with repositories in the named GitHub organization (or organizations), while the second returns the same sorts of information but for pull requests. While this document only shows the help output for the `issues` sub-command here, you can rest assured that the flags and sub-commands available for this sub-command exactly match those available in the `pulls` sub-command. With that preamble in mind, here's the help output for the  `issues` sub-command:

```bash
The subcommand used as the root for all queries for issue-related data

Usage:
  getGhInfo repo issues [command]

Available Commands:
  age               Statistics for the 'age' of open isues
  countClosed       Count of closed issues in the named GitHub organization(s)
  countOpen         Count of open issues in the named GitHub organization(s)
  firstResponseTime Statistics for the 'time to first response' of open isues
  listClosed        List the closed issues in the named GitHub organization(s)
  listOpen          List the open issues in the named GitHub organization(s)
  listUnassigned    List the unassigned and open issues in the named GitHub organization(s)
  staleness         Statistics for the time since the last response for open PRs
  timeToResolution  Statistics for the 'time to resolution' of closed isues

Flags:
  -w, --complete-weeks             only output complete weeks (starting Monday)
  -h, --help                       help for issues
  -l, --lookback-time string       'lookback' time window (eg. 10d, 3w, 2m, 1q, 1y)
  -d, --ref-date string            reference date for time window (YYYY-MM-DD)
  -m, --repo-mapping-file string   name of the repository mapping file to use
  -t, --team string                name of team to restrict repository list to

Global Flags:
  -c, --config string           configuration file to use
  -e, --exclude-private-repos   exclude private repositories from output
  -f, --file string             file/stream for output (defaults to stdout)
  -o, --org-list string         list of orgs to gather information from

Use "getGhInfo repo issues [command] --help" for more information about a command.
```

As you can clearly see, there are a number of sub-commands defined for the `issues` (or `pulls`) sub-command, and each returns a different type of data:

* **The `age` sub-command**: this sub-command returns the statistics related to the age of all of the open issues (or pull requests in the case of the `pulls` sub-command) that existed during the defined time window for all repositories in the named GitHub organization (or organizations) including the minimum, maximum, average, median, first quartile, and third quartile values of the ages of those open issues. In addition to the summary statistics, this sub-command also returns number of open issues in the declared time window, the start and end times of that time window, and a title to make interpretation of the summary statistics easier for the user.

* **The `countOpen` sub-command**: this sub-command returns the number of issues (or pull requests in the case of the `pulls` sub-command) that were open during the defined time window for all repositories in each of the named GitHub organizations along with the total number of issues that were open in this time frame for all repositories in all organizations, the start and end times of the time window used when searching for those issues, and a title (to make it easier for the user to interpret these values).

* **The `countClosed` sub-command**: this sub-command returns the number of issues closed (or pull requests in the case of the `pulls` sub-command) during the defined time window for all repositories in each of the named GitHub organizations along with the total number of issues closed during this time frame for all repositories in all of the named organizations, the start and end times of the time window used when searching for those issues, and a title (to make it easier for the user to interpret these values).

* **The `firstResponseTime` sub-command**: this sub-command returns the statistics related to the "time to first response" for the issues (or pull requests in the case of the `pulls` sub-command) that were open during the defined time window for all repositories in the named GitHub organization (or organizations) including the minimum, maximum, average, median, first quartile, and third quartile values for those "time to first response" values (along with the total number of issues open during this time frame for all repositories in all of the named organizations, the start and end times of the time window used when searching for those issues, and a title to make it easier for the user to interpret these values). The "time to first response" metric tracks how long it took the team to respond to an issue after it was first opened.

* **The `staleness` sub-command**: this sub-command returns the statistics related to the "time since last response" (or "staleness") for the issues (or pull requests in the case of the `pulls` sub-command) that were open during the defined time window for all repositories in the named GitHub organization (or organizations) including the minimum, maximum, average, median, first quartile, and third quartile values for those "staleness" values (along with the total number of issues open during this time frame for all repositories in all of the named organizations, the start and end times of the time window used when searching for those issues, and a title to make it easier for the user to interpret these values). This "staleness" metric tracks how it has been since a team member last responded to a (still) open issue.

* **The `timeToResolution` sub-command**: this sub-command returns the statistics related to the "time to resolution" for the issues closed (or pull requests in the case of the `pulls` sub-command) during the defined time window for all repositories in the named GitHub organization (or organizations) including the minimum, maximum, average, median, first quartile, and third quartile values for those "time to resolution" values (along with the total number of issues open during this time frame for all repositories in all of the named organizations, the start and end times of the time window used when searching for those issues, and a title to make it easier for the user to interpret these values). The "time to resolution" metric tracks the time that it took to resolve (or close) an issue, and is simply the difference between an issue's creation and resolution (or closure) time.

* **The `listOpen` sub-command**: this sub-command returns a list of the issues (or pull requests in the case of the `pulls` sub-command) that were open at some point during the defined time window for all repositories in the named GitHub organization (or organizations) sorted (from greatest to least) by the age of each open issue. The output includes

  * the URL for the issue
  * the issue's title, creation time, and age
  * a flag indicating if the issues is currently open or closed, along with the issue's resolution (or closure) time
  * the creator of the issue (by GitHub ID) along with some associated meta-data from their GitHub profile (the company that they work at and their email) if their profile includes that information
  * a comma-separated list of assignees for that issue

  With this information, the user should be able to filter out the issues that they're interested in using external tools (like `jq`)

* **The `listClosed` sub-command**: this sub-command returns a list of the issues closed (or pull requests in the case of the `pulls` sub-command) during the defined time window for all repositories in the named GitHub organization (or organizations) sorted (from greatest to least) by the age of each open issue. The output includes

  * the URL for the issue
  * the issue's title, creation time, and age
  * a flag indicating if the issues is currently open or closed, along with the issue's resolution (or closure) time
  * the creator of the issue (by GitHub ID) along with some associated meta-data from their GitHub profile (the company that they work at and their email) if their profile includes that information
  * a comma-separated list of assignees for that issue

  With this information, the user should be able to filter out the issues that they're interested in using external tools (like `jq`)

* **The `listUnassigned` sub-command**: this sub-command returns a list of the issues (or pull requests in the case of the `pulls` sub-command) that were open at some point during the defined time window for all repositories in the named GitHub organization (or organizations) and that didn't have anyone assigned to work on them. As is the case with the previously described `listOpen` sub-command, the app sorts the output list (from greatest to least) by the age of each open issue (and the meta-data returned is identical to that returned by the `listOpen` sub-command)

All of these sub-commands support the same set of command-line flags, which are mainly focused on defining a time window for the issues (or pull requests) that you're interested in (see the next section for more detail on those command-line flags and how they're used to specify that time window), but there are two flags used for both of these sub-commands that deserve a bit more discussion, the `-t, --team` flag and the `-m, --repo-mapping-file` flag.

##### The `-t, --team` flag

You can use this flag to define the team that owns the list of repositories that you want to gather information about. The app uses this team name to determine which repositories to gather information for based on a list of repositories pulled in from a repository mapping file (see the next section for details) and the members of the team that "owns" those repositories (based on the users defined to be a part of that team in the configuration file embedded in this repository). As such, this flag is quite useful for restricting the list of repositories that you would like to calculate statistics for (or gather information from). If this flag isn't specified, the app uses the default team defined in the associated configuration file as the `team` for all of these sub-commands.

##### The `-m, --repo-mapping-file` flag

You can use this flag to specify the repository mapping file that's used to map teams to lists of repositories (see the next section of this document for more information on the structure of that file and how it's used). By default, the app uses the file specified in the `default_repo_mapping` key in the associated configuration file if this flag isn't used to override that value, but if that value (currently set to the string `../cpe-datasets/repositories.yml`) doesn't match the location of your repository mapping file, you must use this flag to point to wherever you saved your own repository mapping file locally.

### Mapping teams to repositories

As mentioned previously, the app uses a defined team name, in combination with the teams defined in the associated configuration file and a "repository mapping" file that maps the repositories listed in that file to the teams that manage them. With that mapping in place, the app constructs a list of repositories to gather data for based on the values you passed in for these arguments on the command-line. In order for this process this to work correctly, the teams defined in the configuration file for this app need to match the teams defined in the repository mapping file (by name), and the repository mapping file needs to look something like this:

```yaml
- group: cpe
  repositories:
    - url: https://github.com/CircleCI-Public/Sample-Swift-CFD
      tags: [ "sample-project" ]
    - url: https://github.com/CircleCI-Public/Sample-Flutter-CFD
      tags: [ "sample-project" ]
      ...
  children:
    - group: images
      repositories:
        - url: https://github.com/CircleCI-Public/cimg-android
          tags: [ "image" ]
        - url: https://github.com/CircleCI-Public/cimg-aws
          tags: [ "image" ]
      ...
```

Note that the structure of this repository mapping file is an array of dictionary values containing a `group` key that points to a name for the group (this name must match one of the team names in the configuration file), a `repository` key which points to a list of dictionary entries where each of those entries contains a `url` and a list of `tags`, and an optional `children` key that points to another list containing one or more additional groups that follow this same structure. If one group contains another group in this file, then it's assumed that while the nested groups are a group in its own right, they're also a subgroup of the higher-level group that contains them (and the app assumes that the higher-level group "owns" all of the repositories that are "owned" by any of it's subgroups).

This nested structure for defining the repository mapping file is critical for being able to define complex team structures where some parts of the team are responsible for some repositories and other parts of the group for others, but the combined group is responsible for a third group of repositories. Rest assured that the `repo` sub-commands that utilize this repository mapping file to map teams to lists of repositories managed by those teams are quite adept at putting together the proper list of repositories to collect data from based on the team that the user has defined on the command line using the `-t, --team` flag.

### Defining the time windows for queries

As mentioned in the preceding discussions of the `user` and `repo` commands and sub-commands, there are several flags used by these sub-commands to define the time window for the underlying queries used to gather the data that's of interest to the user. The three flags used for this purpose (which you have, no doubt seen in many of the usage examples shown previously in this document) are as follows:

* **the `-d, --ref-date` flag**: the string value passed into the program using this flag must be of the format `YYYY-MM-DY` and specifies the "reference date" to use when constructing the time window for the query. If the string value passed in using this flag that doesn't correspond to that format (4 digit year, two digit month, and two digit day separated by dashes), then the app exits with an error
* **the `-l, --lookback-time` flag**: the string value passed into the program using this flag must be of the form of an integer number followed by a single letter suffix, and the combination of that integer and suffix that indicates the time period to look **back** from the reference date when defining the time window to query.
* **the `-w, --complete-weeks` flag**: if the user sets this flag, then the app adjusts the time window so that its output only includes data from complete weeks (where a complete week starts at midnight on a Monday and continuing through until midnight the following Sunday night).

It's the combination of the values passed in for these flags that determines the time window that's used when querying for contributions from the users on the team or for issues or pull requests from the underlying repositories. That said, there are a few important things to remember:

* Since the user defines the reference time using a simple date-time string, all the date-time windows for the underlying queries start and/or end at zero hours, UTC (or Coordinated Universal Time); there is no option to shift the starting date-times by anything less than a day using the defined look-back time argument
* If the reference date that's passed in exceeds the current date (shifted by the look-back time, if any, that the user passed in, and assuming that the look-back time is negative), the app prints a warning as part of its output and shows data up to the current date.
* The look-back time that's passed in must be a regular expression of the form `^[+-]?[0-9]+[dwmqy]$`. To translate this into plain English, the look-back time consists of an integer value (with an optional plus or minus that's used to indicate a positive or negative look-back time) followed by a single letter suffix that represents the time units for the look-back time value: `d` for days, `w` for weeks, `m` for months (defined here as a 30 day period), `q` for quarters (defined here as three quarters, or a 90 day period), or `y` for years. For example, you would pass in a look-back time of `12w` if you wanted the start of the time window for the queries to be 12 weeks, or 84 days prior to the reference date that you passed in.
* As mentioned previously, the look-back time passed in can be a positive or negative number. If the look-back time you pass in is negative, you are actually instructing the system to look **ahead** by the corresponding number of days, weeks, etc. from the input reference date. If the number passed in is negative, and the reference date is less than that amount of time back from the current date it isn't an error, but the resulting data is "truncated" and the app prints a warning to its standard error stream denoting this fact.
* Due to limitations in the GitHub GraphQL API, this app can only retrieve only a year's worth of data in a single query. As such, the look-back time passed into the app can't exceed one year (or 365 days) without the app throwing an error. Similarly, if you set the reference date to more than one year ago and fail to include a look-back time as an argument to this app, then the app exits with an error (see below for details on why this is the case). Future changes to this app might detect this scenario and handle it appropriately (by breaking such a query up into two or more queries where none of them exceed a year in length), but that's currently not the case. This doesn't mean that data can't be retrieve from more than a year in the past, just that the total time window requested in a single query (based on the reference time and look-back time defined) can't exceed one year in length.

#### Default behavior: when the user provides only some (or none) of these flags

In addition to whether the value passed in using the look-back flag is a positive or negative, how the time window for the queries maps to the reference time passed into the app (if the user passed a value in) also depends on which of these flags you include as part of the command-line arguments to this app:

* As mentioned previously, if the user sets the `-w, --complete-weeks` flag, then the app shifts the time window (and, perhaps, truncates its length) so that it only returns data from complete weeks. The app accomplishes this by shifting the defined reference time (or current date if a reference time wasn't specified) to either the previous Monday (if the look-back time is positive() or the next Monday (if the look-back time is either negative or undefined), then calculating a time window using the look-back time value and finally adjusting the "other boundary" of the time window, truncating it so that the output of this app only includes data from "complete weeks" within the requested time window.
* If the user didn't include a reference date on the command-line, then the app uses the current date as the end date for the time window.
  * In this situation, if a look-back time is also not included, then the app uses a default look-back time of 90 days to construct the time window for the query (a time window that covers the 90 days prior to the current date)
  * Since the look-back time is positive in this situation, if the `-w, --complete-weeks` flag was also set in this situation the app sets the reference date to the previous Monday (closest to the current date), and adjusts the look-back time accordingly to ensure that the output only includes an even number of weeks (84 days or 12 weeks in this case). 
* If the user included a reference date on the command-line but a look-back time wasn't provided, then the app uses a time window that starts on the reference date ends on the current date. In this case, if the reference date chosen is in the future, then an error the app exits with an error.
* If the user provides both a reference date and a look-back time on the command line, then the app uses those values to define the time window based on the look-back time value, it's units, and whether the specified "look-back" time was positive or negative (where a negative value, as mentioned previously, indicates that the user wants to actually look **ahead** of the reference time by the stated look-back time value).

If you keep these basic rules in mind, it's easy to see that you can define pretty much any time window you would like using these the reference date and look-back time, making it possible to look for data only within a well-defined time window. The only limitation is that any time window that results from applying these rules can't be longer than one year in length; if the defined time window is longer than a year than the app exits with an error.
