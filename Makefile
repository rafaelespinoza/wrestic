GO ?= go
GOSEC ?= gosec

BIN_DIR=bin
MAIN=$(BIN_DIR)/main
PKG_IMPORT_PATH=github.com/rafaelespinoza/wrestic

SRC_PATHS = . ./internal/...
TEST_DIR=/tmp/wrestic_test

.PHONY: build deps gosec test vet

build:
	mkdir -pv $(dir $(MAIN)) && $(GO) build -v -o $(MAIN) \
		-ldflags "\
			-X $(PKG_IMPORT_PATH)/internal/cmd.versionBranchName=$(shell git rev-parse --abbrev-ref HEAD) \
			-X $(PKG_IMPORT_PATH)/internal/cmd.versionBuildTime=$(shell date --utc +%FT%T%z) \
			-X $(PKG_IMPORT_PATH)/internal/cmd.versionCommitHash=$(shell git rev-parse --short=7 HEAD) \
			-X $(PKG_IMPORT_PATH)/internal/cmd.versionGoOSArch=$(shell $(GO) version | awk '{ print $$4 }' | tr '/' '_') \
			-X $(PKG_IMPORT_PATH)/internal/cmd.versionGoVersion=$(shell $(GO) version | awk '{ print $$3 }') \
			-X $(PKG_IMPORT_PATH)/internal/cmd.versionTag=$(shell git describe --tag)"

deps:
	$(GO) mod tidy && $(GO) mod vendor

# Run a security scanner over the source code. This Makefile won't install the
# scanner binary for you, so check out the gosec README for instructions:
# https://github.com/securego/gosec
#
# If necessary, specify the path to the built binary with the GOSEC env var.
gosec:
	$(GOSEC) $(FLAGS) $(SRC_PATHS)

vet:
	$(GO) vet $(FLAGS) $(SRC_PATHS)

# Specify packages to test with P variable. Example:
# make test P='entity repo'
#
# Specify test flags with FLAGS variable. Example:
# make test FLAGS='-v -count=1 -failfast'
test: P ?= ...
test: pkgpath=$(foreach pkg,$(P),$(shell echo ./internal/$(pkg)))
test: _testdirs
test:
	$(GO) test $(pkgpath) $(FLAGS)

_testdirs:
	mkdir -pv $(TEST_DIR) && chmod -v 700 $(TEST_DIR)
