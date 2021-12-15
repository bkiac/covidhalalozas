#!/bin/sh

ARCHIVE=victims.csv
DATE=$(date '+%Y-%m-%d %H:%M:%S')

git config user.email "$GITHUB_ACTOR@users.noreply.github.com"
git config user.name "$GITHUB_ACTOR"
echo "Checking for uncommitted changes in the git working tree."
if expr $(git status $ARCHIVE --porcelain | wc -l) \> 0
then
	git add $ARCHIVE
	git commit -m "Update archive $DATE"
	git push
else
	echo "Working tree clean. Nothing to commit."
fi
