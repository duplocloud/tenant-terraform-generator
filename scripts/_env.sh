#!/bin/bash -eu

# Discover both the AWS Account ID and DuploCloud Default Tenant ID if they are not set.
# Historically the documentation has had the user set these values so allow the user to specifiy the values and use the user provided values.
# To support both commercial and GovCloud regions, first grab the region of the default infrastructure which can then be used in `with_aws`
_VALID_AWS_REGION=$(duplo_api /adminproxy/GetInfrastructureConfig/default | jq -r '.Region')
AWS_ACCOUNT_ID=$([ -z "${aws_account_id:-}" ] && with_aws aws --region "$_VALID_AWS_REGION" sts get-caller-identity | jq -r '.Account' || echo "$aws_account_id")
duplo_default_tenant_id=$([ -z "${tenant_id:-}" ] && duplo_api /adminproxy/GetTenantNames | jq -r '.[] | select(.AccountName == "default") | .TenantId' || echo "$tenant_id")

duplo_host="$duplo_host"
duplo_token="$duplo_token"

backend="-backend-config=bucket=duplo-tfstate-${AWS_ACCOUNT_ID} -backend-config=dynamodb_table=duplo-tfstate-${AWS_ACCOUNT_ID}-lock"

# Test required environment variables
for key in duplo_token duplo_host
do
  eval "[ -n \"\${${key}:-}\" ]" || die "error: $key: environment variable missing or empty"
done

export duplo_host duplo_token duplo_default_tenant_id backend AWS_ACCOUNT_ID
