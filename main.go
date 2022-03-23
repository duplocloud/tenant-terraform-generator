package main

//ReadMe : https://dev.to/pdcommunity/write-terraform-files-in-go-with-hclwrite-2e1j
import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"tenant-terraform-generator/duplosdk"
	tfgenerator "tenant-terraform-generator/tf-generator"
	"tenant-terraform-generator/tf-generator/app"
	awsservices "tenant-terraform-generator/tf-generator/aws-services"
	"tenant-terraform-generator/tf-generator/common"
	"tenant-terraform-generator/tf-generator/tenant"

	"github.com/hashicorp/go-version"

	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

func init() {
	fmt.Println("This will get called on main initialization")
}

func main() {

	// Initialize duplo client and config
	log.Println("[TRACE] <====== Initialize duplo client and config. =====>")
	client := validateAndGetDuploClient()
	config := validateAndGetConfig()
	log.Println("[TRACE] <====== Initialized duplo client and config. =====>")

	log.Println("[TRACE] <====== Initialize target directory with customer name and tenant id. =====>")
	initTargetDir(config)
	log.Println("[TRACE] <====== Initialized target directory with customer name and tenant id. =====>")
	// Chain of responsiblity started.
	// Provider --> Tenant --> Hosts --> Services --> ...
	startTFGeneration(config, client)

}

func validateAndGetDuploClient() *duplosdk.Client {
	host := os.Getenv("duplo_host")
	if len(host) == 0 {
		err := fmt.Errorf("Error - Please provide \"%s\" as env variable.", "duplo_host")
		log.Printf("[TRACE] - %s", err)
		os.Exit(1)
	}
	token := os.Getenv("duplo_token")
	if len(token) == 0 {
		err := fmt.Errorf("Error - Please provide \"%s\" as env variable.", "duplo_token")
		log.Printf("[TRACE] - %s", err)
		os.Exit(1)
	}
	c, err := duplosdk.NewClient(host, token)
	if err != nil {
		err = fmt.Errorf("Error while creating duplo client %s", err)
		log.Printf("[TRACE] - %s", err)
		os.Exit(1)
	}

	sslNoVerify := os.Getenv("ssl_no_verify")
	if len(sslNoVerify) != 0 {
		c.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return c
}

func validateAndGetConfig() *common.Config {

	tenantId := os.Getenv("tenant_id")
	if len(tenantId) == 0 {
		err := fmt.Errorf("Error - Please provide \"%s\" as env variable.", "tenant_id")
		log.Printf("[TRACE] - %s", err)
		os.Exit(1)
	}

	awsAccountId := os.Getenv("aws_account_id")
	if len(tenantId) == 0 {
		err := fmt.Errorf("Error - Please provide \"%s\" as env variable.", "aws_account_id")
		log.Printf("[TRACE] - %s", err)
		os.Exit(1)
	}

	custName := os.Getenv("customer_name")
	if len(custName) == 0 {
		err := fmt.Errorf("Error - Please provide \"%s\" as env variable.", "customer_name")
		log.Printf("[TRACE] - %s", err)
		os.Exit(1)
	}

	duploProviderVersion := os.Getenv("duplo_provider_version")
	if len(custName) == 0 {
		duploProviderVersion = "0.7.0"
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
			err = fmt.Errorf("Error while reading generate_tf_state from env vars %s", err)
			log.Printf("[TRACE] - %s", err)
			os.Exit(1)
		}
		generateTfState = generateTfStateBool
	}

	s3Backend := false
	s3BackendStr := os.Getenv("s3_backend")
	if len(s3BackendStr) == 0 {
		s3Backend = false
	} else {
		s3BackendBool, err := strconv.ParseBool(s3BackendStr)
		if err != nil {
			err = fmt.Errorf("Error while reading s3_backend from env vars %s", err)
			log.Printf("[TRACE] - %s", err)
			os.Exit(1)
		}
		s3Backend = s3BackendBool
	}

	return &common.Config{
		TenantId:             tenantId,
		CustomerName:         custName,
		DuploProviderVersion: duploProviderVersion,
		TenantProject:        tenantProject,
		AwsServicesProject:   awsServicesProject,
		AppProject:           appProject,
		GenerateTfState:      generateTfState,
		AccountID:            awsAccountId,
		S3Backend:            s3Backend,
	}
}

func initTargetDir(config *common.Config) {
	config.TFCodePath = filepath.Join("target", config.CustomerName, "terraform")
	tenantProject := filepath.Join(config.TFCodePath, config.TenantProject)
	err := os.RemoveAll(tenantProject)
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(tenantProject, os.ModePerm)
	config.AdminTenantDir = tenantProject

	awsServicesProject := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	err = os.RemoveAll(awsServicesProject)
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(awsServicesProject, os.ModePerm)
	config.AwsServicesDir = awsServicesProject

	appProject := filepath.Join(config.TFCodePath, config.AppProject)
	err = os.RemoveAll(appProject)
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(appProject, os.ModePerm)
	config.AppDir = appProject

	scriptsPath := filepath.Join("target", config.CustomerName, "scripts")
	err = os.RemoveAll(scriptsPath)
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(scriptsPath, os.ModePerm)
	duplosdk.CopyDirectory("./scripts", scriptsPath)
}

func startTFGeneration(config *common.Config, client *duplosdk.Client) {
	var tf *tfexec.Terraform
	providerGen := &common.Provider{}
	providerGen.Generate(config, client)

	tenantConfig, err := client.TenantGet(config.TenantId)
	if err != nil {
		log.Fatalf("error getting tenant from duplo: %s", err)
	}
	config.TenantName = tenantConfig.AccountName
	// if config.GenerateTfState {
	// 	tf := tfInit(config, config.AdminTenantDir)
	// 	tfNewWorkspace(config, tf)
	// }

	tenantGeneratorList := []tfgenerator.Generator{
		&tenant.Tenant{},
	}
	if config.S3Backend {
		tenantGeneratorList = append(tenantGeneratorList, &tenant.TenantBackend{})
	}

	for _, g := range tenantGeneratorList {
		g.Generate(config, client)
	}

	//tf = tfInit(config, config.AwsServicesDir)
	awsServcesGeneratorList := []tfgenerator.Generator{
		&awsservices.AwsServicesMain{},
		&awsservices.Hosts{},
		&awsservices.ASG{},
		&awsservices.Rds{},
		&awsservices.Redis{},
		&awsservices.Kafka{},
		&awsservices.S3Bucket{},
	}
	if config.S3Backend {
		awsServcesGeneratorList = append(awsServcesGeneratorList, &awsservices.AwsServicesBackend{})
	}
	for _, g := range awsServcesGeneratorList {
		g.Generate(config, client)
	}

	//tf = tfInit(config, config.AppDir)
	appGeneratorList := []tfgenerator.Generator{
		&app.Services{},
		&app.ECS{},
	}
	if config.S3Backend {
		appGeneratorList = append(appGeneratorList, &app.AppBackend{})
	}
	for _, g := range appGeneratorList {
		g.Generate(config, client)
	}
	if config.GenerateTfState {
		tf = tfInit(config, config.AppDir)
		tfDeleteWorkspace(config, tf)
	}

}

func tfInit(config *common.Config, workingDir string) *tfexec.Terraform {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("0.14.11")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}
	if config.S3Backend {
		err = tf.Init(context.Background(), tfexec.Upgrade(true), tfexec.BackendConfig("bucket=duplo-tfstate-"+config.AccountID), tfexec.BackendConfig("dynamodb_table=duplo-tfstate-"+config.AccountID+"-lock"))
	} else {
		err = tf.Init(context.Background(), tfexec.Upgrade(true))
	}

	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}
	return tf
}

func tfNewWorkspace(config *common.Config, tf *tfexec.Terraform) {
	err := tf.WorkspaceNew(context.Background(), config.TenantName)
	if err != nil {
		log.Fatalf("error running tf workspace new: %s", err)
	}
}

func tfDeleteWorkspace(config *common.Config, tf *tfexec.Terraform) {
	err := tf.WorkspaceDelete(context.Background(), config.TenantName)
	if err != nil {
		log.Fatalf("error running tf workspace delete: %s", err)
	}
}
