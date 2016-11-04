.PHONY: install

# for future reference: https://gist.github.com/madhums/45efcb78d5d0d654191f

PKG_NAME=$(shell basename `pwd`)

install:
	# first install glide to manage deps
	go get github.com/Masterminds/glide

	# then get the latest commit
	git pull

	# update dep shas
	glide update
	
	# install deps
	glide install

	# build and install lingo
	go install

test:
	scripts/./pre-push
# waigani xxx
# git hg bazar
# hg launchpad
