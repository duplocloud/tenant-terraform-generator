package tfgenerator

import (
	"tenant-terraform-generator/tf-generator/app"
	awsservices "tenant-terraform-generator/tf-generator/aws-services"
	"tenant-terraform-generator/tf-generator/tenant"
)

var TenantGenerators = []Generator{
	&tenant.Tenant{},
	&tenant.TenantSGRule{},
}

var AWSServicesGenerators = []Generator{
	&awsservices.AwsServicesMain{},
	&awsservices.Hosts{},
	&awsservices.ASG{},
	&awsservices.Rds{},
	&awsservices.Redis{},
	&awsservices.Kafka{},
	&awsservices.S3Bucket{},
	&awsservices.SQS{},
	&awsservices.SNS{},
	&awsservices.MWAA{},
	&awsservices.ES{},
	&awsservices.SsmParams{},
	&awsservices.LoadBalancer{},
	&awsservices.ApiGatewayIntegration{},
	&awsservices.CFD{},
	&awsservices.LambdaFunction{},
	&awsservices.DynamoDB{},
	&awsservices.BYOH{},
	&awsservices.EMR{},
	&awsservices.CloudwatchMetrics{},
	&awsservices.ECR{},
}

var AppGenerators = []Generator{
	&app.AppMain{},
	&app.Services{},
	&app.ECS{},
	&app.K8sConfig{},
	&app.K8sSecret{},
	&app.K8sIngress{},
	&app.K8sSecretProviderClass{},
}
