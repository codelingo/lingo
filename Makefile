.PHONY: install

# for future reference: https://gist.github.com/madhums/45efcb78d5d0d654191f

PKG_NAME=$(shell basename `pwd`)

updateproto:
	go get -u github.com/golang/protobuf/proto
	go get -u github.com/golang/protobuf/protoc-gen-go

install:
	# first install dep to manage deps
	go get -u github.com/golang/dep/cmd/dep

	# then get the latest commit
	git pull
	
	# install deps
	dep ensure -v

	# build and install lingo
	go install

test:
	scripts/./pre-push
# waigani xxx
# git hg bazar
# hg launchpad
