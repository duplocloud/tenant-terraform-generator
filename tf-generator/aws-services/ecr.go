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

const ECR_VAR_PREFIX = "ecr_"

type ECR struct {
}

func (ecr *ECR) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.AwsEcrRepositoryList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== ECR TF generation started. =====>")
		for _, ecr := range *list {
			shortName := ecr.Name
			resourceName := strings.ReplaceAll(shortName, ".", "_")
			log.Printf("[TRACE] Generating terraform config for duplo AWS ECR : %s", shortName)

			varFullPrefix := ECR_VAR_PREFIX + strings.ReplaceAll(shortName, "-", "_") + "_"

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "ecr-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			// initialize the body of the new file object
			rootBody := hclFile.Body()

			ecrBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_ecr_repository",
					resourceName})
			ecrBody := ecrBlock.Body()
			ecrBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			ecrBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			ecrBody.SetAttributeValue("enable_scan_image_on_push",
				cty.BoolVal(ecr.EnableScanImageOnPush))

			ecrBody.SetAttributeValue("enable_tag_immutability",
				cty.BoolVal(ecr.EnableTagImmutability))

			if len(ecr.KmsEncryption) > 0 {
				ecrBody.SetAttributeValue("kms_encryption_key",
					cty.StringVal(ecr.KmsEncryption))
			}
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo AWS ECR : %s", shortName)

			outVars := generateECROutputVars(varFullPrefix, shortName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_ecr_repository." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== ECR TF generation done. =====>")
	}
	return &tfContext, nil
}

func generateECROutputVars(prefix, shortName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "registry_id",
		ActualVal:     "duplocloud_aws_ecr_repository." + shortName + ".registry_id",
		DescVal:       "The registry ID where the repository was created.",
		RootTraversal: true,
	}
	outVarConfigs["registry_id"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_ecr_repository." + shortName + ".arn",
		DescVal:       "Full ARN of the repository.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "repository_url",
		ActualVal:     "duplocloud_aws_ecr_repository." + shortName + ".repository_url",
		DescVal:       "The DNS name of the load balancer.",
		RootTraversal: true,
	}
	outVarConfigs["repository_url"] = var3

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
