#!/bin/bash -eu

# shellcheck disable=SC1091   # VS code can't follow the below file
source "$(dirname "${BASH_SOURCE[0]}")/_util.sh"

tenant="$1" ; shift
case "$tenant" in
default|compliance)
  err "Tenant cannot be named: $tenant" 1>&2
  usage
esac

# Which project to run.
selection="${1:-}"
[ $# -eq 0 ] || shift
if [ -n "$selection" ] && ! [ -d "terraform/$selection" ]; then
  err "No such project: $selection" 1>&2
  usage
fi

# Load environment and utility programs.
export acct tenant

# shellcheck disable=SC1091   # VS code can't follow the below file
source "$(dirname "${BASH_SOURCE[0]}")/_env.sh"

# Utility function to run "terraform destroy" with proper arguments, and clean state.
function tf_destroy() {
  local project="$1" ; shift

  [ -z "${backend:-}" ] && die "internal error: backend should have been configured by _env.sh"

  # Skip projects that are not selected.
  if [ -n "$selection" ] && [ "$selection" != "$project" ]; then
    return 0;
  fi

  # Determine the terraform workspace.
  local ws="$tenant"

  local tf_args=( -auto-approve -input=false "$@" )

  local varfile="configs/$ws/$project.tfvars.json"
  [ -f "$varfile" ] && tf_args=( "${tf_args[@]}" "-var-file=../../$varfile" )

  echo "Project: $project"

  # shellcheck disable=SC2086    # NOTE: we want word splitting
  (cd "terraform/$project" &&
      tf_init $backend &&
      if tf workspace select "$ws"; then
        tf destroy "${tf_args[@]}" && tf workspace select default && tf workspace delete "$ws"
      fi)
}

tf_output() {
    local project="$1" ; shift

    # Determine the terraform workspace.
    local ws="$tenant"

    # shellcheck disable=SC2086    # NOTE: we want word splitting
    (cd "terraform/$project" &&
        tf_init $backend 1>&2 &&
        ( tf workspace select "$ws" 1>&2 || tf workspace new "$ws" 1>&2 ) &&
        tf output -json )
}

if [ "$selection" = "admin-infra" ]; then
  case "$tenant" in
  prod*|nonprod*)
    tf_destroy admin-infra "$@"
    ;;
  *)
    die "$tenant: not an expected infrastructure name"
    ;;
  esac
else
  tf_destroy <--app--> "$@"
  tf_destroy <--aws-services--> "$@"
  tf_destroy <--admin-tenant--> "$@"
fi
