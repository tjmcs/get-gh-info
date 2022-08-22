#!/usr/bin/env bash

# a list of all of the PRs made that were closed as merged (in a pretty-print JSON format)
echo "--------------------------------------------------------------------------------"
echo "PRs that were closed as merged reviewed by the team:"
echo "--------------------------------------------------------------------------------"
jq "[ .[][] | select( .closed ) | select( .merged ) ]" < pr-review-list-12-mos.json

# and a list of all of the repositories that those PRs were against
echo "--------------------------------------------------------------------------------"
echo "Repositories that the team reviewed merged PRs for:"
echo "--------------------------------------------------------------------------------"
jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" < pr-review-list-12-mos.json | sort -u

# broken out by repository "type"; first the Orb repositories
echo "--------------------------------------------------------------------------------"
echo "Orb repository reviews that the team contributed to:"
echo "--------------------------------------------------------------------------------"
jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" < pr-review-list-12-mos.json | sort -u | grep orb

# thenthe image repositories
echo "--------------------------------------------------------------------------------"
echo "Image repository reviews that the team contributed to:"
echo "--------------------------------------------------------------------------------"
jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" < pr-review-list-12-mos.json | sort -u | grep img | grep -v orb

# then the sample project directories
echo "--------------------------------------------------------------------------------"
echo "Sample project repository reviews that the team contributed to:"
echo "--------------------------------------------------------------------------------"
jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" < pr-review-list-12-mos.json | sort -u | grep -Ei "cfd|demo-game"

# then the rest (the "other" repositories)
echo "--------------------------------------------------------------------------------"
echo "Other repository reviews that the team contributed to:"
echo "--------------------------------------------------------------------------------"
jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" < pr-review-list-12-mos.json | sort -u | grep -Ev "img|orb|CFD|cfd|Demo-Game"

# the number of PRs merged
echo "--------------------------------------------------------------------------------"
num=$(jq ".[][] | select( .closed ) | select( .merged ) | .url" < pr-review-list-12-mos.json | wc -l | xargs)
printf "%s\t\t\t  %4d\n" "Number of PRs reviewed:" "$num"

# and the number of repositories that PRs were merged into
num=$(jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" < pr-review-list-12-mos.json | sort -u | wc -l)
printf "%s  %4d\n" "Number of repositories PRs reviewed for:" "$num"

# the number of Orb repositories that PRs were meged into
num=$(jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" < pr-review-list-12-mos.json | sort -u | grep -c orb)
printf "%s\t  %4d\n" "Number of orb repositoriy reviews:"  "$num"

# the number of Orb repositories that PRs were meged into
num=$(jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" < pr-review-list-12-mos.json | sort -u | grep img | grep -vc orb)
printf "%s\t  %4d\n" "Number of image repository reviews:"  "$num"

# the number of Orb repositories that PRs were meged into
num=$(jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" < pr-review-list-12-mos.json | sort -u | grep -Eic "cfd|demo-game")
printf "%s\t  %4d\n" "Number of sample project reviews:"  "$num"

# the list of "other" repositories that PRs were meged into
# cat pr-review-list-12-mos.json | jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" | sort -u | egrep -v "img|orb|CFD|cfd"

# the number of "other" repositories that PRs were meged into
num=$(jq ".[][] | select( .closed ) | select( .merged ) | .repositoryName" < pr-review-list-12-mos.json | sort -u | grep -Evc "img|orb|CFD|cfd|Demo-Game")
printf "%s\t  %4d\n" "Number of 'other' repository reviews:"  "$num"
