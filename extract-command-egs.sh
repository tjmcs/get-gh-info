#!/bin/bash

# example of a command to extract the count/list of pull requests from each
# user that took longer than two weeks to merge from the 'pr-list.json' file
# using jq
for user in felicianotech KyleTryon BytesGuy EricRibeiro Jalexchen Jaryt brivu mrothstein74; do
  echo -n "PRs that took longer than two weeks from $user: "
  jq "[ .[\"$user\"][] | select( .closed ) | select( .merged ) | select( .daysWorked > 14 ) ] | length" < pr-list.json
  echo "  in the following repositories:"
  jq ".[\"$user\"][]| select( .closed ) | select( .merged ) | select( .daysWorked > 14 )" < pr-list.json | awk -F'$' '{ printf "    %s\n", $1 }'
done | tee two-week-plus-prs-"$(TZ=UTC date +%d-%b-%Y)".t

# example of a command to extract the number of pull requests merged
# and count/list of repositories those pull requests were made against
# from the 'pr-list.json' file using jq
for user in felicianotech KyleTryon BytesGuy EricRibeiro Jalexchen Jaryt brivu mrothstein74; do
  echo -n "PRs merged from $user: "
  jq ".[\"$user\"][] | select( .closed ) | select( .merged )" < pr-list.json | grep -Ec 'title' | xargs
  numRepos=$(jq ".[\"$user\"][] | select( .closed ) | select( .merged )" < pr-list.json | grep -E 'Name' | sort -u | wc -l | xargs)
  echo -n "  from $numRepos repositories, "
  echo "which are as follows:"
  jq ".[\"$user\"][] | select( .closed ) | select( .merged )" < pr-list.json | grep -E 'Name' | sort -u | awk -F'$' '{ printf "    %s\n", $1 }'
done | tee pr-nos-"$(TZ=UTC date +%d-%b-%Y)".t

# and an example of a command to extract the number of pull request reviews
# performed and count/list of repositories those pull request reviews were
# performed for from the 'pr-review-list.json' file using jq
for user in felicianotech KyleTryon BytesGuy EricRibeiro Jalexchen Jaryt brivu mrothstein74; do
  echo -n "PR reviews performed by $user: "
  jq ".[\"$user\"][] | select( .closed ) | select( .merged )" < pr-review-list.json | grep -Ec 'title' | xargs
  numRepos=$(jq ".[\"$user\"][] | select( .closed ) | select( .merged )" < pr-review-list.json | grep -E 'Name' | sort -u | wc -l | xargs)
  echo -n "  from $numRepos repositories, "
  echo "which are as follows:"
  jq ".[\"$user\"][] | select( .closed ) | select( .merged )" < pr-review-list.json | grep -E 'Name' | sort -u | awk -F'$' '{ printf "    %s\n", $1 }'
done | tee pr-review-nos-"$(TZ=UTC date +%d-%b-%Y)".t