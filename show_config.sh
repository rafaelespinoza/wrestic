#!/usr/bin/env bash

set -eu -o pipefail

declare CURR_SCRIPT SCRIPT_DIR WRESTIC_BIN
CURR_SCRIPT="$(realpath "${0}")"
SCRIPT_DIR=$(dirname "${CURR_SCRIPT}")
WRESTIC_BIN="${WRESTIC_BIN:-"${SCRIPT_DIR}/bin/wrestic"}"
readonly CURR_SCRIPT SCRIPT_DIR WRESTIC_BIN

declare -r _jq_url='https://jqlang.github.io/jq'
declare -r _miller_url='https://miller.readthedocs.io'
declare -r usage="Usage: ${0} [-Bbds] [-h]

Description:
	Display high level details of each datastore as a table.
	Which datastore is it, where is data sourced from, and where are
	backups destined to.

	There are several perspectives:
	* (b) basic overview of store names and destination names
	* (d) destination names, paths and store names
	* (s) source paths and store names
	* (B) combines Both source and destinations into 1 view

Dependencies:
	* jq: ${_jq_url}
	* miller (mlr): ${_miller_url}

Flags:
	-B show Both source and destination in same view
	-b show basic overview
	-d show destination path details
	-s show source path details
	-h show help menu

Examples:
	# show basic overview only
		$ ${0} -b
	# show destination path and source path views
		$ ${0} -ds
	# show Both source and destination
		$ ${0} -B
	# halp
		$ ${0} -h
"

declare -r jq_store_dest_names='{
	Store: .Name,
	DestNames: (.Destinations | keys | join(", ")),
}'

declare -r jq_store_src_paths='{
	Store: .Name,
	SourcePath: .Sources[]?.Path,
}'

# shellcheck disable=SC2016 # allow this jq expression to use variables
declare -r jq_store_dest_paths='. as $parent |
	(
		.Destinations? | map({
			Store: $parent.Name,
			DestName: .Name,
			DestPath: .Path,
		})
	)[]'

# Note the use of the iterator `.[]`, which is meant to "unroll" array
# fields `SourcePath` and `Dest`, so that individual objects from those
# fields (`Sources` and `Destinations` respectively) end up with their
# own line in final output table.
declare -r jq_omni_view='{
	Store: .Name,
	SourcePath: (.Sources | map(.Path))[],
	Dest: (.Destinations | map({ Name, Path }))[]
}'

function _run() {
	local -r jq_expr="${1:?missing jq_expr}"

	if [[ ! -x ${WRESTIC_BIN} ]]; then
		echo >&2 "it appears that WRESTIC_BIN (${WRESTIC_BIN}) is not executable.
See README for build instructions.
TLDR: $ make build"
		return 1
	fi
	if ! command -v jq >&/dev/null; then
		echo >&2 "this script requires jq; see ${_jq_url}"
		return 1
	fi
	if ! command -v mlr >&/dev/null; then
		echo >&2 "this script requires miller (mlr); see ${_miller_url}"
		return 1
	fi

	"${WRESTIC_BIN}" config show --format json --merge=false |
		jq "${jq_expr}" |
		mlr --ijsonl --opprint --barred cat
}

function main() {
	if [[ "${#}" -eq 0 ]] || [[ "${1}" == '-h' ]]; then
		echo -n "${usage}" >&2
		return 0
	fi

	local show_omni_view=0
	local show_basic_overview=0
	local show_source_paths=0
	local show_dest_paths=0

	while getopts "Bbdsh" opt; do
		case "${opt}" in
			B)
				show_omni_view=1
				;;
			b)
				show_basic_overview=1
				;;
			d)
				show_dest_paths=1
				;;
			s)
				show_source_paths=1
				;;
			h)
				echo -n "${usage}" >&2
				return 0
				;;
			*)
				echo -n "${usage}" >&2
				return 1
				;;
		esac
	done

	[[ "${show_basic_overview}" -eq 1 ]] && _run "${jq_store_dest_names}"
	[[ "${show_source_paths}" -eq 1 ]] && _run "${jq_store_src_paths}"
	[[ "${show_dest_paths}" -eq 1 ]] && _run "${jq_store_dest_paths}"
	[[ "${show_omni_view}" -eq 1 ]] && _run "${jq_omni_view}"
}

main "$@"
