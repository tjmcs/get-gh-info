# Tracking GitHub Contributions

This directory contains a Go program that can be used to gather either information regarding contributions made by an input list of team members to the repositories contained in an input list of GitHub organizations or information about the repositories themselves. The program gathers this data via GitHub's GraphQL API, then outputs the data to a named output file in JSON format (or to the standard output stream if an output file is not specified).

The input arguments for the program can be passed in on the command line using a set of defined command line flags or a configuration file can be used to define "default values" for these same command-line arguments (a file named `config.yml` is included as part of this repository). If the user wants to use a different configuration file, then the name of that file can also be passed using a command-line flag.

## Usage

This program that can be used either to obtain information about the contributions made by one or more users to the repositories in one or more GitHub organizations or it can be used to obtain information about the issues and pull requests associated with those same repositories. The information gathered is output in JSON format, either to the standard output stream (the default) or to a named file (if one is specified on the command-line). The easiest way to see what sorts of options are supported is to take advantage of the help support that is built into the application.  Here, for example, is the general (top-level) help available through the command-line interface (or CLI):

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

As you can easily see from that output, there are two main commands available in this application:

1. the `user` command, which can be used to gather information related to user contributions to the repositories in the specified GitHub organization (or organizations), and
2. the `repo` command, which can be used to gather information related to the issues and pull requests in those same repositories

The usage for both of these commands is outlined in the sections that follow.

### Obtaining information about user contributions

The `user` command can be used to obtain information about the contributions that a given user or set of users has made to the repositories in a defined GitHub organization (or a defined set of GitHub organizations). The help output for that command shows the sorts of options that are available to you as a user:

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

The subcommands that are supported for the `user` command and a brief description of the output for each are shown here:

* **contribSummary** - generates a summary of the contributions made by each user in the input list of users to repositories in the named GitHub organizations, including the number of pull requests, pull request reviews, number of repositories that they have contributed pull requests to, and number of repositories that they have contributed pull request reviews to. In addition, values are included in the summary that show (as a percentage) how the values for each user in the input user list compare with the average for all users in the input team.
* **contribs** - generates a list of the total number of commits made by all users in the defined list of users to repositories in the defined GitHub organizations, along with a detailed list (broken out by user) of the number of commits that each user in the defined list of users made to those same repositories (for historical reasons this data is shown for each repository by the date on which those commits were made, with separate entries for each date/repository combination).
* **contribsByType** - generates a list of the of the total number of pull requests and pull request reviews made by all users in the defined list of users to repositories in the defined GitHub organizations, along with a detailed list (broken out by user) of the details for the pull requests and pull reviews that each user in the list user made to those same repositories.
* **prList** - generates a list of the of the total number of pull requests made by all users in the defined list of users to repositories in the defined GitHub organizations, along with a detailed list (broken out by user) of the details for the pull requests that each user in the list user made to those same repositories.
* **prReviews** - generates a list of the of the total number of pull request reviews made by all users in the defined list of users to repositories in the defined GitHub organizations, along with a detailed list (broken out by user) of the details for the pull request reviews that each user in the list user made to those same repositories.

#### Flags used to control output

The  `user` subcommands shown above all support a common set of flags that can be used on the command line to control the queries made against the GitHub GraphQL interface and, as a result, the output from this application.

##### The `-t, --team` flag

This flag can be used to provide the name of the team that you are interested in retrieving data for or comparing user contributions against (depending on the subcommand). For most of the `user` subcommands this flag can be used as a replacement for the `-u, --user-list` or `-i, --github-id-list` flags (providing a shorthand method of asking for information about the contributions for all of the users in the named team). If this flag is used to provide a name for a team on the command line to these same subcommands, then any users in the list of users defined using either of those flags that are not members of the defined team will be skipped, so it can also be used to ensure that only data from users in that team are included in the output. For the `contribSummary` subcommand, specifically, this flag is used to define the team that you wish to use for comparison when calculating the summary statistics for each user.  If this flag is not included on the command line, then the value of the `default_team` defined in the configuration fill will be used instead.

##### The `-o, --org-list` flag

This flag can be used to provide a comma-separated list of GitHub organizations (by name, not ID) that the user wishes to query for contributions.  All repositories in the list of organizations provided using this flag will be queried for contributions, and the output generated will include all contributions (by type based on the subcommand used) made to any repository in this list of organizations.  If this flag is not used to specify the list of organizations that the user wants to query, then the `orgs` parameter from the configuration file associated with this application will be used to set the default list of organizations that should be queried.

##### The `-u, --user-list` flag

This flag can be used to provide a comma-separated list of users (by name, not GitHub ID) that the user wishes to query for contributions from. The mapping between the user names passed in using this flag and their associated GitHub user IDs must be defined in the associated configuration file If this flag is used to specify the list of users to query for contributions from.  If a user list is not specified, either using this flag or the alternate `-i, --github-id-list` flag (see below for more details), then the list of users in the team defined on the command-line (or the default team if a team was not specified) will be used to construct the list of users to query for. It should also be noted that if users are passed in by name using this flag, those users must be a part of the team being used for comparison; if a named user cannot be found on that team, then that user will be skipped (and the fact that the user in question is being skipped will be noted with a warning printed out to the standard error stream for this application as part of the program run).

##### The `-i, --github-id-list` flag

This flag can be used to provide a comma-separated list of users (by GitHub ID) that the user wishes to query for contributions from. This flag is offered as an alternative to using the `-u, --user-list` flag to accomplish this same task, so if both of these flags are defined on the command-line then an error will be thrown.   If a user list is not specified, either using this flag or the alternate `-u, --user-list` flag (see below for more details), then the list of users in the team defined on the command-line (or the default team if a team was not specified) will be used to construct the list of users to query for. It should also be noted here that there is no check to ensure that users passed in via GitHub ID values using this flag are actually members of the underlying team, so if you are looking for information about contributions from non-team members to repositories in the named organizations, this is the flag that you should use to make that query.

##### The `-c, --config` flag

This flag can be used to specify the configuration file that should be used to obtain things like the default team name, default list of organizations to query for, the list of team names, and the mappings of those team names to team members. By default the `config.yml` file included at the top-level of this repository is used, but some users might find it more useful to create their own configuration file outside of this repository (rather than modifying the default file included in the repository), and this flag is one way that the user can do so (and indicate to the application that they want to use their configuration file instead of the default). It should also be noted that the default configuration file in this repository can be easily overridden simply be creating an alternate `~/.config/getGhInfo.yaml` file containing their own definitions for the default team name, list of organizations, team names, and mapping of team names to user names and GitHub ID values. If this file exists, it will be used instead of the default file that is defined in this repository.

##### The `-f, --file` flag

This flag can be used to direct the output of the application to the named file.  If this flag is not used, then output will be directed to standard output by default.  Error messages are all directed to standard error, so it should be relatively easy to separate any errors or warnings that occur from the results of the queries run, even when this flag has not been set.

In addition to these flags, there are also a set of three flags in the help output shown above that are used to define the time window over which you would look to look for contributions to repositories in the defined GitHub organizations from the defined set of users.

### Obtaining information about GitHub repositories

The `repo` command can be used to obtain information about the repositories in a defined GitHub organization (or a defined set of GitHub organizations). The help output for that command shows the sorts of options that are available to you as a user:

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

As you can clearly see, there are three subcommands that are supported for the `repo` command and a brief description of the output for each are shown here:

* **match** - this subcommand can be used to generate a list of all of the repositories in the named GitHub organization (or list of GitHub organizations) that match a given pattern. The pattern that is passed in as an argument is assumed to be a regular expression, and that regular expression is used to search for repositories with a name that matches that pattern.
* **issues** - this subcommand can be use to gather statistics, report counts, or return lists of issues associated with the repositories that are "owned" by a given team (where the teams are defined in the configuration file used with this application). There are a number of different subcommands, and each of those subcommands reports back different information about the issues associated with those repositories. Since these subcommands are common between the `pulls` subcommand (which is described below) and this subcommand, we will describe those subcommands in a separate section of this document (below).
* **pulls** - this subcommand can be use to gather statistics, report counts, or return lists of pull requests associated with the repositories that are "owned" by a given team (where the teams are defined in the configuration file used with this application). There are a number of different subcommands, and each of those subcommands reports back different information about the issues associated with those repositories. Since these subcommands are common between the `issues` subcommand (which is described above) and this subcommand, we will describe those subcommands in a separate section of this document (below).

#### Flags used to control output

The  `user` subcommands shown above all support a common set of flags that can be used on the command line to control the queries made against the GitHub GraphQL interface and, as a result, the output from this application.

##### The `-o, --org-list` flag

This flag can be used to provide a comma-separated list of GitHub organizations (by name, not ID) that the user wishes to query for contributions.  All repositories in the list of organizations provided using this flag will be queried for contributions, and the output generated will include all contributions (by type based on the subcommand used) made to any repository in this list of organizations.  If this flag is not used to specify the list of organizations that the user wants to query, then the `orgs` parameter from the configuration file associated with this application will be used to set the default list of organizations that should be queried.

##### The `-c, --config` flag

This flag can be used to specify the configuration file that should be used to obtain things like the default team name, default list of organizations to query for, the list of team names, and the mappings of those team names to team members. By default the `config.yml` file included at the top-level of this repository is used, but some users might find it more useful to create their own configuration file outside of this repository (rather than modifying the default file included in the repository), and this flag is one way that the user can do so (and indicate to the application that they want to use their configuration file instead of the default). It should also be noted that the default configuration file in this repository can be easily overridden simply be creating an alternate `~/.config/getGhInfo.yaml` file containing their own definitions for the default team name, list of organizations, team names, and mapping of team names to user names and GitHub ID values. If this file exists, it will be used instead of the default file that is defined in this repository.

##### The `-f, --file` flag

This flag can be used to direct the output of the application to the named file.  If this flag is not used, then output will be directed to standard output by default.  Error messages are all directed to standard error, so it should be relatively easy to separate any errors or warnings that occur from the results of the queries run, even when this flag has not been set.

In addition to these flags, there are also a set of three flags in the help output shown above that are used to define the time window over which you would look to look for contributions to repositories in the defined GitHub organizations from the defined set of users.

### The `match` repository subcommand

The first subcommand for the `repo` command is one that can be used to gather information about the repositories that are available in the named list of GitHub organizations.  The usage for this subcommand is quite simple, although we might extend this subcommand in the future and add more capabilities:
```bash
$ go run getGhInfo.go repo match --help
Constructs a list of all of the repositories in the named (set of) GitHub
organization(s) that have a name matching the define search pattern
passed in by the user.

Usage:
  getGhInfo repo match [flags]

Flags:
  -e, --exclude-private-repos    exclude private repositories from output
  -h, --help                     help for match
  -i, --include-archived-repos   include archived repositories in output
  -p, --search-pattern string    pattern to match against repository names

Global Flags:
  -c, --config string     configuration file to use
  -f, --file string       file/stream for output (defaults to stdout)
  -o, --org-list string   list of orgs to gather information from
```

As you can see, there are a few flags that can be set to control the output of this command, specifically:

##### The `-e, --exclude-private-repos` flag

This flag can be used to exclude private repositories from the list of matching repositories that is returned by this subcommand. By default, this flag is set to `false`, so unless it is set both public and private repositories will be included in the output.

##### The `--include-archived-repos` flag

This flag can be used include information about archived repositories from the output of this subcommand. By default, this flag is set to `false`, so unless it is set information about repositories that have been archived will be filtered out of this subcommand's output.

##### The `-p, --search-pattern` flag

This flag can be used to define the search pattern (as a regular expression) that the repository name must match if the information from that repository is to be included in the output of this subcommand. By default, if a value for the pattern that must be matched is not passed in using this flag, information will be returned for **all** repositories in the named GitHub organization (or organizations), but if it is set then information will be returned **only** for repositories who's names match the pattern passed in using this flag.

### The `issues` and `pulls` repository subcommands

These two `repo` subcommands share a common structure in terms of the subcommands that they each support.  The only real difference is in the name of the subcommand itself (`issues` vs. `pulls`) and the type of data that each return as a result. Obviously, the first returns data and statistics related to the issues associated with repositories in the named GitHub organization (or organizations), while the second returns the same sorts of information but for pull requests.  As such, while we are showing the help output for the `issues` subcommand here the flags that can be set and the subcommands that are available match those that can be set and that are available for the `pulls` subcommand completely. With that preamble in mind, here's the help output for the  `issues` subcommand:

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

As you can clearly see, there are a number of subcommands defined for the `issues` (or `pulls`) subcommand, and each returns a different type of data:

* **The `age` subcommand**: this subcommand returns the statistics related to the age of all of the open issues (or pull requests in the case of the `pulls` subcommand) that existed during the defined time window for all repositories in the named GitHub organization (or organizations) including the minimum, maximum, average, median, first quartile, and third quartile values of the ages of those open issues. In addition, the number of open issues in this time frame that were used to determine these values and the start and end times of the time window are returned along with a title in order to make it easier for the user to interpret these values.

* **The `countOpen` subcommand**: this subcommand returns the number of issues (or pull requests in the case of the `pulls` subcommand) that were open during the defined time window for all repositories in each of the named GitHub organizations along with the total number of issues that were open in this time frame for all repositories in all organizations, the start and end times of the time window used when searching for those issues, and a title (in order to make it easier for the user to interpret these values).

* **The `countClosed` subcommand**: this subcommand returns the number of issues (or pull requests in the case of the `pulls` subcommand) that were closed during the defined time window for all repositories in each of the named GitHub organizations along with the total number of issues that were closed during this time frame for all repositories in all of the named organizations, the start and end times of the time window used when searching for those issues, and a title (in order to make it easier for the user to interpret these values).

* **The `firstResponseTime` subcommand**: this subcommand returns the statistics related to the "time to first response" for the issues (or pull requests in the case of the `pulls` subcommand) that were open during the defined time window for all repositories in the named GitHub organization (or organizations) including the minimum, maximum, average, median, first quartile, and third quartile values for those "time to first response" values. In addition, the total number of open issues that were found in this time frame for all repositories, the start and end times of the time window used when searching for those issues, and a title are returned in order to make it easier for the user to interpret these values. We will discuss our definition of these "time to first response" values a bit more, below, but this metric is intended to be used to determine how long it took the team to respond to an issue after it was first opened.

* **The `staleness` subcommand**: this subcommand returns the statistics related to the "time since last response" (or "staleness") for the issues (or pull requests in the case of the `pulls` subcommand) that were open during the defined time window for all repositories in the named GitHub organization (or organizations) including the minimum, maximum, average, median, first quartile, and third quartile values for those "staleness" values. In addition, the total number of open issues that were found in this time frame for all repositories, the start and end times of the time window used when searching for those issues, and a title are returned in order to make it easier for the user to interpret these values. We will discuss our definition of these "staleness" values a bit more, below, but this metric is intended to be used to determine how it has been since a team member last responded to an open issue.

* **The `timeToResolution` subcommand**: this subcommand returns the statistics related to the "time to resolution" for the issues (or pull requests in the case of the `pulls` subcommand) that were closed during the defined time window for all repositories in the named GitHub organization (or organizations) including the minimum, maximum, average, median, first quartile, and third quartile values for those "time to resolution" values. In addition, the total number of issues that were closed during this time frame for all repositories, the start and end times of the time window used when searching for those issues, and a title are returned in order to make it easier for the user to interpret these values. The "time to resolution" is defined, quite simply, as the time that it took for an issue to be resolved (or closed) once it was created, and is simply the difference between when the issue was created and when it was closed.

* **The `listOpen` subcommand**: this subcommand returns a list of the issues (or pull requests in the case of the `pulls` subcommand) that were open at some point during the defined time window for all repositories in the named GitHub organization (or organizations) sorted (from greatest to least) by the age of each open issue. The output includes

  * the URL for the issue
  * the issue's title, date-time when it was created, and age
  * a flag indicating if the issues is currently open or closed, along with the date-time when it was closed (if it is currently closed)
  * the creator of the issue (by GitHub ID) along with some associated meta-data (the company that they work at and their email) if that information is included in their GitHub profile
  * a comma-separated list of assignees for that issue

  With this information, the user should be able to filter out the issues that they are interested in using external tools (like `jq`)

* **The `listClosed` subcommand**: this subcommand returns a list of the issues (or pull requests in the case of the `pulls` subcommand) that were closed during the defined time window for all repositories in the named GitHub organization (or organizations) sorted (from greatest to least) by the age of each open issue. The output includes

  * the URL for the issue
  * the issue's title, date-time when it was created, and age
  * a flag indicating if the issues is currently open or closed, along with the date-time when it was closed
  * the creator of the issue (by GitHub ID) along with some associated meta-data (the company that they work at and their email) if that information is included in their GitHub profile
  * a comma-separated list of assignees for that issue

  With this information, the user should be able to filter out the issues that they are interested in using external tools (like `jq`)

* **The `listUnassigned` subcommand**: this subcommand returns a list of the issues (or pull requests in the case of the `pulls` subcommand) that were open at some point during the defined time window for all repositories in the named GitHub organization (or organizations) and that did not have anyone assigned to work on them. As is the case with the `listOpen` subcommand (above), the output is sorted (from greatest to least) by the age of each open issue (and the meta-data returned is identical to that returned by the `listOpen` subcommand)

All of these subcommands support the same set of command-line flags, which are mainly focused on defining a time window for the issues (or pull requests) that we are interested in (see the next section for more detail on those command-line flags and how they can be used to specify that time window), but there are two flags used for both of these subcommands that deserve a bit more discussion, the `-t, --team` flag and the `-m, --repo-mapping-file` flag.

##### The `-t, --team` flag

This flag can be used to define the team that owns the list of repositories that we want to gather information about. This team name is used to determine the list of repositories that we are interested in gathering information for (based on a list of repositories pulled in from a repository mapping file, see the next section for details) and the members of the team in that "owns" those repositories (based on the users defined to be a part of that team in the configuration file embedded in this repository). As such, this flag can be quite useful for restricting the list of repositories that we would like to calculate statistics for (or gather information from). If this flag is not specified, the default team defined in the configuration file is used as the `team` for all of these subcommands.

##### The `-m, --repo-mapping-file` flag

This flag can be used to specify the repository mapping file that will be used to map teams to lists of repositories (see the next section of this document for more information on how that file should be structured and how it's used). By default, the file specified in the `default_repo_mapping` key in the configuration file will be used if this flag is not used to override that value, but if that value (currently set to the string `../cpe-datasets/repositories.yml`) does not match the location of your repository mapping file, this flag can be used to point to wherever you have saved your repository mapping file locally.

### Mapping teams to repositories

As was mentioned previously, the value specified for the team using this flag is used, in combination with the teams defined in the configuration file for this application and a "repository mapping" file that maps a list of repositories to the teams that manage those repositories, to construct a list of repositories that we are interested in gathering data for. In order for this process this to work correctly, the teams defined in the configuration file for this application need to match the teams defined in the repository mapping file (by name), and the repository mapping file needs to look something like this:

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

Note that the structure of this repository mapping file is an array of dictionary values containing a `group` key that points to a name for the group (this name must match one of the team names in our configuration file), a `repository` key which points to a list of dictionary entries where each of those entries contains a `url` and a list of `tags`, and an optional `children` key that points to another list containing one or more additional groups that follow this same structure. If groups are nested, then it is assumed that while each of those nested groups can be considered to be a group in its own right, that same group is a subgroup of the higher-level group that it contained within (so the higher-level group is assumed to "own" all of the repositories that are "owned" by each of it's subgroups).

This nested structure for defining the repository mapping file is critical for being able to define complex team structures where some parts of the team will be responsible for some repositories and other parts of the group for others, but the combined group is responsible for a third group of repositories. Rest assured that the `repo` subcommands that utilize this repository mapping file to map teams to lists of repositories managed by those teams are quite adept at putting together the proper list of repositories to collect data from based on the team that the user has defined on the command line using the `-t, --team` flag.

### Defining the time windows for queries

As was mentioned in the discussions of the `user` and `repo` commands and subcommands (above), there are several flags that are used to define the time window to use for the underlying queries that are utilized by these commands and subcommands to gather the data that the user is interested in. The three flags used for this purpose (which you have, no doubt seen in many of the usage examples shown previously in this document) are as follows:

* **the `-d, --ref-date` flag**: the string value passed into the program using this flag is assumed to be of the format `YYYY-MM-DY` that specifies the "reference date" to use when constructing the time window for the query.  If a string value is passed in using this flag that doesn't correspond to that format (4 digit year, two digit month, and two digit day separated by dashes), then an error will be thrown by the program.
* **the `-l, --lookback-time` flag**: the string value passed into the program using this flag must be of the form of an integer number followed by a single letter suffix, and the combination of that integer and suffix that indicates the time period to look **back** from the reference date when defining the time window to query.
* **the `-w, --complete-weeks` flag**: if this flag is set, then the time window will be constructed in such a manner that only complete weeks are shown (where a complete week is defined as starting at midnight on a Monday and continuing through until midnight the following Sunday night)

It is the combination of the values passed in for these flags that determines the time window that will be used when querying for contributions from the users on the team or for issues or pull requests from the underlying repositories. That said, there are a few important things to remember:

* Since the reference time is specified using a simple date-time string, all queries are based on zero hours, UTC; there is no option to shift the starting date-times by anything less than a day using the defined lookback time argument
* If the reference date that is passed in exceeds the current date (shifted by the lookback time, if any, that was passed in if that lookback time is negative), a warning will be printed as part of the output of this application and the results will be truncated to only include data up to the current date.
* The lookback time that is passed in must be a regular expression of the form `^[+-]?[0-9]+[dwmqy]$`. To translate this into plain English, the lookback time consists of an integer value (with an optional plus or minus that can be used to indicate a positive or negative lookback time) followed by a single letter suffix that represents the time units for the lookback time value: `d` for days, `w` for weeks, `m` for months (where a month is defined as 30 days, `q` for quarters (where a quarter is defined as 90 days), or `y` for years.  For example, you would pass in in a lookback time of `12w` if you wanted the start of the time window for the queries to be 12 weeks, or or 84 days prior to the reference date that you passed in.
* As was mentioned previously, the lookback time passed in can be specified as a negative number.  In that case, you are actually instructing the system to look **ahead** by the corresponding number of days, weeks, etc. from the input reference date. If a negative number is passed in, and the reference date is less than that amount of time back from the current date it is not an error, but as we said previously the resulting data will be "truncated" (since we can't look into the future) and that fact will be noted in the standard error stream for this application.
* Due to limitations in the GitHub GraphQL API, only a year's worth of data can be returned in the results of a single query.  As such, the lookback time passed into the application cannot exceed one year (or 365 days) without an error being thrown by the application.  Similarly, if you set the reference date to more than one year ago and fail to include a lookback time as an argument to this application, the same error will be thrown (see below for details on why this is the case).  If there is a need to retrieve more than one year's worth of data in the future we can explore making changes to this application to detect this scenario and handle it appropriately (by breaking such a query up into two or more queries where none of them exceed a year in length). This doesn't mean that data can't be retrieve from more than a year in the past, just that the total time window requested in a single query (based on the reference time and lookback time defined) cannot exceed one year in length.

#### Default behavior (when some, or none of these flags are used)

In addition to whether the value passed in using the lookback flag is a positive or negative, how the time window for the queries we will be making maps to the reference time passed into the application (if one is passed in) also depends on which of these flags are included as part of the command-line arguments to this application:

* As was mentioned previously, if the `-w, --complete-weeks` flag is set, then the time window may be shifted and/or truncated in length to only return complete weeks. This is accomplished by shifting the defined reference time (or current date if a reference time was not specified) to either the previous Monday (if the lookback time is positive() or the next Monday (if the lookback time is either negative or undefined), then calculating a time window using the lookback time value and finally adjusting the "otehr boundary" of the time window, truncating it so that only data from "complete weeks" within the requested time window are included in the output of the application.
* If a reference date is not included on the command-line, then the end date for the time window to use when querying the system is assumed to be the current date.
  * In this situation, if a lookback time is also not included, then a default lookback time of 90 days is used to construct the time window to use (a time window that covers the 90 days prior to the current date)
  * Since the lookback time is positive in this situation, if the `-w, --complete-weeks` flag was also set in this situation the reference date would be set to the previous Monday closest to the current date, and the lookback time would be adjusted accordingly to ensure that only an even number of weeks (84 days or 12 weeks in this case) was included in the output. 
* If a reference date is included on the command-line but a lookback time is not included, then the time window used is assumed to be one that starts on the reference date ends on the current date. In this case, if the reference date chosen is in the future, then an error will be thrown.
* If both a reference date and a lookback time are provided on the command line, then those values are used to define the time window based on the lookback time value, it's units, and whether a positive or negative "lookback" time value was specified (where a negative value, as was mentioned previously, is used to indicate that the user wants to actually look **ahead** of the reference time by the stated lookback time value).

If you keep these basic rules in mind, we think it is easy to see that pretty much any time window can be defined using these the reference date and lookback time, making it possible to look for data only within a well-defined time window.  The only limitation is that any time window that results from applying these rules cannot be longer than one year in length or an error will be thrown.
