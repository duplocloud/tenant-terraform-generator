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

const SSM_VAR_PREFIX = "ssm_"

type SsmParams struct {
}

func (ssmParams *SsmParams) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	log.Println("[TRACE] <====== Ssm params TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.SsmParameterList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	if list != nil {
		for _, ssmParam := range *list {
			shortName := ssmParam.Name
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo SSM Parameter : %s", shortName)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "ssm-param-"+resourceName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			ssmDetails, err := client.SsmParameterGet(config.TenantId, shortName)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			ssmParamBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_ssm_parameter",
					resourceName})
			ssmParamBody := ssmParamBlock.Body()
			ssmParamBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			ssmParamBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			ssmParamBody.SetAttributeValue("type",
				cty.StringVal(ssmParam.Type))

			if len(ssmDetails.Value) > 0 {
				ssmParamBody.SetAttributeValue("value",
					cty.StringVal(ssmDetails.Value))
			}

			if len(ssmDetails.Description) > 0 {
				ssmParamBody.SetAttributeValue("description",
					cty.StringVal(ssmDetails.Description))
			}

			if len(ssmDetails.KeyId) > 0 {
				ssmParamBody.SetAttributeValue("key_id",
					cty.StringVal(ssmDetails.KeyId))
			}

			if len(ssmDetails.AllowedPattern) > 0 {
				ssmParamBody.SetAttributeValue("allowed_pattern",
					cty.StringVal(ssmDetails.AllowedPattern))
			}
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo SSM parameter : %s", shortName)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs := []common.ImportConfig{}
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_ssm_parameter." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
	}
	log.Println("[TRACE] <====== SSM Parameters TF generation done. =====>")

	return &tfContext, nil
}
