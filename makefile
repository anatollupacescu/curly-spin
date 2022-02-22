PROJECT:=$(shell go list -m)

GOOS?=linux
GOARCH?=amd64

RELEASE?=0.0.1
COMMIT := git-$(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

.PHONY: format test clean

format:
	@goimports -w -local $(PROJECT) starter/
	@gci -w -local $(PROJECT) .
	@gofmt -s -w starter/

test:
	# @go test -count 1 -race -cover ./starter/...
	go test -count 10 -race starter/new.go starter/new_test.go starter/logger.go 

clean:
	rm pam spin.pml.trail starter/doc.go