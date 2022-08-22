# Tracking GitHub Contributions

This directory contains a Go program that can be used to gather information regarding contributions made by an input list of team members (listed by GitHub ID) to the repositories contained in an input list of GitHub organizations. The program collects contribution data for each user via GitHub's GraphQL API, then outputs that data to the named output file as a spreadsheet containing the raw contribution data for each team member along with basic statistics for that user (including some statistics comparing that user's contributions to the contributions of the team as a whole).

The input list of team members can be passed in as a comma-separated list of GitHub IDs on the command-line or via a field in a configuration file passed in (by filename) on the command-line. Similarly, the list of GitHub organizations that you want to track contributions to for each user can be passed in as a comma-separated list of organization names on the command-line or that list of organization names can be passed in via a field in that same configuraiton file. Details for the format(s) supported for that configuration file are sketched out, below.

## Usage
