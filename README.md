# Terraform code generation for a DuploCloud Tenant

This utility provides a way to export the terraform code that represents the infrastructure deployed in a DuploCloud Tenant. This is often very useful in order to:
- Generate and persist DuploCloud Terraform IaC which can be version controlled in the future.
- Clone a new Tenant based on an already existing Tenant. 

## Prerequisite

1. Install [Go](https://go.dev/doc/install) 1.15 or later.
2. Install [make](https://www.gnu.org/software/make) tool.
3. Install [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli) version greater than or equals to `v0.14.11`
4. Install [jq](https://stedolan.github.io/jq/download/)
5. Following environment variables to be exported in the shell while running this projects.

```shell
# Required Vars
export customer_name="duplo-masp"
export tenant_name="test"
export cert_arn="arn:aws:acm:us-west-2:128329325849:certificate/1234567890-aaaa-bbbb-ccc-66e7dcd609e1"
export duplo_host="https://msp.duplocloud.net"
export duplo_token="xxx-xxxxx-xxxxxxxx"
export AWS_RUNNER="duplo-admin"
export aws_account_id="1234567890"
```
You can optionally pass following environment variables.

```shell
# Optional Vars
export tenant_project="admin-tenant" # Project name for tenant, Default is admin-tenant.
export aws_services_project="aws-services" #  Project name for tenant, Default is aws-services.
export app_project="app" #  Project name for tenant, Default is app.
export skip_admin_tenant="true" # Whether to skip tf generation for admin-tenant, Default is false.
export skip_aws_services="true" # Whether to skip tf generation for aws_services, Default is false.
export skip_app="true" # Whether to skip tf generation for app, Default is false.
export tf_version=0.14.11  # Terraform version to be used, Default is 0.14.11.
export validate_tf="false" # Whether to validate generated tf code, Default is true.
export generate_tf_state="false" # Whether to import generated tf resources, Default is false. 
                                 # If true please use 'AWS_PROFILE' environment variable, This is required for s3 backend.
```
6. Set **DISABLETFSTATERESOURCECREATION** key as false in DuploCloud. Please contact the DuploCloud team for assistance.

## How to run this project to export DuploCloud Provider terraform code?

- Clone this repository.

- Prepare environment variables and export within the shell as mentioned above.

- Run using  following command

  ```shell
  make run
  ```

- **Output** : target folder is created along with customer name and tenant name as mentioned in the environment variables. This folder will contain all terraform projects as mentioned below.
  
    ```
    ├── target                   # Target folder for terraform code
    │   ├── customer-name        # Folder with customer name
    │     ├── tenant-name        # Folder with tenant name
    │       ├── scripts          # Wrapper scripts to plan, apply and destroy terarform infrastructure.
    │       ├── terraform        # Terraform code generated using this utility.
    │          ├── admin-tenant  # Terraform code for tenant and tenant related resources.
    │          ├── aws-services  # Terraform code for AWS services.
    │          ├── app           # Terraform code for DuploCloud services and ECS.
    ```

  - **Project : admin-tenant** This projects manages creation of DuploCloud tenant and tenant related resources.
  - **Project : aws-services** This project manages data services like Redis, RDS, Kafka, S3 buckets, Cloudfront, EMR, Elastic Search inside DuploCloud.
  - **Project : app** This project manages DuploCloud services like EKS, ECS etc.

## Following DuploCloud resources are supported.
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

## How to use generated terraform code to create a new DuploCloud Tenant, and its resources?

### Prerequisite
1. Following environment variables to be exported in the shell while running this terraform projects.
```shell
export AWS_RUNNER=duplo-admin
export tenant_id="XXXXXXXXXXXXXXXXXXXXXXXXX"
export duplo_host="https://msp.duplocloud.net"
export duplo_token="<duplo-auth-token>"
export aws_account_id="1234567890"
```
2. To run terraform projects you must be in `tenant-name` directory.
```shell
cd target/customer-name/tenant-name
```

### Wrapper Scripts

There are scripts to manage terraform infrastructure. Which will helps to create a DuploCloud infrastructure based on tenant.

- scripts/plan.sh
- scripts/apply.sh
- scripts/destroy.sh

#### Arguments to run the scripts.

- **First Argument:** Name of the new tenant to be created.
- **Second Argument:** Terraform project name. Valid values are - `admin-tenant`, `aws-services` and `app`.

### Terraform Projects

This infrastructure is divided into terraform sub projects which manages different managed DuploCloud resources like tenant, AWS services like Redis, RDS, Kafka, S3 buckets, Elastic Search and DuploCloud services which are containerized.

- **Project - admin-tenant**

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

- **Project - aws-services**

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

### External Configurations (Manual Step)

End user can pass external configurations like RDS instance type, ES version, Docker image version etc. while running these projects.

#### How to create configuration file

- Create configuration folder inside config folder with the name of tenant, Example, Tenant Name is `dev01` then `dev01` folder is created inside config folder first.

- File - **admin-tenant.tfvars.json**
  - This file is used while running **Project - admin-tenant**, You can create file **admin-tenant.tfvars.json** and pass required configuration.
- File - **aws-services.tfvars.json**
  - This file is used while running **Project - aws-services**, You can create file **aws-services.tfvars.json** and pass required configuration.

- File - **app.tfvars.json**
  - This file is used while running **Project - app**, You can create file **app.tfvars.json** and pass required configuration.

    ```
    ├── target                                  # Target folder for terraform code
    │   ├── customer-name                       # Folder with customer name
    │     ├── tenant-name                       # Folder with tenant name
    │       ├── config                          # External configuration folder.
    │          ├── dev01                        # Tenant specific config folder.
    │             ├── admin-tenant.tfvars.json  # admin-tenant project variables.
    │             ├── aws-services.tfvars.json  # aws-services project variables.
    │             ├── app.tfvars.json           # app project variables.
    ```  

### Contributing

If you have enhancements, improvements or fixes, we would love to have your contributions.

#### Developing

If you want to add support of new resource, Follow the steps below.

- Identify project([admin-tenant](./tf-generator/tenant), [aws-services](./tf-generator/aws-services) or [app](./tf-generator/app)), Add generator file for new resource like [redis.go](./tf-generator/aws-services/redis.go)
- Once resource file is added, Register same resource in [generator-registry.go](./tf-generator/generator-registry.go) like **&awsservices.Redis{}**