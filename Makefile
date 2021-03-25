NAME := nullstone

.PHONY: setup test

.DEFAULT_GOAL: default

default: setup

setup:
	cd ~ && go get gotest.tools/gotestsum && cd -
	brew install goreleaser/tap/goreleaser || (cd ~ && go install github.com/goreleaser/goreleaser && cd -)

build:
	goreleaser --snapshot --skip-publish --rm-dist

test:
	go fmt ./...
	gotestsum
