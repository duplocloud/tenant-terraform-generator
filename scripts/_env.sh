#!/bin/bash -eu

AWS_ACCOUNT_ID="$aws_account_id"
duplo_default_tenant_id="$tenant_id"
duplo_host="$duplo_host"
duplo_token="$duplo_token"

backend="-backend-config=bucket=duplo-tfstate-${AWS_ACCOUNT_ID} -backend-config=dynamodb_table=duplo-tfstate-${AWS_ACCOUNT_ID}-lock"

# Test required environment variables
for key in duplo_token duplo_host
do
  eval "[ -n \"\${${key}:-}\" ]" || die "error: $key: environment variable missing or empty"
done

export duplo_host duplo_token duplo_default_tenant_id backend AWS_ACCOUNT_ID
