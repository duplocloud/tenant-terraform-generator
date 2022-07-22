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

const LB_VAR_PREFIX = "lb_"

type LoadBalancer struct {
}

func (lb *LoadBalancer) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	log.Println("[TRACE] <====== Load balancer TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantGetApplicationLBList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	if list != nil {
		for _, lb := range *list {
			shortName, err := extractLbShortName(client, config.TenantId, lb.Name)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			settings, err := client.TenantGetApplicationLbSettings(config.TenantId, lb.Arn)
			if err != nil {
				fmt.Println(err)
				settings = nil
			}
			log.Printf("[TRACE] Generating terraform config for duplo aws load balancer : %s", shortName)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "lb-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			rootBody := hclFile.Body()

			lbBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_load_balancer",
					shortName})
			lbBody := lbBlock.Body()
			lbBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			lbBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			lbBody.SetAttributeValue("enable_access_logs",
				cty.BoolVal(lb.EnableAccessLogs))
			lbBody.SetAttributeValue("is_internal",
				cty.BoolVal(lb.IsInternal))

			if lb.LbType != nil {
				lbBody.SetAttributeValue("load_balancer_type",
					cty.StringVal(lb.LbType.Value))
			}

			if settings != nil {
				lbBody.SetAttributeValue("drop_invalid_headers",
					cty.BoolVal(settings.DropInvalidHeaders))
				if len(settings.WebACLID) > 0 {
					lbBody.SetAttributeValue("web_acl_id",
						cty.StringVal(settings.WebACLID))
				}
			}

			// Fetch all listeners
			listeners, clientErr := client.TenantListApplicationLbListeners(config.TenantId, shortName)
			if clientErr != nil {
				fmt.Println(err)
				listeners = nil
			}
			rootBody.AppendNewline()
			log.Printf("[TRACE] Terraform config is generation started for duplo aws load balancer listener : %s", shortName)
			if listeners != nil {
				for _, listener := range *listeners {
					listenerBlock := rootBody.AppendNewBlock("resource",
						[]string{"duplocloud_aws_load_balancer_listener",
							shortName + "-listener-" + strconv.Itoa(listener.Port)})
					listenerBody := listenerBlock.Body()

					listenerBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "local",
						},
						hcl.TraverseAttr{
							Name: "tenant_id",
						},
					})
					listenerBody.SetAttributeTraversal("load_balancer_name", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "duplocloud_aws_load_balancer",
						},
						hcl.TraverseAttr{
							Name: shortName + ".name",
						},
					})

					listenerBody.SetAttributeValue("protocol",
						cty.StringVal(listener.Protocol.Value))
					listenerBody.SetAttributeValue("protocol",
						cty.NumberIntVal(int64(listener.Port)))

					if len(listener.DefaultActions) > 0 {
						listenerBody.SetAttributeValue("target_group_arn",
							cty.StringVal(listener.DefaultActions[0].TargetGroupArn))
					}
					rootBody.AppendNewline()
				}
			}

			log.Printf("[TRACE] Terraform config is generated for duplo aws load balancer listener.: %s", shortName)
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo aws load balancer : %s", shortName)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs := []common.ImportConfig{}
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_load_balancer." + shortName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
	}
	log.Println("[TRACE] <====== Load balancer TF generation done. =====>")

	return &tfContext, nil
}

func extractLbShortName(client *duplosdk.Client, tenantID string, fullName string) (string, error) {
	prefix, err := client.GetResourcePrefix("duplo3", tenantID)
	if err != nil {
		return "", err
	}
	name, _ := duplosdk.UnprefixName(prefix, fullName)
	return name, nil
}
