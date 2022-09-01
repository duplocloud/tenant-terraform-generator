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
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const KAFKA_VAR_PREFIX = "kafka_"

type Kafka struct {
}

func (k *Kafka) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantListKafkaCluster(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== kafka TF generation started. =====>")
		for _, kafka := range *list {
			shortName := kafka.Name[len("duploservices-"+config.TenantName+"-"):len(kafka.Name)]
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo kafka Instance : %s", shortName)

			clusterInfo, clientErr := client.TenantGetKafkaClusterInfo(config.TenantId, kafka.Arn)
			if clientErr != nil {
				fmt.Println(clientErr)
				return nil, clientErr
			}
			varFullPrefix := KAFKA_VAR_PREFIX + resourceName + "_"
			inputVars := generateKafkaVars(clusterInfo, varFullPrefix)
			tfContext.InputVars = append(tfContext.InputVars, inputVars...)
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "kafka-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_ecache_instance resource
			kafkaBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_kafka_cluster",
					resourceName})
			kafkaBody := kafkaBlock.Body()
			kafkaBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// kafkaBody.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			kafkaBody.SetAttributeValue("name",
				cty.StringVal(shortName))
			kafkaBody.SetAttributeTraversal("storage_size", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "storage_size",
				},
			})
			kafkaBody.SetAttributeTraversal("instance_type", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "instance_type",
				},
			})

			if clusterInfo.CurrentSoftware != nil {
				kafkaBody.SetAttributeTraversal("kafka_version", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "var",
					},
					hcl.TraverseAttr{
						Name: varFullPrefix + "kafka_version",
					},
				})
				if len(clusterInfo.CurrentSoftware.ConfigurationArn) > 0 {
					kafkaBody.SetAttributeValue("configuration_revision",
						cty.NumberIntVal(int64(clusterInfo.CurrentSoftware.ConfigurationRevision)))
					kafkaBody.SetAttributeValue("configuration_arn",
						cty.StringVal(clusterInfo.CurrentSoftware.ConfigurationArn))
				}
			}

			if clusterInfo.BrokerNodeGroup != nil && len(*clusterInfo.BrokerNodeGroup.Subnets) > 0 {
				var vals []cty.Value
				for _, s := range *clusterInfo.BrokerNodeGroup.Subnets {
					vals = append(vals, cty.StringVal(s))
				}
				kafkaBody.SetAttributeValue("subnets",
					cty.ListVal(vals))
			}
			if clusterInfo.EncryptionInfo != nil && clusterInfo.EncryptionInfo.InTransit != nil && clusterInfo.EncryptionInfo.InTransit.ClientBroker != nil {
				kafkaBody.SetAttributeValue("encryption_in_transit",
					cty.StringVal(clusterInfo.EncryptionInfo.InTransit.ClientBroker.Value))
			}
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			log.Printf("[TRACE] Terraform config is generated for duplo kafka instance : %s", shortName)

			outVars := generateKafkaOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_kafka_cluster." + resourceName,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/ECacheDBInstance/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== kafka TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateKafkaVars(duplo *duplosdk.DuploKafkaClusterInfo, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	var1 := common.VarConfig{
		Name:       prefix + "kafka_version",
		DefaultVal: duplo.CurrentSoftware.KafkaVersion,
		TypeVal:    "string",
	}
	varConfigs["kafka_version"] = var1

	var2 := common.VarConfig{
		Name:       prefix + "instance_type",
		DefaultVal: duplo.BrokerNodeGroup.InstanceType,
		TypeVal:    "string",
	}
	varConfigs["instance_type"] = var2

	var3 := common.VarConfig{
		Name:       prefix + "storage_size",
		DefaultVal: strconv.Itoa(duplo.BrokerNodeGroup.StorageInfo.EbsStorageInfo.VolumeSize),
		TypeVal:    "number",
	}
	varConfigs["storage_size"] = var3

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}

func generateKafkaOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "fullname",
		ActualVal:     "duplocloud_aws_kafka_cluster." + resourceName + ".fullname",
		DescVal:       "The full name of the Kakfa cluster.",
		RootTraversal: true,
	}
	outVarConfigs["fullname"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_kafka_cluster." + resourceName + ".arn",
		DescVal:       "The ARN of the Kafka cluster.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "number_of_broker_nodes",
		ActualVal:     "duplocloud_aws_kafka_cluster." + resourceName + ".number_of_broker_nodes",
		DescVal:       "The desired total number of broker nodes in the kafka cluster.",
		RootTraversal: true,
	}
	outVarConfigs["number_of_broker_nodes"] = var3

	var4 := common.OutputVarConfig{
		Name:          prefix + "plaintext_bootstrap_broker_string",
		ActualVal:     "duplocloud_aws_kafka_cluster." + resourceName + ".plaintext_bootstrap_broker_string",
		DescVal:       "The bootstrap broker connect string for plaintext (unencrypted) connections.",
		RootTraversal: true,
	}
	outVarConfigs["plaintext_bootstrap_broker_string"] = var4

	var5 := common.OutputVarConfig{
		Name:          prefix + "plaintext_zookeeper_connect_string",
		ActualVal:     "duplocloud_aws_kafka_cluster." + resourceName + ".plaintext_zookeeper_connect_string",
		DescVal:       "The bootstrap broker connect string for plaintext (unencrypted) connections.",
		RootTraversal: true,
	}
	outVarConfigs["plaintext_zookeeper_connect_string"] = var5

	var6 := common.OutputVarConfig{
		Name:          prefix + "tls_bootstrap_broker_string",
		ActualVal:     "duplocloud_aws_kafka_cluster." + resourceName + ".tls_bootstrap_broker_string",
		DescVal:       "The bootstrap broker connect string for TLS (encrypted) connections.",
		RootTraversal: true,
	}
	outVarConfigs["tls_bootstrap_broker_string"] = var6

	var7 := common.OutputVarConfig{
		Name:          prefix + "tls_zookeeper_connect_string",
		ActualVal:     "duplocloud_aws_kafka_cluster." + resourceName + ".tls_zookeeper_connect_string",
		DescVal:       "he bootstrap broker connect string for plaintext (unencrypted) connections.",
		RootTraversal: true,
	}
	outVarConfigs["tls_zookeeper_connect_string"] = var7

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
