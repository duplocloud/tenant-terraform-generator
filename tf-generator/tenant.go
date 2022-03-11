package tfgenerator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type Tenant struct {
}

func (t *Tenant) Generate(config *Config, client *duplosdk.Client) {
	workingDir := filepath.Join("target", config.CustomerName, config.TenantProject)

	log.Println("[TRACE] <====== Tenant TF generation started. =====>")
	duplo, clientErr := client.TenantGet(config.TenantId)
	//Get tenant from duplo
	if clientErr != nil {
		fmt.Println(clientErr)
		return
	}
	infraConfig, clientErr := client.InfrastructureGetConfig(duplo.PlanID)
	if clientErr != nil {
		fmt.Println(clientErr)
		return
	}

	//1. ==========================================================================================
	// Generate variables
	log.Printf("[TRACE] Genrating vars for Tenant Name : %s", duplo.AccountName)
	generateVars(workingDir, duplo, infraConfig)
	log.Printf("[TRACE] Vars genrated for Tenant Name : %s", duplo.AccountName)

	//2. ==========================================================================================
	// Generate resoueces
	log.Printf("[TRACE] Tenant Name : %s", duplo.AccountName)
	// This is needed for all other resources.
	config.TenantName = duplo.AccountName
	// create new empty hcl file object
	hclFile := hclwrite.NewEmptyFile()

	// create new file on system
	path := filepath.Join(workingDir, "tenant.tf")
	tfFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	// initialize the body of the new file object
	rootBody := hclFile.Body()

	// Add duplocloud_tenant resource
	tenant := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_tenant",
			"tenant"})
	tenantBody := tenant.Body()
	tenantBody.SetAttributeTraversal("account_name", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "tenant_name",
		},
	})
	tenantBody.SetAttributeTraversal("plan_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "infra_name",
		},
	})
	tenantBody.SetAttributeValue("allow_deletion",
		cty.BoolVal(true))

	// Add duplocloud_tenant_config resource

	tenantConfig := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_tenant_config",
			"tenant-config"})
	tenantConfigBody := tenantConfig.Body()
	// tenantConfigBody.SetAttributeValue("tenant_id",
	// 	cty.StringVal(config.TenantId))
	tenantConfigBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "duplocloud_tenant.tenant",
		},
		hcl.TraverseAttr{
			Name: "tenant_id",
		},
	})
	settingBlock := tenantConfigBody.AppendNewBlock("setting",
		nil)
	settingBlockBody := settingBlock.Body()
	settingBlockBody.SetAttributeValue("key",
		cty.StringVal("delete_protection"))
	settingBlockBody.SetAttributeValue("value",
		cty.StringVal("true"))
	rootBody.AppendNewline()

	fmt.Printf("%s", hclFile.Bytes())
	tfFile.Write(hclFile.Bytes())
	log.Println("[TRACE] <====== Tenant TF generation done. =====>")

	// 3. ==========================================================================================
	// Generate outputs
	log.Printf("[TRACE] Genrating output vars for Tenant Name : %s", duplo.AccountName)
	generateOutputVars(workingDir)
	log.Printf("[TRACE] Output vars generated for Tenant Name : %s", duplo.AccountName)

	// 4. ==========================================================================================
	// Import all created resources.
	if config.GenerateTfState {
		importer := &Importer{}
		importer.Import(config, &ImportConfig{
			resourceAddress: "duplocloud_tenant.tenant",
			resourceId:      "v2/admin/TenantV2/" + config.TenantId,
			workingDir:      workingDir,
		})
		importer.Import(config, &ImportConfig{
			resourceAddress: "duplocloud_tenant_config.tenant-config",
			resourceId:      config.TenantId,
			workingDir:      workingDir,
		})
	}
}

func generateVars(workingDir string, duplo *duplosdk.DuploTenant, infraConfig *duplosdk.DuploInfrastructureConfig) {
	varConfigs := make(map[string]VarConfig)

	regionVar := VarConfig{
		Name:       "region",
		DefaultVal: infraConfig.Region,
		TypeVal:    "string",
	}
	varConfigs["region"] = regionVar

	infraVar := VarConfig{
		Name:       "infra_name",
		DefaultVal: duplo.PlanID,
		TypeVal:    "string",
	}
	varConfigs["infra_name"] = infraVar

	certVar := VarConfig{
		Name:       "cert_arn",
		DefaultVal: "null",
		TypeVal:    "string",
	}
	varConfigs["cert_arn"] = certVar

	tenantNameVar := VarConfig{
		Name:       "tenant_name",
		DefaultVal: duplo.AccountName,
		TypeVal:    "string",
	}
	varConfigs["tenant_name"] = tenantNameVar

	vars := make([]VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	varsGenerator := Vars{
		targetLocation: workingDir,
		vars:           vars,
	}
	varsGenerator.Generate()
}

func generateOutputVars(workingDir string) {
	outVarConfigs := make(map[string]OutputVarConfig)

	tenantNameVar := OutputVarConfig{
		Name:          "tenant_name",
		ActualVal:     "duplocloud_tenant.tenant.account_name",
		DescVal:       "The tenant name",
		RootTraversal: true,
	}
	outVarConfigs["tenant_name"] = tenantNameVar

	tenantIdVar := OutputVarConfig{
		Name:          "tenant_id",
		ActualVal:     "duplocloud_tenant.tenant.tenant_id",
		DescVal:       "The tenant ID",
		RootTraversal: true,
	}
	outVarConfigs["tenant_id"] = tenantIdVar

	certVar := OutputVarConfig{
		Name:          "cert_arn",
		ActualVal:     "var.cert_arn",
		DescVal:       "The duplo plan certificate arn.",
		RootTraversal: true,
	}
	outVarConfigs["cert_arn"] = certVar

	outVars := make([]OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}

	outVarsGenerator := OutputVars{
		targetLocation: workingDir,
		outputVars:     outVars,
	}
	outVarsGenerator.Generate()
}
