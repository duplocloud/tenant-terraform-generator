package main

//ReadMe : https://dev.to/pdcommunity/write-terraform-files-in-go-with-hclwrite-2e1j
import (
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
		generateTfStateBool, err := strconv.ParseBool(os.Getenv("generate_tf_state"))
		if err != nil {
			err = fmt.Errorf("Error while reading generate_tf_state from env vars %s", err)
			log.Printf("[TRACE] - %s", err)
			os.Exit(1)
		}
		generateTfState = generateTfStateBool
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
	awsServicesProject := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	err = os.RemoveAll(awsServicesProject)
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(awsServicesProject, os.ModePerm)
	appProject := filepath.Join(config.TFCodePath, config.AppProject)
	err = os.RemoveAll(appProject)
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(appProject, os.ModePerm)
	scriptsPath := filepath.Join("target", config.CustomerName, "scripts")
	err = os.RemoveAll(scriptsPath)
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(scriptsPath, os.ModePerm)
	duplosdk.CopyDirectory("./scripts", scriptsPath)
}

func startTFGeneration(config *common.Config, client *duplosdk.Client) {
	// Register all tf generators here in the list, Sequence matters.
	generatorList := []tfgenerator.Generator{
		&common.Provider{},
		&tenant.Tenant{},
		&tenant.TenantBackend{},
		&awsservices.AwsServicesBackend{},
		&awsservices.Hosts{},
		&awsservices.ASG{},
		&awsservices.Rds{},
		&awsservices.Redis{},
		&awsservices.Kafka{},
		&awsservices.S3Bucket{},
		&app.AppBackend{},
		&app.Services{},
		&app.ECS{},
	}

	for _, g := range generatorList {
		g.Generate(config, client)
	}
	// Generate provider terraform for admin-tenant, aws-services and app
	// providerTFGenerator := &tfgenerator.Provider{}
	// providerTFGenerator.Generate(config, client)

	// Generate admin-tenant terraform
	// tenantTFGenerator := &tfgenerator.Tenant{}
	// tenantTFGenerator.Generate(config, client)

	//tenantBackendGenerator := &tfgenerator.TenantBackend{}
	//tenantBackendGenerator.SetNext(tenantTFGenerator)

	// Generate aws-services terraform
	// awsServicesBackendTFGenerator := &tfgenerator.AwsServicesBackend{}
	// awsServicesBackendTFGenerator.Generate(config, client)

	// hostsTFGenerator := &tfgenerator.Hosts{}
	// hostsTFGenerator.Generate(config, client)

	// rdsTFGenerator := &tfgenerator.Rds{}
	// rdsTFGenerator.Generate(config, client)

	// redisTFGenerator := &tfgenerator.Redis{}
	// redisTFGenerator.Generate(config, client)

	// kafkaTFGenerator := &tfgenerator.Kafka{}
	// kafkaTFGenerator.Generate(config, client)

	// Generate app terraform

	// appBackendTFGenerator := &tfgenerator.AppBackend{}
	// appBackendTFGenerator.Generate(config, client)

	// servicesTFGenerator := &tfgenerator.Services{}
	// servicesTFGenerator.Generate(config, client)

}

// func tfRunTest(config *tfgenerator.Config) {
// 	installer := &releases.ExactVersion{
// 		Product: product.Terraform,
// 		Version: version.Must(version.NewVersion("0.14.11")),
// 	}

// 	execPath, err := installer.Install(context.Background())
// 	if err != nil {
// 		log.Fatalf("error installing Terraform: %s", err)
// 	}
// 	workingDir := filepath.Join("target", config.CustomerName, config.TenantProject)
// 	tf, err := tfexec.NewTerraform(workingDir, execPath)
// 	if err != nil {
// 		log.Fatalf("error running NewTerraform: %s", err)
// 	}

// 	err = tf.Init(context.Background(), tfexec.Upgrade(true))
// 	if err != nil {
// 		log.Fatalf("error running Init: %s", err)
// 	}
// 	err = tf.Import(context.Background(), "duplocloud_tenant.tenant", "v2/admin/TenantV2/"+config.TenantId)
// 	if err != nil {
// 		log.Fatalf("error running Import: %s", err)
// 	}
// 	state, err := tf.Show(context.Background())
// 	if err != nil {
// 		log.Fatalf("error running Show: %s", err)
// 	}
// 	stateJson, err := json.Marshal(state.Values)
// 	fmt.Println(string(stateJson)) // "0.1"
// }
