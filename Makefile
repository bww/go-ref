
# the product we're building
NAME := goref
# the product's main package
MAIN := ./src/cmd
# fix our gopath
GOPATH := $(GOPATH):$(PWD)

# build and packaging
TARGETS	:= $(PWD)/bin
PRODUCT	:= $(TARGETS)/$(NAME)

# sources
SRC = $(shell find src -name \*.go -print)

.PHONY: all build test clean

all: build

$(PRODUCT): $(SRC)
	go build -o $@ $(MAIN)

build: $(PRODUCT) ## Build the product

test: ## Run tests
	go test -test.v cmd

clean: ## Delete the built product and any generated files
	rm -rf $(TARGETS)
