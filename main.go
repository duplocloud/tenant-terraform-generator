package main

//ReadMe : https://dev.to/pdcommunity/write-terraform-files-in-go-with-hclwrite-2e1j
import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	tfgenerator "tenant-terraform-generator/tf-generator"
	"tenant-terraform-generator/tf-generator/common"
)

func init() {
	fmt.Println("This will get called on main initialization")
}

func main() {

	// Initialize duplo client and config
	log.Println("[TRACE] <====== Initialize duplo client and config. =====>")
	validator := common.EnvVarValidator{}
	config, err := validator.Validate()
	if err != nil {
		os.Exit(1)
	}
	client, err := duplosdk.NewClient(config.DuploHost, config.DuploToken)
	if err != nil {
		err = fmt.Errorf("error while creating duplo client %s", err)
		log.Printf("[TRACE] - %s", err)
		os.Exit(1)
	}

	sslNoVerify := os.Getenv("ssl_no_verify")
	if len(sslNoVerify) != 0 {
		client.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	log.Println("[TRACE] <====== Initialized duplo client and config. =====>")

	tenantConfig, err := client.GetTenantByNameForUser(config.TenantName)
	if err != nil {
		log.Fatalf("error getting tenant from duplo: %s", err)
	}
	if tenantConfig == nil {
		log.Fatalf("Tenant not found: Tenant Name - %s ", config.TenantName)
	}
	config.TenantId = tenantConfig.TenantID
	accountID, err := client.TenantGetAwsAccountID(config.TenantId)
	if err != nil {
		log.Fatalf("error getting aws account id from duplo: %s", err)
	}
	config.AccountID = accountID

	log.Printf("[TRACE] Config ==> %+v\n", config)

	tfGeneratorService := tfgenerator.TfGeneratorService{}

	err = tfGeneratorService.PreProcess(config, client)
	if err != nil {
		log.Fatalf("error while pre processing: %s", err)
	}
	err = tfGeneratorService.StartTFGeneration(config, client)
	if err != nil {
		log.Fatalf("error while generating terraform: %s", err)
	}
	err = tfGeneratorService.PostProcess(config, client)
	if err != nil {
		log.Fatalf("error while post processing: %s", err)
	}
	log.Printf("[TRACE] |==========================================================================|")
	log.Printf("[TRACE] Terraform projects are generated at - %s", filepath.Join("./target", config.CustomerName, config.TenantName))
	log.Printf("[TRACE] |==========================================================================|")
}
