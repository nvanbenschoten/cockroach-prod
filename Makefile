all: build check

build:
	bunch go build

GOFILES := $(shell find . -name '*.go' | grep -vF '/.')

check:
	bunch exec errcheck -ignore 'github.com/spf13/cobra:Usage' ./...
	bunch exec golint ./...
	! bunch exec gofmt -s -l -w $(GOFILES) | grep -vF 'No Exceptions'
	! bunch exec goimports -l -w $(GOFILES) | grep -vF 'No Exceptions'
	go vet ./...
	go tool vet --shadow $(GOFILES)

.bootstrap: Bunchfile Bunchfile.lock
	go get github.com/dkulchenko/bunch
	bunch install
	touch $@

-include .bootstrap
