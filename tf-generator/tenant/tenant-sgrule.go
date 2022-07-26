package tenant

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

type TenantSGRule struct {
}

func (tsgrule *TenantSGRule) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.TenantProject)

	list, clientErr := client.TenantGetExtConnSecurityGroupRules(config.TenantId)
	//Get tenant from duplo
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	if len(*list) > 0 {
		log.Println("[TRACE] <====== Tenant SG rule TF generation started. =====>")
		// create new empty hcl file object
		hclFile := hclwrite.NewEmptyFile()

		// create new file on system
		path := filepath.Join(workingDir, "tenant-sg-rules.tf")
		tfFile, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		// initialize the body of the new file object
		var rootBody *hclwrite.Body
		rootBodyCreated := false
		importConfigs := []common.ImportConfig{}
		for i, sgRule := range *list {
			if !rootBodyCreated {
				rootBody = hclFile.Body()
			}
			rootBodyCreated = true
			for j, source := range *sgRule.Sources {
				tenantSgRule := rootBody.AppendNewBlock("resource",
					[]string{"duplocloud_tenant_network_security_rule",
						"tenant-sg-rule" + strconv.Itoa(i+1+j)})
				tenantSgRuleBody := tenantSgRule.Body()
				tenantSgRuleBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "duplocloud_tenant.tenant",
					},
					hcl.TraverseAttr{
						Name: "tenant_id",
					},
				})
				tenantSgRuleBody.SetAttributeValue("protocol",
					cty.StringVal(sgRule.Protocol))
				var sourceType string
				if source.Type == duplosdk.SGSourceTypeTenant {
					tenantSgRuleBody.SetAttributeValue("source_tenant",
						cty.StringVal(source.Value))
					sourceType = "source_tenant"
				} else {
					tenantSgRuleBody.SetAttributeValue("source_address",
						cty.StringVal(source.Value))
					sourceType = "source_address"
				}
				tenantSgRuleBody.SetAttributeValue("from_port",
					cty.NumberIntVal(int64(sgRule.FromPort)))
				tenantSgRuleBody.SetAttributeValue("to_port",
					cty.NumberIntVal(int64(sgRule.ToPort)))
				tenantSgRuleBody.SetAttributeValue("description",
					cty.StringVal(source.Description))
				rootBody.AppendNewline()

				if config.GenerateTfState {
					importConfigs = append(importConfigs, common.ImportConfig{
						ResourceAddress: "duplocloud_tenant_network_security_rule.tenant-sg-rule" + strconv.Itoa(i+1+j),
						ResourceId:      config.TenantId + "/" + strconv.Itoa(sgRule.Type) + "/" + sourceType + "/" + sgRule.Protocol + "/" + strconv.Itoa(sgRule.FromPort) + "/" + strconv.Itoa(sgRule.ToPort),
						WorkingDir:      workingDir,
					})
				}
			}
		}
		tfContext.ImportConfigs = importConfigs
		if rootBodyCreated {
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		}

		log.Println("[TRACE] <====== Tenant SG rules TF generation done. =====>")

	}
	return &tfContext, nil
}
