package awsservices

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

const ES_VAR_PREFIX = "es_"

const (
	// TLSSecurityPolicyPolicyMinTLS10201907 is a TLSSecurityPolicy enum value
	TLSSecurityPolicyPolicyMinTLS10201907 = "Policy-Min-TLS-1-0-2019-07"

	// TLSSecurityPolicyPolicyMinTLS12201907 is a TLSSecurityPolicy enum value
	TLSSecurityPolicyPolicyMinTLS12201907 = "Policy-Min-TLS-1-2-2019-07"
)

type ES struct {
}

func (es *ES) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantListElasticSearchDomains(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== Elastic Search TF generation started. =====>")
		kms, kmsClientErr := client.TenantGetTenantKmsKey(config.TenantId)
		for _, es := range *list {
			shortName := es.Name
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo Elastic Search : %s", shortName)

			varFullPrefix := ES_VAR_PREFIX + resourceName + "_"
			inputVars := generateESVars(es, varFullPrefix)
			tfContext.InputVars = append(tfContext.InputVars, inputVars...)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "es-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_elasticsearch resource
			esBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_elasticsearch",
					resourceName})
			esBody := esBlock.Body()
			esBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			esBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			esBody.SetAttributeValue("use_latest_tls_cipher",
				cty.BoolVal(es.DomainEndpointOptions.TLSSecurityPolicy.Value == TLSSecurityPolicyPolicyMinTLS12201907))
			esBody.SetAttributeValue("require_ssl",
				cty.BoolVal(es.DomainEndpointOptions.EnforceHTTPS))
			esBody.SetAttributeTraversal("elasticsearch_version", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "elasticsearch_version",
				},
			})
			esBody.SetAttributeValue("enable_node_to_node_encryption",
				cty.BoolVal(es.NodeToNodeEncryptionOptions.Enabled))
			if es.EBSOptions.EBSEnabled {
				esBody.SetAttributeValue("storage_size",
					cty.NumberIntVal(int64(es.EBSOptions.VolumeSize)))
			}

			if len(es.ClusterConfig.InstanceType.Value) > 0 {
				clusterConfigBlock := esBody.AppendNewBlock("cluster_config",
					nil)
				clusterConfigBody := clusterConfigBlock.Body()

				if es.ClusterConfig.DedicatedMasterEnabled {
					clusterConfigBody.SetAttributeValue("dedicated_master_enabled",
						cty.BoolVal(es.ClusterConfig.DedicatedMasterEnabled))
					if es.ClusterConfig.DedicatedMasterCount > 0 {
						clusterConfigBody.SetAttributeValue("dedicated_master_count",
							cty.NumberIntVal(int64(es.ClusterConfig.DedicatedMasterCount)))
					}
					clusterConfigBody.SetAttributeValue("dedicated_master_type",
						cty.StringVal(es.ClusterConfig.DedicatedMasterType.Value))
				}
				clusterConfigBody.SetAttributeValue("instance_count",
					cty.NumberIntVal(int64(es.ClusterConfig.InstanceCount)))

				clusterConfigBody.SetAttributeTraversal("instance_type", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "var",
					},
					hcl.TraverseAttr{
						Name: varFullPrefix + "instance_type",
					},
				})
			}

			if es.EncryptionAtRestOptions.Enabled {
				encryptBlock := esBody.AppendNewBlock("encrypt_at_rest",
					nil)
				encryptBody := encryptBlock.Body()
				if es.EncryptionAtRestOptions.KmsKeyID != "" {
					if kms != nil && kmsClientErr == nil && (es.EncryptionAtRestOptions.KmsKeyID == kms.KeyArn || es.EncryptionAtRestOptions.KmsKeyID == kms.KeyID) {
						encryptBody.SetAttributeTraversal("kms_key_id", hcl.Traversal{
							hcl.TraverseRoot{
								Name: "data.duplocloud_tenant_aws_kms_key.tenant_kms",
							},
							hcl.TraverseAttr{
								Name: "key_id",
							},
						})
					} else {
						encryptBody.SetAttributeValue("kms_key_id",
							cty.StringVal(es.EncryptionAtRestOptions.KmsKeyID))
					}
				}
			}

			esBody.SetAttributeTraversal("selected_zone", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "selected_zone",
				},
			})
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo elastic search instance : %s", shortName)

			outVars := generateESOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_elasticsearch." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== Elastic Search TF generation done. =====>")
	}
	return &tfContext, nil
}

func generateESVars(duplo duplosdk.DuploElasticSearchDomain, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	var1 := common.VarConfig{
		Name:       prefix + "instance_type",
		DefaultVal: duplo.ClusterConfig.InstanceType.Value,
		TypeVal:    "string",
	}
	varConfigs["instance_type"] = var1

	var2 := common.VarConfig{
		Name:       prefix + "elasticsearch_version",
		DefaultVal: duplo.ElasticSearchVersion,
		TypeVal:    "string",
	}
	varConfigs["elasticsearch_version"] = var2

	var3 := common.VarConfig{
		Name:       prefix + "selected_zone",
		DefaultVal: "1",
		TypeVal:    "number",
	}
	varConfigs["selected_zone"] = var3

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}

func generateESOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "es_vpc_endpoint",
		ActualVal:     "duplocloud_aws_elasticsearch." + resourceName + ".endpoints[\"vpc\"]",
		DescVal:       "ES VPC endpoint.",
		RootTraversal: true,
	}
	outVarConfigs["es_vpc_endpoint"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_elasticsearch." + resourceName + ".arn",
		DescVal:       "The ARN of the ElasticSearch instance.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "domain_id",
		ActualVal:     "duplocloud_aws_elasticsearch." + resourceName + ".domain_id",
		DescVal:       "The domain ID of the ElasticSearch instance.",
		RootTraversal: true,
	}
	outVarConfigs["domain_id"] = var3

	var4 := common.OutputVarConfig{
		Name:          prefix + "domain_name",
		ActualVal:     "duplocloud_aws_elasticsearch." + resourceName + ".domain_name",
		DescVal:       "The full name of the ElasticSearch instance.",
		RootTraversal: true,
	}
	outVarConfigs["domain_name"] = var4

	var5 := common.OutputVarConfig{
		Name:          prefix + "endpoints",
		ActualVal:     "duplocloud_aws_elasticsearch." + resourceName + ".endpoints",
		DescVal:       "The endpoints to use when connecting to the ElasticSearch instance.",
		RootTraversal: true,
	}
	outVarConfigs["endpoints"] = var5

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
