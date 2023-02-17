package common

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type IValidator interface {
	Validate() (*Config, error)
}

type EnvVarValidator struct {
}

func (envVar *EnvVarValidator) Validate() (*Config, error) {
	host := os.Getenv("duplo_host")
	if len(host) == 0 {
		err := fmt.Errorf("error - Please provide \"%s\" as env variable", "duplo_host")
		log.Printf("[TRACE] - %s", err)
		return nil, err
	}
	token := os.Getenv("duplo_token")
	if len(token) == 0 {
		err := fmt.Errorf("error - Please provide \"%s\" as env variable", "duplo_token")
		log.Printf("[TRACE] - %s", err)
		return nil, err
	}

	tenantName := os.Getenv("tenant_name")
	if len(tenantName) == 0 {
		err := fmt.Errorf("error - please provide \"%s\" as env variable", "tenant_name")
		log.Printf("[TRACE] - %s", err)
		return nil, err
	}
	custName := os.Getenv("customer_name")
	if len(custName) == 0 {
		err := fmt.Errorf("error - please provide \"%s\" as env variable", "customer_name")
		log.Printf("[TRACE] - %s", err)
		return nil, err
	}

	// certArn := os.Getenv("cert_arn")
	// if len(certArn) == 0 {
	// 	err := fmt.Errorf("error - please provide \"%s\" as env variable", "cert_arn")
	// 	log.Printf("[TRACE] - %s", err)
	// 	return nil, err
	// }

	duploProviderVersion := os.Getenv("duplo_provider_version")
	if len(duploProviderVersion) == 0 {
		duploProviderVersion = "0.9.0"
	}

	tfVersion := os.Getenv("tf_version")
	if len(tfVersion) == 0 {
		tfVersion = "0.14.11"
	}

	tenantProject := os.Getenv("tenant_project")
	if len(tenantProject) == 0 {
		tenantProject = "admin-tenant"
	}

	awsServicesProject := os.Getenv("aws_services_project")
	if len(awsServicesProject) == 0 {
		awsServicesProject = "aws-services"
	}

	appProject := os.Getenv("app_project")
	if len(appProject) == 0 {
		appProject = "app"
	}

	generateTfState := false

	generateTfStateStr := os.Getenv("generate_tf_state")
	if len(generateTfStateStr) == 0 {
		generateTfState = false
	} else {
		generateTfStateBool, err := strconv.ParseBool(generateTfStateStr)
		if err != nil {
			err = fmt.Errorf("error while reading generate_tf_state from env vars %s", err)
			log.Printf("[TRACE] - %s", err)
			return nil, err
		}
		generateTfState = generateTfStateBool
	}

	validateTf := true
	validateTfStr := os.Getenv("validate_tf")
	if len(validateTfStr) == 0 {
		validateTf = true
	} else {
		validateTf, _ = strconv.ParseBool(generateTfStateStr)
	}

	s3Backend := true
	s3BackendStr := os.Getenv("s3_backend")
	if len(s3BackendStr) == 0 {
		s3Backend = true
	} else {
		s3BackendBool, err := strconv.ParseBool(s3BackendStr)
		if err != nil {
			err = fmt.Errorf("error while reading s3_backend from env vars %s", err)
			log.Printf("[TRACE] - %s", err)
			return nil, err
		}
		s3Backend = s3BackendBool
	}

	skipTenant := false
	skipTenantStr := os.Getenv("skip_admin_tenant")
	if len(skipTenantStr) == 0 {
		skipTenant = false
	} else {
		skipTenant, _ = strconv.ParseBool(skipTenantStr)
	}

	skipAwsServices := false
	skipAwsServicesStr := os.Getenv("skip_aws_services")
	if len(skipAwsServicesStr) == 0 {
		skipAwsServices = false
	} else {
		skipAwsServices, _ = strconv.ParseBool(skipAwsServicesStr)
	}

	skipApp := false
	skipAppStr := os.Getenv("skip_app")
	if len(skipAppStr) == 0 {
		skipApp = false
	} else {
		skipApp, _ = strconv.ParseBool(skipAppStr)
	}

	return &Config{
		DuploHost:            host,
		DuploToken:           token,
		TenantName:           tenantName,
		CustomerName:         custName,
		DuploProviderVersion: duploProviderVersion,
		TenantProject:        tenantProject,
		AwsServicesProject:   awsServicesProject,
		AppProject:           appProject,
		GenerateTfState:      generateTfState,
		S3Backend:            s3Backend,
		CertArn:              certArn,
		ValidateTf:           validateTf,
		TFVersion:            tfVersion,
		SkipAdminTenant:      skipTenant,
		SkipAwsServices:      skipAwsServices,
		SkipApp:              skipApp,
	}, nil
}
