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

const DYN_DB_VAR_PREFIX = "dynamodb_"

type DynamoDB struct {
}

func (dynamodb *DynamoDB) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantDynamoDBList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		for _, dynamodb := range *list {
			shortName, _ := extractDynamoDBName(client, config.TenantId, dynamodb.Name)
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for DynamoDB : %s", shortName)

			dynamodbInfo, clientErr := client.DynamoDBTableGetV2(config.TenantId, shortName)
			if clientErr != nil {
				fmt.Println(clientErr)
				return nil, nil
			}
			varFullPrefix := DYN_DB_VAR_PREFIX + resourceName + "_"
			// inputVars := generateKafkaVars(clusterInfo, varFullPrefix)
			// tfContext.InputVars = append(tfContext.InputVars, inputVars...)
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "dynamodb-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_ecache_instance resource
			dynamodbBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_dynamodb_table_v2",
					resourceName})
			dynamodbBody := dynamodbBlock.Body()
			dynamodbBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// kafkaBody.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			dynamodbBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			if dynamodbInfo.BillingModeSummary != nil && dynamodbInfo.BillingModeSummary.BillingMode != nil {
				dynamodbBody.SetAttributeValue("billing_mode",
					cty.StringVal(dynamodbInfo.BillingModeSummary.BillingMode.Value))
			}
			if dynamodbInfo.ProvisionedThroughput != nil && dynamodbInfo.ProvisionedThroughput.ReadCapacityUnits > 0 {
				dynamodbBody.SetAttributeValue("read_capacity",
					cty.NumberIntVal(int64(dynamodbInfo.ProvisionedThroughput.ReadCapacityUnits)))

			}
			if dynamodbInfo.ProvisionedThroughput != nil && dynamodbInfo.ProvisionedThroughput.WriteCapacityUnits > 0 {
				dynamodbBody.SetAttributeValue("write_capacity",
					cty.NumberIntVal(int64(dynamodbInfo.ProvisionedThroughput.WriteCapacityUnits)))
			}
			if dynamodbInfo.StreamSpecification != nil && dynamodbInfo.StreamSpecification.StreamEnabled {
				dynamodbBody.SetAttributeValue("stream_enabled",
					cty.BoolVal(dynamodbInfo.StreamSpecification.StreamEnabled))
			}
			if dynamodbInfo.StreamSpecification != nil && dynamodbInfo.StreamSpecification.StreamViewType != nil {
				dynamodbBody.SetAttributeValue("stream_view_type",
					cty.StringVal(dynamodbInfo.StreamSpecification.StreamViewType.Value))
			}
			if dynamodbInfo.AttributeDefinitions != nil && len(*dynamodbInfo.AttributeDefinitions) > 0 {
				for _, attr := range *dynamodbInfo.AttributeDefinitions {
					attrBlock := dynamodbBody.AppendNewBlock("attribute", nil)
					attrBody := attrBlock.Body()
					attrBody.SetAttributeValue("name",
						cty.StringVal(attr.AttributeName))
					attrBody.SetAttributeValue("type",
						cty.StringVal(attr.AttributeType.Value))
				}
			}
			if dynamodbInfo.KeySchema != nil && len(*dynamodbInfo.KeySchema) > 0 {
				for _, keySchema := range *dynamodbInfo.KeySchema {
					ksBlock := dynamodbBody.AppendNewBlock("key_schema", nil)
					ksBody := ksBlock.Body()
					ksBody.SetAttributeValue("attribute_name",
						cty.StringVal(keySchema.AttributeName))
					ksBody.SetAttributeValue("key_type",
						cty.StringVal(keySchema.KeyType.Value))
				}
			}
			if dynamodbInfo.GlobalSecondaryIndexes != nil && len(*dynamodbInfo.GlobalSecondaryIndexes) > 0 {
				for _, gsi := range *dynamodbInfo.GlobalSecondaryIndexes {
					gsiBlock := dynamodbBody.AppendNewBlock("global_secondary_index", nil)
					gsiBody := gsiBlock.Body()
					gsiBody.SetAttributeValue("name",
						cty.StringVal(gsi.IndexName))
					gsiBody.SetAttributeValue("projection_type",
						cty.StringVal(gsi.Projection.ProjectionType.Value))
					for _, attribute := range *gsi.KeySchema {
						if attribute.KeyType.Value == "HASH" {
							gsiBody.SetAttributeValue("hash_key",
								cty.StringVal(attribute.AttributeName))
						}
						if attribute.KeyType.Value == "RANGE" {
							gsiBody.SetAttributeValue("range_key",
								cty.StringVal(attribute.AttributeName))
						}
					}
					if gsi.ProvisionedThroughput != nil && gsi.ProvisionedThroughput.ReadCapacityUnits > 0 {
						gsiBody.SetAttributeValue("read_capacity",
							cty.NumberIntVal(int64(gsi.ProvisionedThroughput.ReadCapacityUnits)))

					}
					if gsi.ProvisionedThroughput != nil && gsi.ProvisionedThroughput.WriteCapacityUnits > 0 {
						gsiBody.SetAttributeValue("write_capacity",
							cty.NumberIntVal(int64(gsi.ProvisionedThroughput.WriteCapacityUnits)))
					}
					if len(gsi.Projection.NonKeyAttributes) > 0 {
						var vals []cty.Value
						for _, s := range gsi.Projection.NonKeyAttributes {
							vals = append(vals, cty.StringVal(s))
						}
						gsiBody.SetAttributeValue("non_key_attributes", cty.SetVal(vals))
					}
				}
			}
			if dynamodbInfo.LocalSecondaryIndexes != nil && len(*dynamodbInfo.LocalSecondaryIndexes) > 0 {
				for _, lsi := range *dynamodbInfo.LocalSecondaryIndexes {
					lsiBlock := dynamodbBody.AppendNewBlock("local_secondary_index", nil)
					lsiBody := lsiBlock.Body()
					lsiBody.SetAttributeValue("name",
						cty.StringVal(lsi.IndexName))
					lsiBody.SetAttributeValue("projection_type",
						cty.StringVal(lsi.Projection.ProjectionType.Value))
					for _, attribute := range *lsi.KeySchema {
						if attribute.KeyType.Value == "RANGE" {
							lsiBody.SetAttributeValue("range_key",
								cty.StringVal(attribute.AttributeName))
						}
					}
					if len(lsi.Projection.NonKeyAttributes) > 0 {
						var vals []cty.Value
						for _, s := range lsi.Projection.NonKeyAttributes {
							vals = append(vals, cty.StringVal(s))
						}
						lsiBody.SetAttributeValue("non_key_attributes", cty.SetVal(vals))
					}
				}
			}
			if dynamodbInfo.SSEDescription != nil && dynamodbInfo.SSEDescription.Enabled {
				sseBlock := dynamodbBody.AppendNewBlock("server_side_encryption", nil)
				sseBody := sseBlock.Body()
				sseBody.SetAttributeValue("enabled",
					cty.BoolVal(dynamodbInfo.SSEDescription.Enabled))
				sseBody.SetAttributeValue("kms_key_arn",
					cty.StringVal(dynamodbInfo.SSEDescription.KMSMasterKeyArn))
			}
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			log.Printf("[TRACE] Terraform config is generated for duplo DynamoDB instance : %s", shortName)

			outVars := generateDynamoDBOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_dynamodb_table_v2." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
	}
	log.Println("[TRACE] <====== DynamoDB TF generation done. =====>")
	return &tfContext, nil
}

func generateDynamoDBOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "stream_arn",
		ActualVal:     "duplocloud_aws_dynamodb_table_v2." + resourceName + ".stream_arn",
		DescVal:       "The Stream ARN of the dynamodb table.",
		RootTraversal: true,
	}
	outVarConfigs["stream_arn"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_dynamodb_table_v2." + resourceName + ".arn",
		DescVal:       "The ARN of the dynamodb table.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}

func extractDynamoDBName(client *duplosdk.Client, tenantID string, fullname string) (string, error) {
	prefix, err := client.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return "", err
	}
	name, _ := duplosdk.UnprefixName(prefix, fullname)
	return name, nil
}
