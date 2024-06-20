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

type PlanImage struct {
	InfraName string
}

func (p PlanImage) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	workingDir := filepath.Join(config.AdminInfraDir, config.AdminInfra)
	planImg, clientErr := client.PlanImageGetList(p.InfraName)
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
	// initialize the body of the new file object
	for _, val := range *planImg {
		planBlock := rootBody.AppendNewBlock("resource",
			[]string{"duplocloud_plan_image", "image-" + val.Name})
		planBody := planBlock.Body()
		planBody.SetAttributeValue("plan_id", cty.StringVal(p.InfraName))
		planBody.SetAttributeValue("delete_unspecified_images", cty.BoolVal(false))
		image := planBody.AppendNewBlock("image", nil).Body()
		image.SetAttributeValue("name", cty.StringVal(val.Name))
		image.SetAttributeValue("image_id", cty.StringVal(val.ImageId))
		image.SetAttributeValue("os", cty.StringVal(val.OS))
		image.SetAttributeValue("username", cty.StringVal(val.Username))
		for _, v := range *val.Tags {
			tag := planBody.AppendNewBlock("tags", nil).Body()
			tag.SetAttributeValue("key", cty.StringVal(v.Key))
			tag.SetAttributeValue("value", cty.StringVal(v.Value))
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
			ResourceAddress: "duplocloud_plan_certificate." + resourceName,
			ResourceId:      p.InfraName,
			WorkingDir:      workingDir,
		})
		tfContext.ImportConfigs = importConfigs
	}

	return &tfContext, nil
}
