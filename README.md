# Tracking GitHub Contributions

This directory contains a Go program that can be used to gather information regarding contributions made by an input list of team members to the repositories contained in an input list of GitHub organizations. The program collects contribution data for each user via GitHub's GraphQL API, then outputs that data to the named output file in JSON format, with the output containing either the raw data for each team member (along with summary contribution data for all users in the list) or some basic contribution numbers for each user in the input list user (including some statistics comparing that user's contributions to the contributions of the team as a whole).

The input list of team members can be passed in as a comma-separated list of GitHub IDs on the command-line or via a field in a configuration file passed in (by filename) on the command-line. Similarly, the list of GitHub organizations that you want to track contributions to for each user can be passed in as a comma-separated list of organization names on the command-line or that list of organization names can be passed in via a field in that same configuraiton file. Details for the format(s) supported for that configuration file are sketched out, below.

## Usage

There are a number of commands supported by this program that can be used to obtain different types of contribution data for each user in the input list (or all users on the default team if an input list of users is not provided).  The easiest way to see what sorts of options are supported is to take advantage of the help support that is built into the application.  Here, for example, is the general (top-level) help available through the command-line interface (or CLI):

```bash
$ go run getGhInfo.go --help
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
  -f, --file string       file/stream to output data to (defaults to standard output)
  -h, --help              help for getGhInfo
  -o, --org-list string   list of orgs to gather information from

Use "getGhInfo [command] --help" for more information about a command.
```

As you can easily see from that output, this application can be used to gather repository-related data using the `repo` subcommand or it can be used to gather data related to user contributions using the `user` subcommand.  Since the `repo` subcommand is currently under development, we will focus most of the rest of this document on the `user` subcommand and the sorts of data that can be gathered about user contributions across the named organizations using that subcommand.

Taking a look at the help for the `user` subcommand, you can see that it supports a number of additional subcommands of its own:
```bash
$ go run getGhInfo.go user --help
The subcommand used as the root for all commands that make
user-related queries

Usage:
  getGhInfo user [command]

Available Commands:
  contribSummary Generates a summary (including statistics) of contributions
  contribs       Generates a list of commits and PRs made
  contribsByType Generates a list of PRs and PR reviews made
  prList         Generates a list the pull requests made
  prReviews      Generates a list the pull request reviews made

Flags:
  -i, --github-id-list string   list of GitHub IDs to gather contributions for
  -h, --help                    help for user
  -l, --lookback-time string    the 'lookback' time window (eg. 10d, 3w, 2m, 1q, 1y)
  -d, --ref-date string         reference date for time window (in YYYY-MM-DD format)
  -u, --user-list string        list of users to gather contributions for

Global Flags:
  -c, --config string     configuration file to use
  -f, --file string       file/stream to output data to (defaults to standard output)
  -o, --org-list string   list of orgs to gather information from

Use "getGhInfo user [command] --help" for more information about a command.
```

The subcommands that are supported for the `user` subcommand and a brief description of the output for each are shown here:

* **contribSummary** - generates a summary (in JSON format) of the contributions made by each user in the input `user-list` to repositories in the named  `org-list`, including the number of pull requests, pull request reviews, repositories that they have contributed pull requests to, and repositories in which they have contributed pull request reviews to. In addition, output values are generated that can be used to compare these values to the average of each of those values by all users in the input `team` (more details on this can be seen in the detailed description of this subcommand, below).
* **contribs** - generates a list of the total number of commits made by all users in input `user-list` to repositories in the named  `org-list` along with a detailed list by user of the number of commits that each user has made to those same repositories (for historical reasons this data is further broken out for each repository by the date on which those commits were made).
* **contribsByType** - generates a list of the of the total number of pull requests and pull request reviews made by all users in the input `user-list` to each of repositories in the named  `org-list` that they contributed to as a team, along with a detailed list (broken out by user) of the details for the pull requests and pull reviews that each user in the list user made to those same repositories.
* **prList** - generates a list of the of the total number of pull requests made by all users in the input `user-list` to each of repositories in the named  `org-list` that they contributed to as a team, along with a detailed list (broken out by user) of the details for the pull requests that each user in the list user made to those same repositories.
* **prReviews** - generates a list of the of the total number of pull request reviews made by all users in the input `user-list` to each of repositories in the named  `org-list` that they contributed to as a team, along with a detailed list (broken out by user) of the details for the pull request reviews that each user in the list user made to those same repositories.

Details for the flags supported by each of these subcommands are shown in the next section.

#### Flags used to control output

The four `user` subcommands shown above all support a common set of flags that can be used on the command line to control the queries made against the GitHub GraphQL interface and, as a result, the output from this application.

##### The `-o, --org-list` flag

This flag can be used to provide a comma-separated list of GitHub organizations (by name, not ID) that the user wishes to query for contributions.  All repositories in the list of organizations provided using this flag will be queried for contributions, and the output generated will include all contributions (by type based on the subcommand used) made to any repository in this list of organizations.  If this flag is not used to specify the list of organizations that the user wants to query, then the `orgs` parameter from the configuration file associated with this application will be used to set the default list of organizations that should be queried.

##### The `-u, --user-list` flag

This flag can be used to provide a comma-separated list of users (by user name, not GitHub ID) that the user wishes to query for contributions from. The mapping between the user names passed in using this flag and their associated GitHub user IDs must be defined in the associated configuration file If this flag is used to specify the list of users to query for contributions from.  If a user list is not specified using this flag (or the alternate `-i, --github-id-list` flag, see below for more details), then the list of users in the team being used for comparison (if a team was passed in on the command line) will be used.  If no team is specified, then the `default_team` from the associated configuration file will be used to define the list of users to query for contributions from.

##### The `-i, --github-id-list` flag

This flag can be used to provide a comma-separated list of users (by GitHub ID) that the user wishes to query for contributions from. This flag is offered as an alternative to using the `-u, --user-list` flag to accomplish this same task.  If both of those flags are used then an error will be thrown.  If a user list is not specified using this flag or the alternate  `-u, --user-list` flag (see above for details), then the list of users in the team being used for comparison (if a team was passed in on the command line) will be used.  If no team is specified, then the `default_team` from the associated configuration file will be used to define the list of users to query for contributions from.

##### The `-f, --file` flag

This flag can be used to direct the output of the application to the named file.  If this flag is not used, then output will be directed to standard output by default.  Error messages are all directed to standard error, so it should be relatively easy to separate any errors or warnings that occur from the results of the queries run, even when this flag has not been set.

#### Defining the time window for queries

There are two flags that are used to define the time window that should be queried for contributions, and those two parameters are as follows:

* **the `-d, --ref-date` flag**: the string value passed into the program using this flag is assumed to be of the format `YYYY-MM-DY` that specifies the "reference date" to use when constructing the time window for the query.  If a string value is passed in using this flag that doesn't correspond to that format (4 digit year, two digit month, and two digit day separated by dashes), then an error will be thrown by the program.
* **the `-l, --lookback-time` flag**: the string value passed into the program using this flag must be of the form of an integer number followed by a single letter suffix, and the combination of that integer and suffix that indicates the time period to look **back** from the reference date when defining the time window to query.

It is the combination of the values passed in for these two flags that determines the time window that will be used when querying for contributions from the users on the team. That said, there are a few important things to remember:

* All queries are based on zero hours, UTC; there is no option to shift the starting date-times by anything less than a day using the defined lookback time argument
* If the reference date that is passed in exceeds the current date plus the lookback time (if any) that was passed in, then an error will likely result.
* The lookback time passed in must match a regular expression of the form `^[+-]?[0-9]+[dwmqy]$`. As you can clearly see, this means that the lookback time consists of an integer value (which can be either positive or negative) with a single letter suffix that represents the time units for that value: `d` for days, `w` for weeks, `m` for months (where a month is defined as 30 days, `q` for quarters (where a quarter is defined as 90 days), or `y` for years.  So, for example, you would pass in in a lookback time of `12w` if you wanted the start of the time window for the queries to be 12 weeks, or or 84 days prior to the reference date.
* As was mentioned previously, the lookback time passed in can be a negative number.  In that case, you are actually instructing the system to look **ahead** by the corresponding number of days, weeks, etc. from the input reference date; if a negative number is passed in and represents more time than exists between the reference date and the current date it is not an error, but there will obviously be no contributions to be found in the database that are more recent than the current date's contributions.
* Due to limitations in the GitHub GraphQL API, only a year's worth of data can be returned in a single query.  As such, the lookback time passed into the application cannot exceed one year (or 365 days) without an error being thrown by the application.  If there is a need to retrieve more than one year's worth of data in the future we can explore making changes to this application to detect this scenario and handle it appropriately (by breaking such a query up into two or more queries where none of them exceed a year in length). This doesn't mean that data can't be retrieve from more than a year in the past, just that the lookback time cannot exceed one year (plus or minus) relative to the reference date.

In addition to whether the value passed in using the lookback flag is a positive or negative, which of these two flags are included on the command-line will determine how the values that are included on the command line are used to define the time window. 

* If a reference date is not included on the command-line, then the end date for the time window to use when querying the system is assumed to be the current date.
  * In this situation, if a lookback time is also not included, then a default lookback time of 90 days is used to construct the time window to use (a time window that covers the 90 days prior to the current date)
* If a reference date is included on the command-line but a lookback time is not included on the command-line, then the time window used is assumed to be one that starts on the reference date ends on the current date.
* If both a reference date and a lookback time are provided on the command line, then those values are used to define the time window based on the lookback time value, it's units, and whether a positive or negative "lookback" time value was specified (where a negative value, as was mentioned previously, is used to indicate that the user wants to actually look **ahead** of the reference time by the stated lookback time value)

If you keep these basic rules in mind, we think it is easy to see that pretty much any time window can be defined using these two parameters, making it possible to look for contributions of the specified time from the specified users listed.  The only limitation is that any time window defined cannot be longer than one year in length (in whatever units were used to specify the lookback time)
