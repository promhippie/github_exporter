# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.8. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
BINGO_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Below generated variables ensure that every time a tool under each variable is invoked, the correct version
# will be used; reinstalling only if needed.
# For example for bingo variable:
#
# In your main Makefile (for non array binaries):
#
#include .bingo/Variables.mk # Assuming -dir was set to .bingo .
#
#command: $(BINGO)
#	@echo "Running bingo"
#	@$(BINGO) <flags/args..>
#
BINGO := $(GOBIN)/bingo-v0.8.1-0.20230820182247-0568407746a2
$(BINGO): $(BINGO_DIR)/bingo.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/bingo-v0.8.1-0.20230820182247-0568407746a2"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=bingo.mod -o=$(GOBIN)/bingo-v0.8.1-0.20230820182247-0568407746a2 "github.com/bwplotka/bingo"

CALENS := $(GOBIN)/calens-v0.3.0
$(CALENS): $(BINGO_DIR)/calens.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/calens-v0.3.0"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=calens.mod -o=$(GOBIN)/calens-v0.3.0 "github.com/restic/calens"

REFLEX := $(GOBIN)/reflex-v0.3.1
$(REFLEX): $(BINGO_DIR)/reflex.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/reflex-v0.3.1"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=reflex.mod -o=$(GOBIN)/reflex-v0.3.1 "github.com/cespare/reflex"

REVIVE := $(GOBIN)/revive-v1.3.4
$(REVIVE): $(BINGO_DIR)/revive.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/revive-v1.3.4"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=revive.mod -o=$(GOBIN)/revive-v1.3.4 "github.com/mgechev/revive"

STATICCHECK := $(GOBIN)/staticcheck-v0.4.6
$(STATICCHECK): $(BINGO_DIR)/staticcheck.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/staticcheck-v0.4.6"
	@cd $(BINGO_DIR) && GOWORK=off $(GO) build -mod=mod -modfile=staticcheck.mod -o=$(GOBIN)/staticcheck-v0.4.6 "honnef.co/go/tools/cmd/staticcheck"

