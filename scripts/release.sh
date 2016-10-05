#!/bin/bash

# Requires go 1.5 or higher
# Must be run from root of github.com/codelingo/lingo
# To run:
# $ scripts/release.sh 0.1.0

set -x
if [ $# -eq 0 ]
  then
    echo "No arguments supplied, need verion"
    exit
fi

version="v$1"

repoRoot=$GOPATH/src/github.com/codelingo/lingo

description=`cat $repoRoot/scripts/next_release_notes.md`

# build changelog and add to description
lastTag=`git describe --abbrev=0 --tags`
lastReleaseSHA=`git rev-list -n 1 $lastTag`
# lastReleaseSHA="3284553324fb95b5bc2e592d03a7e71a2f94681f"
changelog=`git log --oneline --decorate $lastReleaseSHA..HEAD`
description="$description"$'\r'$'\r'"# Changelog"$'\r'$'\r'"$changelog"

# init array
compressedFilenames=()

binpath=$repoRoot/bin

v="
windows,386;\
linux,386;\
windows,amd64;\
linux,amd64;\
darwin,amd64;"

# following uses https://github.com/aktau/github-release

# first delete tag and release if already made
github-release delete \
    --user codelingo \
    --repo lingo \
    --tag $version
git tag -d $version && git push --tags

# Make a tag and push as release to github
git tag $version && git push --tags
github-release release \
    --user codelingo \
    --repo lingo \
    --tag $version \
    --name $version \
    --description "$description" \
    --pre-release

# Build and push each bin to release
echo $v | while IFS=',' read -d';' os arch;  do 

    if [ "$os" == "windows" ]; then
    ext=.exe
    else
    ext=""
    fi

	GOOS=$os GOARCH=$arch go build -o $binpath/lingo$ext -v github.com/codelingo/lingo

	cd bin
	filename=lingo-$version-$os-$arch
	if [ "$os" == "windows" ]; then
		fn="$filename.zip"
        rm $binpath/$fn
		zip $binpath/$fn lingo.exe
        rm lingo.exe
	else
		fn="$filename.tar.gz"
		tar -cvzf $binpath/$fn lingo
	    rm lingo
    fi
    cd ..


	github-release upload \
    --user codelingo \
    --repo lingo \
    --tag $version \
    --name $fn \
    --file $binpath/$fn

done

echo "Done!"