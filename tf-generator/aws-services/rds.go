package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const RDS_VAR_PREFIX = "rds_"

type Rds struct {
}

func (r *Rds) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.RdsInstanceList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== RDS TF generation started. =====>")
		for _, rds := range *list {
			shortName := rds.Identifier[len("duplo"):len(rds.Identifier)]
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo RDS Instance : %s", rds.Identifier)

			varFullPrefix := RDS_VAR_PREFIX + resourceName + "_"
			inputVars := generateRdsVars(rds, varFullPrefix)
			tfContext.InputVars = append(tfContext.InputVars, inputVars...)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "rds-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_rds_instance resource
			rdsBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_rds_instance",
					resourceName})
			rdsBody := rdsBlock.Body()
			rdsBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// rdsBody.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			name := shortName + "-${local.tenant_name}"
			rdsNameTokens := hclwrite.Tokens{
				{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenIdent, Bytes: []byte(name)},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
			}
			rdsBody.SetAttributeRaw("name", rdsNameTokens)
			// rdsBody.SetAttributeValue("name",
			// 	cty.StringVal(shortName+"-"+config.TenantName))
			rdsBody.SetAttributeValue("engine",
				cty.NumberIntVal(int64(rds.Engine)))
			rdsBody.SetAttributeTraversal("engine_version", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "engine_version",
				},
			})
			rdsBody.SetAttributeTraversal("size", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "size",
				},
			})

			if len(rds.SnapshotID) > 0 {
				rdsBody.SetAttributeValue("snapshot_id",
					cty.StringVal(rds.SnapshotID))
			} else {
				rdsBody.SetAttributeTraversal("master_username", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "var",
					},
					hcl.TraverseAttr{
						Name: varFullPrefix + "master_username",
					},
				})
			}
			rdsBody.SetAttributeValue("master_password",
				cty.StringVal(rds.MasterPassword))
			// if len(rds.DBParameterGroupName) > 0 {
			// 	rdsBody.SetAttributeValue("parameter_group_name",
			// 		cty.StringVal(rds.DBParameterGroupName))
			// }
			// rdsBody.SetAttributeValue("store_details_in_secret_manager",
			// 	cty.BoolVal(rds.StoreDetailsInSecretManager))
			rdsBody.SetAttributeTraversal("encrypt_storage", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "encrypt_storage",
				},
			})
			rdsBody.SetAttributeValue("enable_logging",
				cty.BoolVal(rds.EnableLogging))
			rdsBody.SetAttributeValue("multi_az",
				cty.BoolVal(rds.MultiAZ))
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo RDS instance : %s", rds.Identifier)

			outVars := generateRdsOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_rds_instance." + resourceName,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/RDSDBInstance/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== RDS TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateRdsVars(duplo duplosdk.DuploRdsInstance, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	var1 := common.VarConfig{
		Name:       prefix + "engine_version",
		DefaultVal: duplo.EngineVersion,
		TypeVal:    "string",
	}
	varConfigs["engine_version"] = var1

	var2 := common.VarConfig{
		Name:       prefix + "size",
		DefaultVal: duplo.SizeEx,
		TypeVal:    "string",
	}
	varConfigs["size"] = var2

	var3 := common.VarConfig{
		Name:       prefix + "encrypt_storage",
		DefaultVal: strconv.FormatBool(duplo.EncryptStorage),
		TypeVal:    "bool",
	}
	varConfigs["encrypt_storage"] = var3

	var4 := common.VarConfig{
		Name:       prefix + "master_password",
		DefaultVal: duplo.MasterPassword,
		TypeVal:    "string",
	}
	varConfigs["master_password"] = var4

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}

func generateRdsOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "fullname",
		ActualVal:     "duplocloud_rds_instance." + resourceName + ".identifier",
		DescVal:       "The full name of the RDS instance.",
		RootTraversal: true,
	}
	outVarConfigs["fullname"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_rds_instance." + resourceName + ".arn",
		DescVal:       "The ARN of the RDS instance.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "endpoint",
		ActualVal:     "duplocloud_rds_instance." + resourceName + ".endpoint",
		DescVal:       "The endpoint of the RDS instance.",
		RootTraversal: true,
	}
	outVarConfigs["endpoint"] = var3

	var4 := common.OutputVarConfig{
		Name:          prefix + "host",
		ActualVal:     "duplocloud_rds_instance." + resourceName + ".host",
		DescVal:       "The DNS hostname of the RDS instance.",
		RootTraversal: true,
	}
	outVarConfigs["host"] = var4

	var5 := common.OutputVarConfig{
		Name:          prefix + "port",
		ActualVal:     "duplocloud_rds_instance." + resourceName + ".port",
		DescVal:       "The listening port of the RDS instance.",
		RootTraversal: true,
	}
	outVarConfigs["port"] = var5

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
