SHELL:=/bin/bash
VERSION?=$(shell git describe --tags --always)
CURRENT_DOCKER_IMAGE=quintilesims/iqvbot:$(VERSION)
LATEST_DOCKER_IMAGE=quintilesims/iqvbot:latest

deps: 
	go get github.com/golang/mock/mockgen/model
	go install github.com/golang/mock/mockgen

mocks:
	mockgen -package mock github.com/quintilesims/iqvbot/utils SlackClient > mock/mock_slack_client.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a --ldflags "-X main.Version=$(VERSION)" -o iqvbot . 
	docker build -t $(CURRENT_DOCKER_IMAGE) .

test:
	go test ./... -v

run:
	ngrok http 9090

release: build
	docker push $(CURRENT_DOCKER_IMAGE)
	docker tag  $(CURRENT_DOCKER_IMAGE) $(LATEST_DOCKER_IMAGE)
	docker push $(LATEST_DOCKER_IMAGE)

.PHONY: deps mocks build test run release
