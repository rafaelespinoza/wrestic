#!/usr/bin/env sh

set -eu

CURRENT_DIR="$(dirname "$(realpath "$0")")"
readonly CURRENT_DIR

# these binaries are expected to be available.
readonly RESTIC="restic" # see https://github.com/restic/restic

# commonly-used paths.
readonly WORKDIR_ROOT=/tmp/wrestic_test/testdata
readonly SECRETS_DIR="${WORKDIR_ROOT}/secrets"
readonly REPOS_ROOT="${WORKDIR_ROOT}/repos"
readonly SRCDATA_ROOT="${WORKDIR_ROOT}/srcdata"

readonly SECRET_FILE_A="${SECRETS_DIR}/a"
readonly SECRET_FILE_B="${SECRETS_DIR}/b"

restic_init () {
	reponame="${1:?missing reponame}"
	keyfile="${2:?missing keyfile}"

	repopath="${REPOS_ROOT}/${reponame}"

	"${RESTIC}" -r "${repopath}" --password-command "cat ${keyfile}" "${subcmd}" init
}

initialize () {
	mkdir -pv "${SECRETS_DIR}" "${REPOS_ROOT}" "${SRCDATA_ROOT}" && chmod -v 0700 "${WORKDIR_ROOT}"
	cp -iv "${CURRENT_DIR}/internal/config/testdata/wrestic.toml" "${WORKDIR_ROOT}/wrestic.toml"

	# set up password configuration
	echo "secret_test_a" > "${SECRET_FILE_A}"
	echo "secret_test_b" > "${SECRET_FILE_B}"

	# set up restic repositories
	restic_init alfa "${SECRET_FILE_A}"
	restic_init bravo "${SECRET_FILE_A}"
	restic_init charlie "${SECRET_FILE_B}"

	# populate source directories with dummy data.
	mkdir -pv "${SRCDATA_ROOT}/foo"
	mkdir -pv "${SRCDATA_ROOT}/bar"
	mkdir -pv "${SRCDATA_ROOT}/qux"
	touch "${SRCDATA_ROOT}/foo/README.md"
	touch "${SRCDATA_ROOT}/bar/README.md"
	touch "${SRCDATA_ROOT}/qux/README.md"
}

usage () {
	>&2 echo "testdata subcmd

Description:
	Automate initialization of testdata.

Subcommands:
	init

	teardown-all-testdata
"
}

main () {
	if [ "$#" -eq 0 ] || [ "$1" = "-h" ]; then
		usage
		return 1
	fi

	readonly subcmd="${1:?missing subcmd}"; shift

	case "${subcmd}" in
		init )
			initialize
			;;
		teardown-all-testdata )
			rm -rf "${WORKDIR_ROOT}"
			;;
		* )
			usage
			return 1
			;;
	esac
}

main "$@"
