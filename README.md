# Terraform code generation for a DuploCloud Tenant

This utility provides a way to export the terraform code that represents the infrastructure deployed in a DuploCloud Tenant. This is often very useful in order to:
- Generate and persist DuploCloud Terraform IaC which can be version controlled in the future.
- Clone a new Tenant based on an already existing Tenant. 

## Prerequisite

1. Install [Go](https://go.dev/doc/install) 1.15 or later.
2. Install [make](https://www.gnu.org/software/make) tool.
3. Install [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli) version greater than or equals to `v1.4.2`
4. Install [jq](https://stedolan.github.io/jq/download/)
5. Following environment variables to be exported in the shell while running this projects.

```shell
# Required Vars
export tenant_name="test"
export duplo_host="https://msp.duplocloud.net"
export duplo_token="xxx-xxxxx-xxxxxxxx"
```
You can optionally pass following environment variables.

```shell
# Optional Vars
export duplo_provider_version="0.9.0" # DuploCloud provider version to be used..
export tenant_project="admin-tenant" # Project name for tenant, Default is admin-tenant.
export aws_services_project="aws-services" #  Project name for tenant, Default is aws-services.
export app_project="app" #  Project name for tenant, Default is app.
export skip_admin_tenant="true" # Whether to skip tf generation for admin-tenant, Default is false.
export skip_aws_services="true" # Whether to skip tf generation for aws_services, Default is false.
export skip_app="true" # Whether to skip tf generation for app, Default is false.
export tf_version=v1.4.2  # Terraform version to be used, Default is v1.4.2.
export validate_tf="false" # Whether to validate generated tf code, Default is true.
export enable_k8s_secret_placeholder="false" # Whether to put 'replace-me' placeholder for k8s secret instead of actual value.
export k8s_secret_placeholder="replace-me" # Placeholder for k8s secret when enable_k8s_secret_placeholder is true.
export generate_tf_state="false" # Whether to import generated tf resources, Default is false. 
                                 # If true please use 'AWS_PROFILE' environment variable, This is required for s3 backend.
```
6. Set **DisableTfStateResourceCreation** key as false in Administrator ➝ System Settings ➝ System Configs in DuploCloud UI. Please contact the DuploCloud team for assistance.

## How to run this project to export DuploCloud Provider terraform code?

- Clone this repository.

- Prepare environment variables and export within the shell as mentioned above.

- Run using  following command

  ```shell
  make run
  ```

- **Output** : target folder is created along with customer name and tenant name as mentioned in the environment variables. This folder will contain all terraform projects as mentioned below.
  
    ```
    ├── target                # Target folder for terraform code
    │   ├── customer-name       # Folder with customer name
    |      ├── config           # Folder contains tfvars file for workspace
    |          ├── infra01                      # Infra related to tenant specific config folder.
    │             ├── infra.tfvars.json         # infra project variables
    |          ├── dev01                        # Tenant specific config folder.
    │             ├── tenant.tfvars.json        # tenant project variables.
    │             ├── services.tfvars.json      # services project variables.
    │             ├── app.tfvars.json           # app project variables.
    │      ├── scripts          # Wrapper scripts to plan, apply and destroy terarform infrastructure.
    │      ├── terraform        # Terraform code generated using this utility.
    |         ├── infra           # Terraform code for infra and plan related resources.
    │         ├── tenant          # Terraform code for tenant and tenant related resources.
    │         ├── services        # Terraform code for AWS services.
    │         ├── app             # Terraform code for DuploCloud services and ECS.
    ```

  - **Project : tenant** This projects manages creation of DuploCloud tenant and tenant related resources.
  - **Project : services** This project manages data services like Redis, RDS, Kafka, S3 buckets, Cloudfront, EMR, Elastic Search inside DuploCloud.
  - **Project : app** This project manages DuploCloud services like EKS, ECS etc.
  - **Project : infra** This project manages DuploCloud services like infrastructure, plan_certificates etc.

## Following DuploCloud resources are supported.
   - `duplocloud_infrastructure`
   - `duplocloud_infrastructure_setting`
   - `duplocloud_infrastructure_subnet`
   - `duplocloud_plan_certificates`
   - `duplocloud_plan_configs`
   - `duplocloud_plan_image`
   - `duplocloud_plan_kms_v2`
   - `duplocloud_plan_settings`
   - `duplocloud_plan_waf_v2`
   - `duplocloud_tenant`
   - `duplocloud_tenant_network_security_rule`
   - `duplocloud_asg_profile`
   - `duplocloud_aws_host`
   - `duplocloud_aws_kafka_cluster`
   - `duplocloud_rds_instance`
   - `duplocloud_ecache_instance`
   - `duplocloud_s3_bucket`
   - `duplocloud_aws_sns_topic`
   - `duplocloud_aws_sqs_queue`
   - `duplocloud_duplo_service`
   - `duplocloud_duplo_service_lbconfigs`
   - `duplocloud_duplo_service_params`
   - `duplocloud_ecs_task_definition`
   - `duplocloud_ecs_service`
   - `duplocloud_aws_mwaa_environment`
   - `duplocloud_aws_elasticsearch`
   - `duplocloud_k8_secret`
   - `duplocloud_k8_config_map`
   - `duplocloud_k8_ingress`
   - `duplocloud_k8_secret_provider_class`
   - `duplocloud_aws_ssm_parameter`
   - `duplocloud_aws_load_balancer`
   - `duplocloud_aws_load_balancer_listener`
   - `duplocloud_aws_api_gateway_integration`
   - `duplocloud_aws_ecr_repository`
   - `duplocloud_aws_cloudfront_distribution`
   - `duplocloud_aws_lambda_function`
   - `duplocloud_aws_lambda_permission`
   - `duplocloud_aws_dynamodb_table_v2`
   - `duplocloud_byoh`
   - `duplocloud_emr_cluster`
   - `duplocloud_aws_cloudwatch_metric_alarm`
   - `duplocloud_aws_cloudwatch_event_rule`
   - `duplocloud_aws_cloudwatch_event_target`
   - `duplocloud_aws_target_group_attributes`
   - `duplocloud_aws_lb_listener_rule`
   - `duplocloud_aws_batch_scheduling_policy`
   - `duplocloud_aws_batch_job_definition`
   - `duplocloud_aws_batch_compute_environment`
   - `duplocloud_aws_batch_job_queue`
   - `duplocloud_aws_timestreamwrite_database`
   - `duplocloud_aws_timestreamwrite_table`
   - `duplocloud_k8s_cron_job`
   - `duplocloud_k8s_job`
   
## How to use generated terraform code to create a new DuploCloud Tenant, and its resources?

### Prerequisite
1. Following environment variables to be exported in the shell while running this terraform projects.
```shell
export tenant_name="tenant for which app, infra, services and tenant resource need to be fetched"
export duplo_host="https://msp.duplocloud.net"
export duplo_token="<duplo-auth-token>"
```
2. To run terraform projects you must be in `customer-name` directory.
```shell
cd target/customer-name
```

### Wrapper Scripts

There are scripts to manage terraform infrastructure. Which will helps to create a DuploCloud infrastructure based on tenant.

- scripts/plan.sh
- scripts/apply.sh
- scripts/destroy.sh

#### Arguments to run the scripts.

- **First Argument:** Name of the new tenant to be created.
- **Second Argument:** Terraform project name. Valid values are - `infra`, `tenant`, `services` and `app`.

### Terraform Projects

This infrastructure is divided into terraform sub projects which manages different managed DuploCloud resources like tenant, AWS services like Redis, RDS, Kafka, S3 buckets, Elastic Search and DuploCloud services which are containerized.

- **Project - tenant**

  This projects manages DuploCloud infrastructure and tenant, Run this project using following command using tenant-name and project name.

  - Dry-run

    - ```shell
       scripts/plan.sh <tenant-name> admin-tenant
      ```

  - Actual Deployment

    - ```shell
      scripts/apply.sh <tenant-name> admin-tenant
      ```

  - Destroy created infrastructure

    - ```shell
      scripts/destroy.sh <tenant-name> admin-tenant
      ```
  **Note** : Please provide required variables `infra_name` and `cert_arn` in `vars.tf`.

- **Project - services**

  This project manages AWS services like Redis, RDS, Kafka, S3 buckets, Elastic Search, etc. inside DuploCloud.

  - Dry-run

    - ```shell
       scripts/plan.sh <tenant-name> aws-services
      ```
  - Actual Deployment

    - ```shell
      scripts/apply.sh <tenant-name> aws-services
      ```
  - Destroy created infrastructure

    - ```shell
      scripts/destroy.sh <tenant-name> aws-services
      ```

- **Project - app**

  This project manages containerized applications inside DuploCloud like EKS services, ECS, Docker Native service etc.

  - Dry-run

    - ```shell
       scripts/plan.sh <tenant-name> app
      ```

  - Actual Deployment

    - ```shell
      scripts/apply.sh <tenant-name> app
      ```

  - Destroy created infrastructure

    - ```shell
      scripts/destroy.sh <tenant-name> app
      ```

- **Project - infra**

  This project manages infrastructure and plan related to tenant name passed via environment variable.

  - Dry-run

    - ```shell
       scripts/plan.sh <infra-name> infra
      ```

  - Actual Deployment

    - ```shell
      scripts/apply.sh <infra-name> infra
      ```

  - Destroy created infrastructure

    - ```shell
      scripts/destroy.sh <infra-name> infra
      ```

### External Configurations (Manual Step)

End user can pass external configurations like RDS instance type, ES version, Docker image version etc. while running these projects.


> Note: Both json and tfvar Terraform file extensions are supported.  See Terraform [documentation](https://developer.hashicorp.com/terraform/language/values/variables#variable-definitions-tfvars-files) for more details about the structure of each file type.

### Contributing

If you have enhancements, improvements or fixes, we would love to have your contributions.

#### Developing

If you want to add support of new resource, Follow the steps below.

- Identify project([admin-tenant](./tf-generator/tenant), [aws-services](./tf-generator/aws-services) or [app](./tf-generator/app)), Add generator file for new resource like [redis.go](./tf-generator/aws-services/redis.go)
- Once resource file is added, Register same resource in [generator-registry.go](./tf-generator/generator-registry.go) like **&awsservices.Redis{}**
