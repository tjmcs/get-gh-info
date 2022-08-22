#!/usr/bin/env bash

# and a list of all of the repositories that those PRs were against
echo "--------------------------------------------------------------------------------"
echo "Repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u

# broken out by repository "type"; first the Orb repositories
echo "--------------------------------------------------------------------------------"
echo "Orb repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | grep -i orb

# then the image repositories
echo "--------------------------------------------------------------------------------"
echo "Image repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | grep -i img | grep -iv orb

# then the sample project directories
echo "--------------------------------------------------------------------------------"
echo "Sample project repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
sampleIncludes="cfd|demo-game"
jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | grep -Ei "$sampleIncludes"

# then the sample project directories
echo "--------------------------------------------------------------------------------"
echo "More typical project repositories contributed to:"
echo "--------------------------------------------------------------------------------"
typicalIncludes="sdk-ts|visual-config-editor|developer-hub-indexer"
jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | grep -Ei "$typicalIncludes"

# then the rest (the "other" repositories)
echo "--------------------------------------------------------------------------------"
echo "Other repositories that the team contributed to:"
echo "--------------------------------------------------------------------------------"
otherExcludes="orb|img|$sampleIncludes|$typicalIncludes"
jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | grep -Eiv "$otherExcludes"

# the number of repositories modified
echo "--------------------------------------------------------------------------------"
totalModified=$(jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | wc -l)
printf "%s\t  %4d\n" "Repositories contributed to:" "$totalModified"

# the number of Orb repositories modified
num=$(jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | grep -ic orb)
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Number of orb repositories:"  "$num" "$pctContrib"

# the number of Image repositories that modified
num=$(jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | grep -i img | grep -ivc orb)
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Number of image repositories:"  "$num" "$pctContrib"

# the number of Sample Project repositories modified
num=$(jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | grep -Eic "$sampleIncludes")
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Number of sample projects:"  "$num" "$pctContrib"

# the number of "typical" repositories modified
num=$(jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | grep -Eic "$typicalIncludes")
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Number of 'typical' repos:"  "$num" "$pctContrib"

# the number of "other" repositories modified
num=$(jq '.ByUser[][][] | .repositoryName' < all-contrib-list-12-mos.json | sort -u | grep -Eivc "$otherExcludes")
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Number of 'other' repos:"  "$num" "$pctContrib"

# the number of contributions made
totalContribs=$(jq '[.ByUser[][][] | .numContributions | tonumber] | add' < all-contrib-list-12-mos.json)
pctContrib=$(echo "$num / $totalModified * 100." | bc -l)
printf "%s\t  %4d\n" "Total contributions made:" "$totalContribs"

# and show the breakdown in the contributions by the type of asset;
# first the orb repositories
num=$(jq -c '.AllUsers | keys[]' all-contrib-list-12-mos.json | grep -i orb | while read key; do
    echo -n $(jq ".AllUsers | .$key | .totalContributions" all-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
pctContrib=$(echo "$num / $totalContribs * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Contribs to orb repositories:"  "$num" "$pctContrib"

# then the image repositories
num=$(jq -c '.AllUsers | keys[]' all-contrib-list-12-mos.json | grep -i img | grep -iv orb | while read key; do
    echo -n $(jq ".AllUsers | .$key | .totalContributions" all-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
pctContrib=$(echo "$num / $totalContribs * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Contribs to image repositories:"  "$num" "$pctContrib"

# then the sample project repositories
num=$(jq -c '.AllUsers | keys[]' all-contrib-list-12-mos.json | grep -Ei "$sampleIncludes" | while read key; do
    echo -n $(jq ".AllUsers | .$key | .totalContributions" all-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
pctContrib=$(echo "$num / $totalContribs * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Contribs to sample projects:"  "$num" "$pctContrib"

# the more 'typical' repositories
num=$(jq -c '.AllUsers | keys[]' all-contrib-list-12-mos.json | grep -Ei "$typicalIncludes" | while read key; do
    echo -n $(jq ".AllUsers | .$key | .totalContributions" all-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
pctContrib=$(echo "$num / $totalContribs * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Contribs to 'typical' repos:"  "$num" "$pctContrib"

# and finally the 'other' repositories
num=$(jq -c '.AllUsers | keys[]' all-contrib-list-12-mos.json | grep -Eiv "$otherExcludes" | while read key; do
    echo -n $(jq ".AllUsers | .$key | .totalContributions" all-contrib-list-12-mos.json) | sed 's/null//g' && echo -n " ";
done | sed 's/ $/\n/' | tr -s '[:space:]' | sed 's/ / + /g' | bc -l)
pctContrib=$(echo "$num / $totalContribs * 100." | bc -l)
printf "%s\t  %4d (%.2f%%)\n" "Contribs to 'other' repos:"  "$num" "$pctContrib"