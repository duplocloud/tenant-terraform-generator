package admininfra

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type PlanWaf struct {
	InfraName string
}

func (p PlanWaf) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	workingDir := filepath.Join(config.AdminInfraDir, config.AdminInfra)
	planWAF, clientErr := client.PlanWAFGetList(p.InfraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	tfContext := common.TFContext{}
	hclFile := hclwrite.NewEmptyFile()

	// create new file on system
	path := filepath.Join(workingDir, p.InfraName+"_plan.tf")
	tfFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	resourceName := common.GetResourceName(p.InfraName)
	rootBody := hclFile.Body()
	for _, waf := range *planWAF {
		// initialize the body of the new file object
		planBlock := rootBody.AppendNewBlock("resource",
			[]string{"duplocloud_plan_waf", "plan_waf" + waf.WebAclName})

		palnBody := planBlock.Body()
		palnBody.SetAttributeValue("plan_id", cty.StringVal(p.InfraName))
		palnBody.SetAttributeValue("waf_arn", cty.StringVal(waf.WebAclId))
		palnBody.SetAttributeValue("waf_name", cty.StringVal(waf.WebAclName))
		palnBody.SetAttributeValue("dashboard_url", cty.StringVal(waf.DashboardUrl))

		_, err = tfFile.Write(hclFile.Bytes())
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		if config.GenerateTfState {
			importConfigs := []common.ImportConfig{}
			importConfigs = append(importConfigs, common.ImportConfig{
				ResourceAddress: "duplocloud_plan_waf." + resourceName,
				ResourceId:      p.InfraName,
				WorkingDir:      workingDir,
			})
			tfContext.ImportConfigs = importConfigs
		}
	}
	return &tfContext, nil
}
