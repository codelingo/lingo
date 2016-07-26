#!/bin/bash

# Requires go 1.5 or higher
# To run:
#
# $ ./cross_compile.sh 0.1.0

set -e

version=$1
description=$(cat ./next_release_notes.md)
binpath=$GOPATH/src/github.com/codelingo/lingo/bin

v="
darwin,amd64;\
windows,amd64;\
windows,386;\
linux,386;\
linux,amd64;
"
# First build all bins

echo $v | while IFS=',' read -d';' os arch;  do 
filename=lingo-$version-$os-$arch

GOOS=$os GOARCH=$arch go build -o $binpath/lingo -v github.com/codelingo/lingo

if [ "$os" == "windows" ]; then
	fn=$binpath/$filename.zip
	zip $fn $binpath/lingo
else
	fn=$binpath/$filename.tar.gz
	tar -cvzf fn$ $binpath/lingo
fi
compressedFilenames+=($fn)

rm $binpath/lingo

done

# Then make a tag

echo "git tag $version && git push --tags"

# Then make a release
# https://github.com/aktau/github-release
echo "github-release release \
    --user codelingo \
    --repo lingo \
    --tag $version \
    --name $version \
    --description $description \
    --pre-release"

# Then push each bin for the release

for cFilename in "${compressedFilenames[@]}"
do:

echo "github-release upload \
    --user codelingo \
    --repo lingo \
    --tag $version \
    --name $cFilename \
    --file $binpath/$cFilename"

done

