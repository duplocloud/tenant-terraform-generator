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

type PlanConfig struct {
	InfraName string
}

func (p PlanConfig) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	workingDir := filepath.Join(config.AdminInfraDir, config.AdminInfra)
	planConfig, clientErr := client.PlanConfigGetList(p.InfraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	tfContext := common.TFContext{}
	hclFile := hclwrite.NewEmptyFile()

	// create new file on system
	path := filepath.Join(workingDir, p.InfraName+"_plan.tf")
	tfFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resourceName := common.GetResourceName(p.InfraName)
	rootBody := hclFile.Body()

	// initialize the body of the new file object
	planBlock := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_plan_configs", "plan_config"})

	planBody := planBlock.Body()
	planBody.SetAttributeValue("plan_id", cty.StringVal(p.InfraName))
	planBody.SetAttributeValue("delete_unspecified_configs", cty.BoolVal(false))
	if planConfig != nil && len(*planConfig) > 0 {
		for _, v := range *planConfig {
			conf := planBody.AppendNewBlock("config", nil).Body()
			conf.SetAttributeValue("key", cty.StringVal(v.Key))
			conf.SetAttributeValue("value", cty.StringVal(v.Value))
			conf.SetAttributeValue("type", cty.StringVal(v.Type))

		}
	}
	_, err = tfFile.Write(hclFile.Bytes())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if config.GenerateTfState {
		importConfigs := []common.ImportConfig{}
		importConfigs = append(importConfigs, common.ImportConfig{
			ResourceAddress: "duplocloud_plan_configs." + resourceName,
			ResourceId:      p.InfraName,
			WorkingDir:      workingDir,
		})
		tfContext.ImportConfigs = importConfigs
	}

	return &tfContext, nil
}
