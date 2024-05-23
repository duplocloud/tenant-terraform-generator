package adminproject

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

/*
"plan_id": {
	Description: "The ID of the plan to configure.",
	Type:        schema.TypeString,
	Required:    true,
	ForceNew:    true,
},
"config": {
	Description: "A list of configs to manage.",
	Type:        schema.TypeList,
	Optional:    true,
	Elem:        CustomDataExSchema(),
},
"delete_unspecified_configs": {
	Description: "Whether or not this resource should delete any configs not specified by this resource. " +
		"**WARNING:**  It is not recommended to change the default value of `false`.",
	Type:     schema.TypeBool,
	Optional: true,
	Default:  false,
*/

type PlanConfig struct {
	InfraName string
}

func (p PlanConfig) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	workingDir := filepath.Join(config.AdminProjectDir, config.AdminProject)
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

	palnBody := planBlock.Body()
	palnBody.SetAttributeValue("plan_id", cty.StringVal(p.InfraName))
	palnBody.SetAttributeValue("delete_unspecified_configs", cty.BoolVal(false))
	if planConfig != nil && len(*planConfig) > 0 {
		vals := make([]cty.Value, 0, len(*planConfig))
		for _, v := range *planConfig {
			m := make(map[string]string)
			m["key"] = v.Key
			m["value"] = v.Value
			m["type"] = v.Type
			om := common.MapStringToMapVal(m)
			val := cty.ObjectVal(om)
			vals = append(vals, val)
		}
		palnBody.SetAttributeValue("config", cty.ListVal(vals))
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
