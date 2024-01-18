NAME := nullstone

.PHONY: setup test docs

.DEFAULT_GOAL: default

default: setup

setup:
	cd ~ && go get gotest.tools/gotestsum && cd -
	brew install goreleaser/tap/goreleaser || (cd ~ && go install github.com/goreleaser/goreleaser && cd -)

build:
	goreleaser --snapshot --skip-publish --rm-dist

test:
	go fmt ./...
	gotestsum ./...

docs:
	go run ./docs/main.go

upgrade-aws:
	go get -u github.com/aws/aws-sdk-go-v2/...
