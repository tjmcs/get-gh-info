#!/usr/bin/env bash

# a list of all of the PRs made that were closed as merged (in a pretty-print JSON format)
echo "--------------------------------------------------------------------------------"
echo "PRs that were closed as merged by the team:"
echo "--------------------------------------------------------------------------------"
jq "[ .pullRequests.ByUser | .[][][] | select( .closed ) | select( .merged ) ]" < all-pull-contrib-list-12-mos.json

# a list of all of the PRs made that were closed as merged (in a pretty-print JSON format)
echo "--------------------------------------------------------------------------------"
echo "PR Reviews that were closed as merged by the team:"
echo "--------------------------------------------------------------------------------"
jq "[ .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) ]" < all-pull-contrib-list-12-mos.json

# and a list of all of the repositories that those PRs were against
echo "--------------------------------------------------------------------------------"
echo "Repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u

# broken out by repository "type"; first the Orb repositories
echo "--------------------------------------------------------------------------------"
echo "Orb repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | grep -i orb

# then the image repositories
echo "--------------------------------------------------------------------------------"
echo "Image repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | grep img | grep -iv orb

# then the sample project directories
echo "--------------------------------------------------------------------------------"
echo "Sample project repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
sampleIncludes="cfd|CFD|Demo-Game"
jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | grep -Ei "$sampleIncludes"

# then the sample project directories
echo "--------------------------------------------------------------------------------"
echo "'Typical' project repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
typicalIncludes="sdk-ts|visual-config-editor|developer-hub-indexer"
jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | grep -Ei "$typicalIncludes"


# then the rest (the "other" repositories)
echo "--------------------------------------------------------------------------------"
echo "Other repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
otherExcludes="orb|img|$sampleIncludes|$typicalIncludes"
jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | grep -Eiv "$otherExcludes"

# and the number of repositories modified
echo "--------------------------------------------------------------------------------"
totalModified=$(jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | wc -l)
printf "%s  %4d\n" "Number of repositories modified:" "$totalModified"

# the number of Orb repositories modified
num=$(jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | grep -ic orb)
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Number of orb repositories:"  "$num" "$pctContrib"

# the number of Image repositories that modified
num=$(jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | grep img | grep -ivc orb)
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Number of image repositories:"  "$num" "$pctContrib"

# the number of Sample Project repositories modified
num=$(jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | grep -Eic "$sampleIncludes")
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Number of sample projects:"  "$num" "$pctContrib"

# the number of "typical" Project repositories modified
num=$(jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | grep -Eic "$typicalIncludes")
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Number of 'typical' projects:"  "$num" "$pctContrib"

# the number of "other" repositories modified
num=$(jq '.pullRequests.ByUser, .pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .repositoryName' < all-pull-contrib-list-12-mos.json | sort -u | grep -Eivc "$otherExcludes")
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Number of 'other' repositories:"  "$num" "$pctContrib"

# then show the total number of PRs merged
numPRs=$(jq '.pullRequests.ByUser | .[][][] | select( .closed ) | select( .merged ) | .url' < all-pull-contrib-list-12-mos.json | wc -l | xargs)
printf "%s\t\t  %4d\n" "Number of PRs merged:" "$numPRs"

# and the number of Pull Request Reviews performed
numPrReviews=$(jq '.pullRequestReviews.ByUser | .[][][] | select( .closed ) | select( .merged ) | .url' < all-pull-contrib-list-12-mos.json | wc -l | xargs)
printf "%s\t\t  %4d\n" "Number of PR reviews:" "$numPrReviews"

# calculate the total contributions made from the number or PRs and PR reviews
totalContribs=$(echo "$numPRs + $numPrReviews" | bc -l)

# and show the breakdown in the contributions by the type of asset;
# first the orb repositories
numPrs=$(jq -c '.pullRequests.AllUsers | keys[]' all-pull-contrib-list-12-mos.json | grep -i orb | while read key; do
    echo -n $(jq ".pullRequests.AllUsers | .$key | .totalContributions" all-pull-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
numPrReviews=$(jq -c '.pullRequestReviews.AllUsers | keys[]' all-pull-contrib-list-12-mos.json | grep -i orb | while read key; do
    echo -n $(jq ".pullRequestReviews.AllUsers | .$key | .totalContributions" all-pull-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
num=$(echo "$numPrs + $numPrReviews" | bc -l)
pctContrib=$(echo "$num / $totalContribs * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Contribs to orb repositories:"  "$num" "$pctContrib"

# then the image repositories
numPrs=$(jq -c '.pullRequests.AllUsers | keys[]' all-pull-contrib-list-12-mos.json | grep -i img | grep -iv orb | while read key; do
    echo -n $(jq ".pullRequests.AllUsers | .$key | .totalContributions" all-pull-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
numPrReviews=$(jq -c '.pullRequestReviews.AllUsers | keys[]' all-pull-contrib-list-12-mos.json | grep -i img | grep -iv orb | while read key; do
    echo -n $(jq ".pullRequestReviews.AllUsers | .$key | .totalContributions" all-pull-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
num=$(echo "$numPrs + $numPrReviews" | bc -l)
pctContrib=$(echo "$num / $totalContribs * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Contribs to image repositories:"  "$num" "$pctContrib"

# then the sample project repositories
numPrs=$(jq -c '.pullRequests.AllUsers | keys[]' all-pull-contrib-list-12-mos.json | grep -Ei "$sampleIncludes" | while read key; do
    echo -n $(jq ".pullRequests.AllUsers | .$key | .totalContributions" all-pull-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
numPrReviews=$(jq -c '.pullRequestReviews.AllUsers | keys[]' all-pull-contrib-list-12-mos.json | grep -Ei "$sampleIncludes" | while read key; do
    echo -n $(jq ".pullRequestReviews.AllUsers | .$key | .totalContributions" all-pull-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
num=$(echo "$numPrs + $numPrReviews" | bc -l)
pctContrib=$(echo "$num / $totalContribs * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Contribs to sample projects:"  "$num" "$pctContrib"

# the more 'typical' repositories
numPrs=$(jq -c '.pullRequests.AllUsers | keys[]' all-pull-contrib-list-12-mos.json | grep -Ei "$typicalIncludes" | while read key; do
    echo -n $(jq ".pullRequests.AllUsers | .$key | .totalContributions" all-pull-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
numPrReviews=$(jq -c '.pullRequestReviews.AllUsers | keys[]' all-pull-contrib-list-12-mos.json | grep -Ei "$typicalIncludes" | while read key; do
    echo -n $(jq ".pullRequestReviews.AllUsers | .$key | .totalContributions" all-pull-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
num=$(echo "$numPrs + $numPrReviews" | bc -l)
pctContrib=$(echo "$num / $totalContribs * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Contribs to 'typical' repos:"  "$num" "$pctContrib"

# and finally the 'other' repositories
numPrs=$(jq -c '.pullRequests.AllUsers | keys[]' all-pull-contrib-list-12-mos.json | grep -Eiv "$otherExcludes" | while read key; do
    echo -n $(jq ".pullRequests.AllUsers | .$key | .totalContributions" all-pull-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
numPrReviews=$(jq -c '.pullRequestReviews.AllUsers | keys[]' all-pull-contrib-list-12-mos.json | grep -Eiv "$otherExcludes" | while read key; do
    echo -n $(jq ".pullRequestReviews.AllUsers | .$key | .totalContributions" all-pull-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
# num=$(echo "$numPrs + $numPrReviews" | bc -l)
num=$(echo "$numPrs + $numPrReviews" | bc -l)
pctContrib=$(echo "$num / $totalContribs * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Contribs to 'other' repos:"  "$num" "$pctContrib"