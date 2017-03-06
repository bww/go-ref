
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

# tests
TEST_PACKAGES := ./src/cmd
TEST_FIXTURES := basic

.PHONY: all build test clean

all: build

$(PRODUCT): $(SRC)
	go build -o $@ $(MAIN)

build: $(PRODUCT) ## Build the product

test: export REF_TEST_DATA := $(PWD)/test
test: ## Run tests
	go test -test.v $(TEST_PACKAGES)
	$(PWD)/test/bin/run.sh $(addprefix $(PWD)/test/data/, $(TEST_FIXTURES))

clean: ## Delete the built product and any generated files
	rm -rf $(TARGETS)
