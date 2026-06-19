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
	cp ./docs/CLI.md ../docs/src/getting-started/cli/docs.md

upgrade-aws:
	go get -u github.com/aws/aws-sdk-go-v2/...

upgrade-gcp:
	go get -u cloud.google.com/...

upgrade-k8s:
	go get -u k8s.io/...
