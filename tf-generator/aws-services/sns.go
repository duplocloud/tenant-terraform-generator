package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const SNS_VAR_PREFIX = "sns_"

type SNS struct {
}

func (sns *SNS) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	log.Println("[TRACE] <====== SNS Topic TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantListSnsTopic(config.TenantId)
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tenantKms, clientErr := client.TenantGetTenantKmsKey(config.TenantId)
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		for _, sns := range *list {
			// shortName, err := extractSnsTopicName(client, config.TenantId, sns.Name)
			shortName, err := extractSnsTopicName(client, config.TenantId, sns.Name)
			if err != nil {
				return nil, err
			}
			log.Printf("[TRACE] Generating terraform config for duplo SNS Topic : %s", shortName)
			varFullPrefix := SNS_VAR_PREFIX + strings.ReplaceAll(shortName, "-", "_") + "_"
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "sns-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_sns_topic resource
			snsBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_sns_topic",
					shortName})
			snsBody := snsBlock.Body()
			snsBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			snsBody.SetAttributeValue("name",
				cty.StringVal(shortName))
			snsBody.SetAttributeValue("kms_key_id",
				cty.StringVal(tenantKms.KeyArn))
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo SNS Topic : %s", shortName)

			outVars := generateSnsOutputVars(sns, varFullPrefix, shortName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_sns_topic." + shortName,
					ResourceId:      config.TenantId + "/" + sns.Name,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
	}
	log.Println("[TRACE] <====== SNS Topic TF generation done. =====>")
	return &tfContext, nil
}

func generateSnsOutputVars(duplo duplosdk.DuploAwsResource, prefix, shortName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	arnVar := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_sns_topic." + shortName + ".arn",
		DescVal:       "The ARN of the SNS topic.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = arnVar

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}

func extractSnsTopicName(client *duplosdk.Client, tenantID string, topicName string) (string, error) {
	accountID, err := client.TenantGetAwsAccountID(tenantID)
	if err != nil {
		return "", err
	}
	prefix, err := client.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return "", err
	}
	parts := strings.Split(topicName, ":"+accountID+":")
	fullname := parts[1]
	name, _ := duplosdk.UnwrapName(prefix, accountID, fullname, true)
	return name, nil
}
