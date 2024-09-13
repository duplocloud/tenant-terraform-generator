package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	log.Printf("[TRACE] RDS generate")
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

			// if len(rds.ClusterIdentifier) > 0 || len(rds.ReplicationSourceIdentifier) > 0 {
			if len(rds.DuploRdsRole) > 0 && strings.ToLower(rds.DuploRdsRole) == "reader" {
				varFullPrefix := RDS_VAR_PREFIX

				inputVars := generateRdsRRVars(rds, RDS_VAR_PREFIX)
				tfContext.InputVars = append(tfContext.InputVars, inputVars...)

				rrBlock := rootBody.AppendNewBlock("resource",
					[]string{"duplocloud_rds_read_replica",
						resourceName})
				rrBody := rrBlock.Body()
				rrBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "local",
					},
					hcl.TraverseAttr{
						Name: "tenant_id",
					},
				})

				clusterIdentifier := rds.ReplicationSourceIdentifier
				if len(clusterIdentifier) == 0 {
					clusterIdentifier = rds.ClusterIdentifier
				}
				rrBody.SetAttributeTraversal("cluster_identifier", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "duplocloud_rds_instance." + strings.TrimPrefix(strings.TrimSuffix(common.GetResourceName(clusterIdentifier), "_cluster"), "duplo"),
					},
					hcl.TraverseAttr{
						Name: "cluster_identifier",
					},
				})

				name := shortName
				rdsNameTokens := hclwrite.Tokens{
					{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
					{Type: hclsyntax.TokenIdent, Bytes: []byte(name)},
					{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
				}
				rrBody.SetAttributeRaw("name", rdsNameTokens)

				rrBody.SetAttributeTraversal("size", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "var",
					},
					hcl.TraverseAttr{
						Name: varFullPrefix + "size",
					},
				})
				if rds.V2ScalingConfiguration != nil {
					scalingBody := rrBody.AppendNewBlock("v2_scaling_configuration", nil).Body()
					scalingBody.SetAttributeTraversal("min_capacity", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "var",
						},
						hcl.TraverseAttr{
							Name: varFullPrefix + "scaling_conf_min_capacity",
						},
					})
					scalingBody.SetAttributeTraversal("max_capacity", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "var",
						},
						hcl.TraverseAttr{
							Name: varFullPrefix + "scaling_conf_max_capacity",
						},
					})
				}
				if rds.EnablePerformanceInsights {
					inightBody := rrBody.AppendNewBlock("performance_insights", nil).Body()
					inightBody.SetAttributeTraversal("enabled", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "var",
						},
						hcl.TraverseAttr{
							Name: varFullPrefix + "performance_insights_enabled",
						},
					})
					inightBody.SetAttributeTraversal("retention_period", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "var",
						},
						hcl.TraverseAttr{
							Name: varFullPrefix + "performance_insights_retention_period",
						},
					})
					inightBody.SetAttributeTraversal("kms_key_id", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "var",
						},
						hcl.TraverseAttr{
							Name: varFullPrefix + "performance_insights_kms_key_id",
						},
					})
				}
				lifecycleBody := rrBody.AppendNewBlock("lifecycle", nil).Body()
				lifecycle := common.StringSliceToListVal([]string{"engine_version"})
				lifecycleBody.SetAttributeValue("ignore_changes", cty.ListVal(lifecycle))
				outVars := generateRdsOutputVars(varFullPrefix, resourceName, "duplocloud_rds_read_replica")
				tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
				// Import all created resources.
				if config.GenerateTfState {
					importConfigs = append(importConfigs, common.ImportConfig{
						ResourceAddress: "duplocloud_rds_read_replica." + resourceName,
						ResourceId:      "v2/subscriptions/" + config.TenantId + "/RDSDBInstance/" + shortName,
						WorkingDir:      workingDir,
					})
					tfContext.ImportConfigs = importConfigs
				}

			} else {

				varFullPrefix := RDS_VAR_PREFIX
				inputVars := generateRdsVars(rds, varFullPrefix)
				tfContext.InputVars = append(tfContext.InputVars, inputVars...)

				if len(rds.MasterPassword) > 0 {
					randomBlock := rootBody.AppendNewBlock("resource",
						[]string{"random_password",
							resourceName + "_password"})
					randomBody := randomBlock.Body()
					randomBody.SetAttributeValue("length",
						cty.NumberIntVal(int64(16)))
					randomBody.SetAttributeValue("special",
						cty.BoolVal(false))
				}
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
				name := shortName
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
					rdsBody.SetAttributeTraversal("master_password", hcl.Traversal{
						hcl.TraverseRoot{
							Name: " random_password." + resourceName + "_password",
						},
						hcl.TraverseAttr{
							Name: "result",
						},
					})
				}

				if len(rds.DBParameterGroupName) > 0 {
					rdsBody.SetAttributeValue("parameter_group_name",
						cty.StringVal(rds.DBParameterGroupName))
				}
				rdsBody.SetAttributeValue("store_details_in_secret_manager",
					cty.BoolVal(rds.StoreDetailsInSecretManager))
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
				if rds.StorageType != "" {
					rdsBody.SetAttributeValue("storage_type", cty.StringVal(rds.StorageType))
					if rds.Iops != 0 {
						rdsBody.SetAttributeValue("iops", cty.NumberIntVal(int64(rds.Iops)))
					}
				}
				if rds.V2ScalingConfiguration != nil && (rds.V2ScalingConfiguration.MaxCapacity > 0 || rds.V2ScalingConfiguration.MinCapacity > 0) {
					scalingBody := rdsBody.AppendNewBlock("v2_scaling_configuration", nil).Body()
					scalingBody.SetAttributeTraversal("min_capacity", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "var",
						},
						hcl.TraverseAttr{
							Name: varFullPrefix + "scaling_config_min_capacity",
						},
					})
					scalingBody.SetAttributeTraversal("max_capacity", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "var",
						},
						hcl.TraverseAttr{
							Name: varFullPrefix + "scaling_config_max_capacity",
						},
					})
				}
				if rds.EnablePerformanceInsights {
					inightBody := rdsBody.AppendNewBlock("performance_insights", nil).Body()
					inightBody.SetAttributeTraversal("enabled", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "var",
						},
						hcl.TraverseAttr{
							Name: varFullPrefix + "performance_insights_enabled",
						},
					})
					inightBody.SetAttributeTraversal("retention_period", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "var",
						},
						hcl.TraverseAttr{
							Name: varFullPrefix + "performance_insights_retention_period",
						},
					})
					inightBody.SetAttributeTraversal("kms_key_id", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "var",
						},
						hcl.TraverseAttr{
							Name: varFullPrefix + "performance_insights_kms_key_id",
						},
					})
				}

				rdsBody.SetAttributeValue("enable_iam_auth", cty.BoolVal(rds.EnableIamAuth))
				lifecycleBody := rdsBody.AppendNewBlock("lifecycle", nil).Body()
				lifecycle := common.StringSliceToListVal([]string{"engine_version"})
				lifecycleBody.SetAttributeValue("ignore_changes", cty.ListVal(lifecycle))
				outVars := generateRdsOutputVars(varFullPrefix, resourceName, "duplocloud_rds_instance")
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

			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo RDS instance : %s", rds.Identifier)

		}
		log.Println("[TRACE] <====== RDS TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateRdsRRVars(duplo duplosdk.DuploRdsInstance, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	size := duplo.SizeEx
	if len(size) == 0 {
		size = "db.t2.medium"
	}
	var1 := common.VarConfig{
		Name:       prefix + "size",
		DefaultVal: size,
		TypeVal:    "string",
	}
	varConfigs["size"] = var1

	var2 := common.VarConfig{
		Name:       prefix + "scaling_config_max_capacity",
		DefaultVal: fmt.Sprintf("%f", duplo.V2ScalingConfiguration.MaxCapacity),
		TypeVal:    "number",
	}
	varConfigs["max_capacity"] = var2

	var3 := common.VarConfig{
		Name:       prefix + "scaling_config_min_capacity",
		DefaultVal: fmt.Sprintf("%f", duplo.V2ScalingConfiguration.MaxCapacity),
		TypeVal:    "number",
	}
	varConfigs["min_capacity"] = var3

	var4 := common.VarConfig{
		Name:       prefix + "performance_insights_enabled",
		DefaultVal: strconv.FormatBool(duplo.EnablePerformanceInsights),
		TypeVal:    "bool",
	}
	varConfigs["enabled"] = var4

	var5 := common.VarConfig{
		Name:       prefix + "performance_insights_kms_key_id",
		DefaultVal: duplo.PerformanceInsightsKMSKeyId,
		TypeVal:    "string",
	}
	varConfigs["kms_key_id"] = var5

	var6 := common.VarConfig{
		Name:       prefix + "performance_insights_retention_period",
		DefaultVal: strconv.Itoa(duplo.PerformanceInsightsRetentionPeriod),
		TypeVal:    "number",
	}
	varConfigs["retention_period"] = var6

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}

func generateRdsVars(duplo duplosdk.DuploRdsInstance, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	var1 := common.VarConfig{
		Name:       prefix + "engine_version",
		DefaultVal: duplo.EngineVersion,
		TypeVal:    "string",
	}
	varConfigs["engine_version"] = var1
	size := duplo.SizeEx
	if len(size) == 0 {
		size = "db.t2.medium"
	}
	var2 := common.VarConfig{
		Name:       prefix + "size",
		DefaultVal: size,
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
		Name:       prefix + "master_username",
		DefaultVal: duplo.MasterUsername,
		TypeVal:    "string",
	}
	varConfigs["master_username"] = var4

	var5 := common.VarConfig{
		Name:       prefix + "scaling_config_max_capacity",
		DefaultVal: fmt.Sprintf("%f", duplo.V2ScalingConfiguration.MaxCapacity),
		TypeVal:    "number",
	}
	varConfigs["max_capacity"] = var5

	var6 := common.VarConfig{
		Name:       prefix + "scaling_config_min_capacity",
		DefaultVal: fmt.Sprintf("%f", duplo.V2ScalingConfiguration.MaxCapacity),
		TypeVal:    "number",
	}
	varConfigs["min_capacity"] = var6

	var7 := common.VarConfig{
		Name:       prefix + "performance_insights_enabled",
		DefaultVal: strconv.FormatBool(duplo.EnablePerformanceInsights),
		TypeVal:    "bool",
	}
	varConfigs["enabled"] = var7
	if duplo.PerformanceInsightsKMSKeyId != "" {
		var8 := common.VarConfig{
			Name:       prefix + "performance_insights_kms_key_id",
			DefaultVal: duplo.PerformanceInsightsKMSKeyId,
			TypeVal:    "string",
		}
		varConfigs["kms_key_id"] = var8
	}
	var9 := common.VarConfig{
		Name:       prefix + "performance_insights_retention_period",
		DefaultVal: strconv.Itoa(duplo.PerformanceInsightsRetentionPeriod),
		TypeVal:    "number",
	}
	varConfigs["retention_period"] = var9
	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}

func generateRdsOutputVars(prefix, resourceName string, resourceType string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "fullname",
		ActualVal:     resourceType + "." + resourceName + ".identifier",
		DescVal:       "The full name of the RDS instance.",
		RootTraversal: true,
	}
	outVarConfigs["fullname"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     resourceType + "." + resourceName + ".arn",
		DescVal:       "The ARN of the RDS instance.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "endpoint",
		ActualVal:     resourceType + "." + resourceName + ".endpoint",
		DescVal:       "The endpoint of the RDS instance.",
		RootTraversal: true,
	}
	outVarConfigs["endpoint"] = var3

	var4 := common.OutputVarConfig{
		Name:          prefix + "host",
		ActualVal:     resourceType + "." + resourceName + ".host",
		DescVal:       "The DNS hostname of the RDS instance.",
		RootTraversal: true,
	}
	outVarConfigs["host"] = var4

	var5 := common.OutputVarConfig{
		Name:          prefix + "port",
		ActualVal:     resourceType + "." + resourceName + ".port",
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
