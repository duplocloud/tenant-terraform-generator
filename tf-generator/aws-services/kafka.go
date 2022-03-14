package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type Kafka struct {
}

func (k *Kafka) Generate(config *common.Config, client *duplosdk.Client) {
	log.Println("[TRACE] <====== kafka TF generation started. =====>")
	workingDir := filepath.Join("target", config.CustomerName, config.AwsServicesProject)
	list, clientErr := client.TenantListKafkaCluster(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return
	}

	if list != nil {
		for _, kafka := range *list {
			shortName := kafka.Name[len("duploservices-"+config.TenantName+"-"):len(kafka.Name)]
			log.Printf("[TRACE] Generating terraform config for duplo kafka Instance : %s", shortName)

			clusterInfo, clientErr := client.TenantGetKafkaClusterInfo(config.TenantId, kafka.Arn)
			if clientErr != nil {
				fmt.Println(clientErr)
				return
			}
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "kafka-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_ecache_instance resource
			kafkaBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_kafka_cluster",
					shortName})
			kafkaBody := kafkaBlock.Body()
			// kafkaBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
			// 	hcl.TraverseRoot{
			// 		Name: "duplocloud_tenant.tenant",
			// 	},
			// 	hcl.TraverseAttr{
			// 		Name: "tenant_id",
			// 	},
			// })
			kafkaBody.SetAttributeValue("tenant_id",
				cty.StringVal(config.TenantId))
			kafkaBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			kafkaBody.SetAttributeValue("storage_size",
				cty.NumberIntVal(int64(clusterInfo.BrokerNodeGroup.StorageInfo.EbsStorageInfo.VolumeSize)))

			kafkaBody.SetAttributeValue("instance_type",
				cty.StringVal(clusterInfo.BrokerNodeGroup.InstanceType))

			if clusterInfo.CurrentSoftware != nil {
				kafkaBody.SetAttributeValue("kafka_version",
					cty.StringVal(clusterInfo.CurrentSoftware.KafkaVersion))
				kafkaBody.SetAttributeValue("configuration_revision",
					cty.NumberIntVal(int64(clusterInfo.CurrentSoftware.ConfigurationRevision)))
				kafkaBody.SetAttributeValue("configuration_arn",
					cty.StringVal(clusterInfo.CurrentSoftware.ConfigurationArn))

			}

			//fmt.Printf("%s", hclFile.Bytes())
			tfFile.Write(hclFile.Bytes())
			log.Printf("[TRACE] Terraform config is generated for duplo kafka instance : %s", shortName)

			// Import all created resources.
			if config.GenerateTfState {
				importer := &common.Importer{}
				importer.Import(config, &common.ImportConfig{
					ResourceAddress: "duplocloud_aws_kafka_cluster." + shortName,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/ECacheDBInstance/" + shortName,
					WorkingDir:      workingDir,
				})
			}
		}
	}
	log.Println("[TRACE] <====== kafka TF generation done. =====>")
}
