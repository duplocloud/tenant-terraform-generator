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

const SQS_VAR_PREFIX = "sqs_"

type SQS struct {
}

func (sqs *SQS) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	log.Println("[TRACE] <====== SQS TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantListSQS(config.TenantId)

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	if list != nil {
		for _, sqs := range *list {
			shortName, err := extractSqsName(client, config.TenantId, sqs.Name)
			resourceName := common.GetResourceName(shortName)
			if err != nil {
				return nil, err
			}
			log.Printf("[TRACE] Generating terraform config for duplo SQS : %s", shortName)
			varFullPrefix := SQS_VAR_PREFIX + resourceName + "_"
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "sqs-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_sqs_queue resource
			sqsBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_sqs_queue",
					resourceName})
			sqsBody := sqsBlock.Body()
			sqsBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			sqsBody.SetAttributeValue("name",
				cty.StringVal(shortName))
			if strings.HasSuffix(sqs.Name, ".fifo") {
				sqsBody.SetAttributeValue("fifo_queue",
					cty.BoolVal(strings.HasSuffix(sqs.Name, ".fifo")))
			}
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo SQS : %s", shortName)

			outVars := generateSQSOutputVars(sqs, varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs := []common.ImportConfig{}
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_sqs_queue." + resourceName,
					ResourceId:      config.TenantId + "/" + sqs.Name,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
	}
	log.Println("[TRACE] <====== SQS TF generation done. =====>")
	return &tfContext, nil
}

func generateSQSOutputVars(duplo duplosdk.DuploAwsResource, prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	urlVar := common.OutputVarConfig{
		Name:          prefix + "url",
		ActualVal:     "duplocloud_aws_sqs_queue." + resourceName + ".url",
		DescVal:       "The URL for the created Amazon SQS queue.",
		RootTraversal: true,
	}
	outVarConfigs["url"] = urlVar

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}

func extractSqsName(client *duplosdk.Client, tenantID string, sqsUrl string) (string, error) {
	accountID, err := client.TenantGetAwsAccountID(tenantID)
	if err != nil {
		return "", err
	}
	prefix, err := client.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return "", err
	}
	parts := strings.Split(sqsUrl, "/"+accountID+"/")
	fullname := parts[1]
	if strings.HasSuffix(fullname, ".fifo") {
		fullname = strings.TrimSuffix(fullname, ".fifo")
	}
	name, _ := duplosdk.UnwrapName(prefix, accountID, fullname, true)
	return name, nil
}
