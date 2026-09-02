# variables for compiling and pushing
APP       = $(shell basename $(PWD))
VERSION   = $(shell git describe --always --dirty --tags)
GOFLAGS   = -trimpath
LDFLAGS   = -w -s
GOFILES   = $(wildcard *.go)
TIMESTAMP = $(shell date -Iseconds)

all:		$(APP)

# Build the application
$(APP): go.mod go.sum $(GOFILES) Makefile
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(APP)
	@ls -l $(APP)
	@file $(APP)

# run static checks
vet:
	go vet
	go tool staticcheck ./...
	go tool deadcode -test -generated ./...
	go tool gocritic check --enableAll ./...

# run security and compliance checks
sec:
	go tool govulncheck -show verbose ./...
	go tool gosec ./...
	# go tool go-libyear --versions --releases go.mod   ## not needed as there are no external dependencies

# run unit tests
test:
	go test

# Coverage analysis
cover:
	go test -cover -coverprofile=coverage.out -v ./...
	go tool cover -html=coverage.out


# Cleanup
clean:
	go clean
	rm -f coverage.out

# show version information
version:
	@echo $(VERSION)
	@go version


# update linting tools
toolupdate:
	for tool in $(shell go tool | grep / | cut -d "(" -f2 | cut -d ")" -f1); do \
	  echo "Updating $$tool..."; \
	  go get $$tool@latest; \
	done
	go mod tidy

.PHONY: vet sec test cover clean version
.PHONY: toolupdate
