package tenant

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type Tenant struct {
}

func (t *Tenant) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.TenantProject)

	log.Println("[TRACE] <====== Tenant TF generation started. =====>")
	duplo, clientErr := client.TenantGet(config.TenantId)
	//Get tenant from duplo
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	infraConfig, clientErr := client.InfrastructureGetConfig(duplo.PlanID)
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	//1. ==========================================================================================
	// Generate variables
	log.Printf("[TRACE] Genrating vars for Tenant Name : %s", duplo.AccountName)
	inputVars := generateVars(duplo, infraConfig)
	tfContext.InputVars = inputVars
	log.Printf("[TRACE] Vars genrated for Tenant Name : %s", duplo.AccountName)

	//2. ==========================================================================================
	// Generate resoueces
	log.Printf("[TRACE] Tenant Name : %s", duplo.AccountName)
	// This is needed for all other resources.
	config.TenantName = duplo.AccountName
	// create new empty hcl file object
	hclFile := hclwrite.NewEmptyFile()

	// create new file on system
	path := filepath.Join(workingDir, "main.tf")
	tfFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// initialize the body of the new file object
	rootBody := hclFile.Body()

	localsBlock := rootBody.AppendNewBlock("locals",
		nil)
	localsBlockBody := localsBlock.Body()

	localsBlockBody.SetAttributeTraversal("region", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "region",
		},
	})
	localsBlockBody.SetAttributeTraversal("plan_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "infra_name",
		},
	})
	localsBlockBody.SetAttributeTraversal("cert_arn", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "cert_arn",
		},
	})
	localsBlockBody.SetAttributeTraversal("tenant_name", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "terraform",
		},
		hcl.TraverseAttr{
			Name: "workspace",
		},
	})
	rootBody.AppendNewline()

	// Add duplocloud_tenant resource
	tenant := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_tenant",
			"tenant"})
	tenantBody := tenant.Body()
	tenantBody.SetAttributeTraversal("account_name", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "local",
		},
		hcl.TraverseAttr{
			Name: "tenant_name",
		},
	})
	tenantBody.SetAttributeTraversal("plan_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "local",
		},
		hcl.TraverseAttr{
			Name: "plan_id",
		},
	})
	tenantBody.SetAttributeValue("allow_deletion",
		cty.BoolVal(true))
	rootBody.AppendNewline()
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
	outVars := generateOutputVars(workingDir)
	tfContext.OutputVars = outVars
	log.Printf("[TRACE] Output vars generated for Tenant Name : %s", duplo.AccountName)

	// 4. ==========================================================================================
	// Import all created resources.
	if config.GenerateTfState {
		importConfigs := []common.ImportConfig{}
		importConfigs = append(importConfigs, common.ImportConfig{
			ResourceAddress: "duplocloud_tenant.tenant",
			ResourceId:      "v2/admin/TenantV2/" + config.TenantId,
			WorkingDir:      workingDir,
		}, common.ImportConfig{
			ResourceAddress: "duplocloud_tenant_config.tenant-config",
			ResourceId:      config.TenantId,
			WorkingDir:      workingDir,
		})
		tfContext.ImportConfigs = importConfigs
		// importer := &common.Importer{}
		// importer.Import(config, &common.ImportConfig{
		// 	ResourceAddress: "duplocloud_tenant.tenant",
		// 	ResourceId:      "v2/admin/TenantV2/" + config.TenantId,
		// 	WorkingDir:      workingDir,
		// })
		// importer.Import(config, &common.ImportConfig{
		// 	ResourceAddress: "duplocloud_tenant_config.tenant-config",
		// 	ResourceId:      config.TenantId,
		// 	WorkingDir:      workingDir,
		// })
	}
	return &tfContext, nil
}

func generateVars(duplo *duplosdk.DuploTenant, infraConfig *duplosdk.DuploInfrastructureConfig) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	regionVar := common.VarConfig{
		Name:       "region",
		DefaultVal: infraConfig.Region,
		TypeVal:    "string",
	}
	varConfigs["region"] = regionVar

	infraVar := common.VarConfig{
		Name:       "infra_name",
		DefaultVal: duplo.PlanID,
		TypeVal:    "string",
	}
	varConfigs["infra_name"] = infraVar

	certVar := common.VarConfig{
		Name:       "cert_arn",
		DefaultVal: "null",
		TypeVal:    "string",
	}
	varConfigs["cert_arn"] = certVar

	tenantNameVar := common.VarConfig{
		Name:       "tenant_name",
		DefaultVal: duplo.AccountName,
		TypeVal:    "string",
	}
	varConfigs["tenant_name"] = tenantNameVar

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	// varsGenerator := common.Vars{
	// 	TargetLocation: workingDir,
	// 	Vars:           vars,
	// }
	// varsGenerator.Generate()
	return vars
}

func generateOutputVars(workingDir string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	tenantNameVar := common.OutputVarConfig{
		Name:          "tenant_name",
		ActualVal:     "duplocloud_tenant.tenant.account_name",
		DescVal:       "The tenant name",
		RootTraversal: true,
	}
	outVarConfigs["tenant_name"] = tenantNameVar

	tenantIdVar := common.OutputVarConfig{
		Name:          "tenant_id",
		ActualVal:     "duplocloud_tenant.tenant.tenant_id",
		DescVal:       "The tenant ID",
		RootTraversal: true,
	}
	outVarConfigs["tenant_id"] = tenantIdVar

	certVar := common.OutputVarConfig{
		Name:          "cert_arn",
		ActualVal:     "var.cert_arn",
		DescVal:       "The duplo plan certificate arn.",
		RootTraversal: true,
	}
	outVarConfigs["cert_arn"] = certVar

	regionVar := common.OutputVarConfig{
		Name:          "region",
		ActualVal:     "var.region",
		DescVal:       "The duplo plan region.",
		RootTraversal: true,
	}
	outVarConfigs["region"] = regionVar

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}

	// outVarsGenerator := common.OutputVars{
	// 	TargetLocation: workingDir,
	// 	OutputVars:     outVars,
	// }
	// outVarsGenerator.Generate()
	return outVars
}
