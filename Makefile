NAME := nullstone

.PHONY: setup test

.DEFAULT_GOAL: default

default: setup

setup:
	cd ~ && go get gotest.tools/gotestsum && cd -

build:
	@go build -ldflags "-X 'main.Version=$(VERSION)'" -o dist/$(NAME) .

test:
	go fmt ./...
	gotestsum
