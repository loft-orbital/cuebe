# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run

BUILD_FILES = $(shell go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}' ./...)
BUILD_DIR = bin/
BIN = $(addprefix $(BUILD_DIR), $(notdir $(CURDIR)))

LDFLAGS = -s -w

$(BIN): $(BUILD_FILES)
	$(GOBUILD) -trimpath -o "$@" -ldflags='$(LDFLAGS)'
build: $(BIN)
.PHONY: build

test: $(BUILD_FILES)
		$(GOTEST) -cover -race ./... -coverprofile c.out -timeout 10s
.PHONY: test

test-short: $(BUILD_FILES)
	$(GOTEST) -short ./...  -coverprofile c.out
.PHONY: test-short

bench: $(BUILD_FILES)
	$(GOTEST) -run=xxx -bench=. ./...
.PHONY: bench

doc: $(BUILD_FILES)
		$(GOCMD) doc -all
.PHONY: doc
