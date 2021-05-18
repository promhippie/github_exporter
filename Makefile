include .bingo/Variables.mk

SHELL := bash
NAME := github_exporter
IMPORT := github.com/promhippie/$(NAME)
BIN := bin
DIST := dist

ifeq ($(OS), Windows_NT)
	EXECUTABLE := $(NAME).exe
	UNAME := Windows
else
	EXECUTABLE := $(NAME)
	UNAME := $(shell uname -s)
endif

GOBUILD ?= CGO_ENABLED=0 go build
PACKAGES ?= $(shell go list ./...)
SOURCES ?= $(shell find . -name "*.go" -type f -not -path "./node_modules/*")
GENERATE ?= $(PACKAGES)

TAGS ?= netgo

ifndef OUTPUT
	ifneq ($(DRONE_TAG),)
		OUTPUT ?= $(subst v,,$(DRONE_TAG))
	else
		OUTPUT ?= testing
	endif
endif

ifndef VERSION
	ifneq ($(DRONE_TAG),)
		VERSION ?= $(subst v,,$(DRONE_TAG))
	else
		VERSION ?= $(shell git rev-parse --short HEAD)
	endif
endif

ifndef DATE
	DATE := $(shell date -u '+%Y%m%d')
endif

ifndef SHA
	SHA := $(shell git rev-parse --short HEAD)
endif

LDFLAGS += -s -w -extldflags "-static" -X "$(IMPORT)/pkg/version.String=$(VERSION)" -X "$(IMPORT)/pkg/version.Revision=$(SHA)" -X "$(IMPORT)/pkg/version.Date=$(DATE)"
GCFLAGS += all=-N -l

.PHONY: all
all: build

.PHONY: sync
sync:
	go mod download

.PHONY: clean
clean:
	go clean -i ./...
	rm -rf $(BIN) $(DIST)

.PHONY: fmt
fmt:
	gofmt -s -w $(SOURCES)

.PHONY: vet
vet:
	go vet $(PACKAGES)

.PHONY: staticcheck
staticcheck: $(STATICCHECK)
	$(STATICCHECK) -tags '$(TAGS)' $(PACKAGES)

.PHONY: lint
lint: $(GOLINT)
	for PKG in $(PACKAGES); do $(GOLINT) -set_exit_status $$PKG || exit 1; done;

.PHONY: generate
generate:
	go generate $(GENERATE)

.PHONY: changelog
changelog: $(CALENS)
	$(CALENS) >| CHANGELOG.md

.PHONY: test
test:
	go test -coverprofile coverage.out $(PACKAGES)

.PHONY: install
install: $(SOURCES)
	go install -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' ./cmd/$(NAME)

.PHONY: build
build: $(BIN)/$(EXECUTABLE)

$(BIN)/$(EXECUTABLE): $(SOURCES)
	$(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(BIN)/$(EXECUTABLE)-debug: $(SOURCES)
	$(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -gcflags '$(GCFLAGS)' -o $@ ./cmd/$(NAME)

.PHONY: release
release: $(DIST) release-linux release-darwin release-windows release-reduce release-checksum

$(DIST):
	mkdir -p $(DIST)

.PHONY: release-linux
release-linux: $(DIST) \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-386 \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-amd64 \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-arm-5 \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-arm-6 \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-arm-7 \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-arm64 \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-mips \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-mips64 \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-mipsle \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-mips64le

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-386:
	GOOS=linux GOARCH=386 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-arm-5:
	GOOS=linux GOARCH=arm GOARM=5 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-arm-6:
	GOOS=linux GOARCH=arm GOARM=6 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-arm-7:
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-mips:
	GOOS=linux GOARCH=mips $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-mips64:
	GOOS=linux GOARCH=mips64 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-mipsle:
	GOOS=linux GOARCH=mipsle $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-linux-mips64le:
	GOOS=linux GOARCH=mips64le $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

.PHONY: release-darwin
release-darwin: $(DIST) \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-darwin-amd64 \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-darwin-arm64

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

.PHONY: release-windows
release-windows: $(DIST) \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-windows-4.0-386.exe \
	$(DIST)/$(EXECUTABLE)-$(OUTPUT)-windows-4.0-amd64.exe

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-windows-4.0-386.exe:
	GOOS=windows GOARCH=386 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

$(DIST)/$(EXECUTABLE)-$(OUTPUT)-windows-4.0-amd64.exe:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

.PHONY: release-reduce
release-reduce:
	cd $(DIST); $(foreach file,$(wildcard $(DIST)/$(EXECUTABLE)-*),upx $(notdir $(file));)

.PHONY: release-checksum
release-checksum:
	cd $(DIST); $(foreach file,$(wildcard $(DIST)/$(EXECUTABLE)-*),sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)

.PHONY: release-finish
release-finish: release-reduce release-checksum

.PHONY: docs
docs:
	hugo -s docs/

.PHONY: envvars
envvars:
	go run hack/generate-envvars-docs.go

.PHONY: metrics
metrics:
	go run hack/generate-metrics-docs.go

.PHONY: watch
watch:
	$(REFLEX) -c reflex.conf
