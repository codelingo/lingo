#!/bin/bash

# Requires go 1.5 or higher
# Must be run from root of github.com/codelingo/lingo
#
# $ scripts/cross_compile.sh 0.1.0

set -x
if [ $# -eq 0 ]
  then
    echo "No arguments supplied, need verion"
    exit
fi

version=$1


repoRoot=$GOPATH/src/github.com/codelingo/lingo

description=`cat $repoRoot/scripts/next_release_notes.md`

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
	GOOS=$os GOARCH=$arch go build -o $binpath/lingo -v github.com/codelingo/lingo

	cd bin
	filename=lingo-$version-$os-$arch
	if [ "$os" == "windows" ]; then
		fn="$filename.zip"
		zip $binpath/$fn lingo
	else
		fn="$filename.tar.gz"
		tar -cvzf $binpath/$fn lingo
	fi
	cd ..

	rm $binpath/lingo

	github-release upload \
    --user codelingo \
    --repo lingo \
    --tag $version \
    --name $fn \
    --file $binpath/$fn

done

echo "Done!"