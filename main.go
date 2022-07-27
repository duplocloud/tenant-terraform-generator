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

	tenantConfig, err := client.TenantGet(config.TenantId)
	if err != nil {
		log.Fatalf("error getting tenant from duplo: %s", err)
	}
	if tenantConfig == nil {
		log.Fatalf("Tenant not found: Tenant Id - %s ", config.TenantId)
	}
	config.TenantName = tenantConfig.AccountName

	log.Println("[TRACE] <====== Initialize target directory with customer name and tenant id. =====>")
	initTargetDir(config)
	log.Printf("[TRACE] Config ==> %+v\n", config)
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

	certArn := os.Getenv("cert_arn")
	if len(custName) == 0 {
		err := fmt.Errorf("Error - Please provide \"%s\" as env variable.", "cert_arn")
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
		CertArn:              certArn,
	}
}

func initTargetDir(config *common.Config) {
	config.TFCodePath = filepath.Join("target", config.CustomerName, config.TenantName, "terraform")
	tenantProject := filepath.Join(config.TFCodePath, config.TenantProject)
	err := os.RemoveAll(tenantProject)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(tenantProject, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	config.AdminTenantDir = tenantProject

	awsServicesProject := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	err = os.RemoveAll(awsServicesProject)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(awsServicesProject, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	config.AwsServicesDir = awsServicesProject

	appProject := filepath.Join(config.TFCodePath, config.AppProject)
	err = os.RemoveAll(appProject)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(appProject, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	config.AppDir = appProject

	scriptsPath := filepath.Join("target", config.CustomerName, config.TenantName, "scripts")
	err = os.RemoveAll(scriptsPath)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(scriptsPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = duplosdk.CopyDirectory("./scripts", scriptsPath)
	if err != nil {
		log.Fatal(err)
	}
	err = duplosdk.Copy(".gitignore", filepath.Join("target", config.CustomerName, config.TenantName, ".gitignore"))
	if err != nil {
		log.Fatal(err)
	}
	err = duplosdk.Copy(".envrc", filepath.Join("target", config.CustomerName, config.TenantName, ".envrc"))
	if err != nil {
		log.Fatal(err)
	}

}

func startTFGeneration(config *common.Config, client *duplosdk.Client) {
	// var tf *tfexec.Terraform
	providerGen := &common.Provider{}
	providerGen.Generate(config, client)

	// if config.GenerateTfState {
	// 	tf := tfInit(config, config.AdminTenantDir)
	// 	tfNewWorkspace(config, tf)
	// }

	log.Println("[TRACE] <====== Start TF generation for tenant project. =====>")
	// Register New TF generator for Tenant Project
	tenantGeneratorList := []tfgenerator.Generator{
		&tenant.Tenant{},
		&tenant.TenantSGRule{},
	}
	if config.S3Backend {
		tenantGeneratorList = append(tenantGeneratorList, &tenant.TenantBackend{})
	}

	starTFGenerationForProject(config, client, tenantGeneratorList, config.AdminTenantDir)
	validateAndFormatTfCode(config.AdminTenantDir)
	log.Println("[TRACE] <====== End TF generation for tenant project. =====>")

	log.Println("[TRACE] <====== Start TF generation for aws services project. =====>")
	// Register New TF generator for AWS Services project
	awsServcesGeneratorList := []tfgenerator.Generator{
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
	}
	if config.S3Backend {
		awsServcesGeneratorList = append(awsServcesGeneratorList, &awsservices.AwsServicesBackend{})
	}
	starTFGenerationForProject(config, client, awsServcesGeneratorList, config.AwsServicesDir)
	validateAndFormatTfCode(config.AwsServicesDir)
	log.Println("[TRACE] <====== End TF generation for aws services project. =====>")

	log.Println("[TRACE] <====== Start TF generation for app project. =====>")
	// Register New TF generator for App Services project
	appGeneratorList := []tfgenerator.Generator{
		&app.AppMain{},
		&app.Services{},
		&app.ECS{},
		&app.K8sConfig{},
		&app.K8sSecret{},
	}
	if config.S3Backend {
		appGeneratorList = append(appGeneratorList, &app.AppBackend{})
	}
	starTFGenerationForProject(config, client, appGeneratorList, config.AppDir)
	validateAndFormatTfCode(config.AppDir)
	log.Println("[TRACE] <====== End TF generation for app project. =====>")

	// if config.GenerateTfState && config.S3Backend {
	// 	tf = tfInit(config, config.AppDir)
	// 	tfDeleteWorkspace(config, tf)
	// }

}

func starTFGenerationForProject(config *common.Config, client *duplosdk.Client, generatorList []tfgenerator.Generator, targetLocation string) {

	tfContext := common.TFContext{
		TargetLocation: targetLocation,
		InputVars:      []common.VarConfig{},
		OutputVars:     []common.OutputVarConfig{},
	}

	// 1. Generate Duplo TF resources.
	for _, g := range generatorList {
		c, err := g.Generate(config, client)
		if err != nil {
			log.Fatalf("error running admin tenant tf generation: %s", err)
		}
		if c != nil {
			if len(c.InputVars) > 0 {
				tfContext.InputVars = append(tfContext.InputVars, c.InputVars...)
			}
			if len(c.OutputVars) > 0 {
				tfContext.OutputVars = append(tfContext.OutputVars, c.OutputVars...)
			}
			if len(c.ImportConfigs) > 0 {
				tfContext.ImportConfigs = append(tfContext.ImportConfigs, c.ImportConfigs...)
			}
		}
	}
	// 2. Generate input vars.
	if len(tfContext.InputVars) > 0 {
		varsGenerator := common.Vars{
			TargetLocation: tfContext.TargetLocation,
			Vars:           tfContext.InputVars,
		}
		varsGenerator.Generate()
	}
	// 3. Generate output vars.
	if len(tfContext.OutputVars) > 0 {
		outVarsGenerator := common.OutputVars{
			TargetLocation: tfContext.TargetLocation,
			OutputVars:     tfContext.OutputVars,
		}
		outVarsGenerator.Generate()
	}
	// 4. Import all resources
	if config.GenerateTfState && len(tfContext.ImportConfigs) > 0 {
		tfInitializer := common.TfInitializer{
			WorkingDir: targetLocation,
			Config:     config,
		}
		tf := tfInitializer.InitWithWorkspace()
		importer := &common.Importer{}
		for _, ic := range tfContext.ImportConfigs {
			//importer.Import(config, &ic)
			importer.ImportWithoutInit(config, &ic, tf)
		}
		//tfInitializer.DeleteWorkspace(config, tf)
	}
}

func validateAndFormatTfCode(tfDir string) {
	log.Printf("[TRACE] Validation and formatting of terraform code generated at %s is started.", tfDir)
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("0.14.11")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
	tf, err := tfexec.NewTerraform(tfDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}
	log.Printf("[TRACE] Validation of terraform code generated at %s is started.", tfDir)
	_, err = tf.Validate(context.Background())
	if err != nil {
		log.Fatalf("error running terraform validate: %s", err)
	}
	log.Printf("[TRACE] Validation of terraform code generated at %s is done.", tfDir)
	log.Printf("[TRACE] Formatting of terraform code generated at %s is started.", tfDir)
	err = tf.FormatWrite(context.Background())
	if err != nil {
		log.Fatalf("error running terraform format: %s", err)
	}
	log.Printf("[TRACE] Formatting of terraform code generated at %s is done.", tfDir)
	log.Printf("[TRACE] Validation and formatting of terraform code generated at %s is done.", tfDir)
}
