
export GOPATH := $(GOPATH):$(PWD)

.PHONY: all deps test build

all: test

deps:

test:
	go test -test.v ./src/cmd

build:
	go build -o bin/goref ./src/cmd
