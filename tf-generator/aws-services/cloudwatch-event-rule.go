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

type CloudwatchEventRule struct {
}

func (cwer *CloudwatchEventRule) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.DuploCloudWatchEventRuleList(config.TenantId)

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== Cloudwatch event rules TF generation started. =====>")
		for _, cwer := range *list {
			shortName := cwer.Name[len("duploservices-"+config.TenantName+"-"):len(cwer.Name)]
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo Cloudwatch event rules : %s", shortName)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "cw-event-rule-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			// initialize the body of the new file object
			rootBody := hclFile.Body()

			cwerBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_cloudwatch_event_rule",
					resourceName})
			cwerBody := cwerBlock.Body()
			cwerBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			cwerBody.SetAttributeValue("name",
				cty.StringVal(cwer.Name))

			if len(cwer.Description) > 0 {
				cwerBody.SetAttributeValue("description",
					cty.StringVal(cwer.Description))
			}
			if len(cwer.EventBusName) > 0 {
				cwerBody.SetAttributeValue("event_bus_name",
					cty.StringVal(cwer.EventBusName))
			}
			if len(cwer.RoleArn) > 0 {
				cwerBody.SetAttributeValue("role_arn",
					cty.StringVal(cwer.RoleArn))
			}
			if len(cwer.ScheduleExpression) > 0 {
				cwerBody.SetAttributeValue("schedule_expression",
					cty.StringVal(cwer.ScheduleExpression))
			}
			if cwer.State != nil && len(cwer.State.Value) > 0 {
				cwerBody.SetAttributeValue("state",
					cty.StringVal(cwer.State.Value))
			}
			targetList, _ := client.DuploCloudWatchEventTargetsList(config.TenantId, cwer.Name)
			if targetList != nil && len(*targetList) > 0 {
				rootBody.AppendNewline()
				for _, target := range *targetList {
					targetResourceName := resourceName + "-target"
					cwetBlock := rootBody.AppendNewBlock("resource",
						[]string{"duplocloud_aws_cloudwatch_event_target",
							targetResourceName})
					cwetBody := cwetBlock.Body()
					cwetBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "local",
						},
						hcl.TraverseAttr{
							Name: "tenant_id",
						},
					})
					cwetBody.SetAttributeTraversal("rule_name", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "duplocloud_aws_cloudwatch_event_rule." + resourceName,
						},
						hcl.TraverseAttr{
							Name: "fullname",
						},
					})
					cwetBody.SetAttributeValue("target_arn",
						cty.StringVal(target.Arn))
					cwetBody.SetAttributeValue("target_id",
						cty.StringVal(target.Id))
					if len(target.RoleArn) > 0 {
						cwetBody.SetAttributeValue("role_arn",
							cty.StringVal(target.RoleArn))
					}
					if len(cwer.EventBusName) > 0 {
						cwetBody.SetAttributeValue("event_bus_name",
							cty.StringVal(cwer.EventBusName))
					}
					if config.GenerateTfState {
						importConfigs = append(importConfigs, common.ImportConfig{
							ResourceAddress: "duplocloud_aws_cloudwatch_event_target." + targetResourceName,
							ResourceId:      config.TenantId + "/" + cwer.Name + "/" + target.Id,
							WorkingDir:      workingDir,
						})
					}
					rootBody.AppendNewline()
				}

			}
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo Cloudwatch metrics : %s", shortName)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_cloudwatch_event_rule." + resourceName,
					ResourceId:      config.TenantId + "/" + cwer.Name,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== Cloudwatch event rule TF generation done. =====>")
	}

	return &tfContext, nil
}
