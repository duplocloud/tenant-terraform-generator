#!/bin/bash -eu

# Load environment and utility programs.
export tenant
# shellcheck disable=SC1090    # VS code won't enable following the files.
source "$(dirname "${BASH_SOURCE[0]}")/_util.sh"
# shellcheck disable=SC1090    # VS code won't enable following the files.
source "$(dirname "${BASH_SOURCE[0]}")/_env.sh"

duplo_api "$@"
