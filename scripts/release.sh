#!/bin/bash

# Requires go 1.5 or higher
# Run from root of repo
#
# $ scripts/cross_compile.sh 0.1.0

set -e
if [ $# -eq 0 ]
  then
    echo "No arguments supplied, need verion"
    exit
fi

version=$1


repoRoot=$GOPATH/src/github.com/codelingo/lingo

description=$(cat $repoRoot/scripts/next_release_notes.md)

# init array
compressedFilenames=()

binpath=$repoRoot/bin

v="
windows,386;\
linux,386;\
windows,amd64;\
linux,amd64;\
darwin,amd64;"

# Make a tag
git tag $version && git push --tags

# Push tag as release to github
# https://github.com/aktau/github-release
github-release release \
    --user codelingo \
    --repo lingo \
    --tag $version \
    --name $version \
    --description $description \
    --pre-release

# Build and push each bin to release
echo $v | while IFS=',' read -d';' os arch;  do 
	echo "HERE"
	GOOS=$os GOARCH=$arch go build -o $binpath/lingo -v github.com/codelingo/lingo

	filename=lingo-$version-$os-$arch
	if [ "$os" == "windows" ]; then
		fn="$filename.zip"
		zip $binpath/$fn $binpath/lingo
	else
		fn="$filename.tar.gz"
		tar -cvzf $binpath/$fn $binpath/lingo
	fi

	rm $binpath/lingo

	github-release upload \
    --user codelingo \
    --repo lingo \
    --tag $version \
    --name $fn \
    --file $binpath/$fn

done

echo "Done!"