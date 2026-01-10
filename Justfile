#!/usr/bin/env -S just -f

GO := "go"
GOSEC := "gosec"
BIN_DIR := "bin"
MAIN := BIN_DIR / "wrestic"
PKG_IMPORT_PATH := "github.com/rafaelespinoza/wrestic"
TESTDATA_SCRIPT := justfile_directory() / "testdata.sh"
SHOWCONF_SCRIPT := justfile_directory() / "show_config.sh"
SRC_PATHS := ". ./internal/..."

# list recipes
@default:
    just -f {{ justfile() }} --list --unsorted

# compile binary with versioning info
[group('build')]
build:
    #!/usr/bin/env bash
    set -eu -o pipefail
    mkdir -pv {{ BIN_DIR }}
    branch_name=$(git rev-parse --abbrev-ref HEAD)
    build_time=$(date --utc +%FT%T%z)
    commit_hash=$(git rev-parse --short=7 HEAD)
    go_os_arch=$({{ GO }} version | awk '{ print $4 }' | tr '/' '_')
    go_version=$({{ GO }} version | awk '{ print $3 }')
    tag=$(git describe --tag --always)
    _base_var_path='{{ PKG_IMPORT_PATH }}/internal/cmd'
    ldflags="
    -X ${_base_var_path}.versionBranchName=${branch_name}
    -X ${_base_var_path}.versionBuildTime=${build_time}
    -X ${_base_var_path}.versionCommitHash=${commit_hash}
    -X ${_base_var_path}.versionGoOSArch=${go_os_arch}
    -X ${_base_var_path}.versionGoVersion=${go_version}
    -X ${_base_var_path}.versionTag=${tag}"
    {{ GO }} build -v -o {{ MAIN }} -ldflags="${ldflags}"

# get module dependencies, tidy them up
[group('build')]
deps:
    {{ GO }} mod tidy

# This Justfile won't install the scanner binary for you, so check out the gosec
# README for instructions: https://github.com/securego/gosec.
# If necessary, specify the path to the built binary with the GOSEC variable.
#
# run a security scanner over the source code
[group('static')]
gosec *args:
	{{ GOSEC }} {{ args }} {{ SRC_PATHS }}

# examine source code for suspicious constructs
[group('static')]
vet *args:
    {{ GO }} vet {{ args }} {{ SRC_PATHS }}

# run tests
[group('test')]
test *args:
    {{ GO }} test {{ args }} {{ SRC_PATHS }}

# create testdata
[group('test')]
init-testdata: (_testdata "init")

# remove testdata
[group('test')]
clean-testdata: (_testdata "teardown-all-testdata")

_testdata CMD:
    {{ TESTDATA_SCRIPT }} {{ CMD }}

# quick view of configured datastores and destinations
show_config *args:
    {{ SHOWCONF_SCRIPT }} {{ args }}
