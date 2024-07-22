package admininfra

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type PlanImage struct {
}

var varFullPrefix = "image_"

func (p PlanImage) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	infraName := config.DuploPlanId
	workingDir := filepath.Join(config.InfraDir, config.Infra)
	planImg, clientErr := client.PlanImageGetList(infraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	tfContext := common.TFContext{}
	hclFile := hclwrite.NewEmptyFile()

	// create new file on system
	path := filepath.Join(workingDir, "plan.tf")
	tfFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	resourceName := common.GetResourceName(infraName)
	rootBody := hclFile.Body()
	// initialize the body of the new file object
	for _, val := range *planImg {
		planBlock := rootBody.AppendNewBlock("resource",
			[]string{"duplocloud_plan_image", "image-" + val.Name})
		planBody := planBlock.Body()
		planBody.SetAttributeTraversal("plan_id", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "duplocloud_infrastructure",
			},
			hcl.TraverseAttr{
				Name: "infra",
			},
			hcl.TraverseAttr{
				Name: "infra_name",
			},
		})
		planBody.SetAttributeValue("delete_unspecified_images", cty.BoolVal(false))
		image := planBody.AppendNewBlock("image", nil).Body()
		image.SetAttributeTraversal("name", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "var",
			},
			hcl.TraverseAttr{
				Name: varFullPrefix + "name",
			},
		})
		image.SetAttributeTraversal("image_id", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "var",
			},
			hcl.TraverseAttr{
				Name: "image_id",
			},
		})

		image.SetAttributeTraversal("os", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "var",
			},
			hcl.TraverseAttr{
				Name: varFullPrefix + "os",
			},
		})

		image.SetAttributeTraversal("username", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "var",
			},
			hcl.TraverseAttr{
				Name: varFullPrefix + "username",
			},
		})

		planBody.SetAttributeTraversal("depends_on", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "[duplocloud_infrastructure",
			},
			hcl.TraverseAttr{
				Name: "infra]",
			},
		})

	}

	rootBody.AppendNewline()

	_, err = tfFile.Write(hclFile.Bytes())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if config.GenerateTfState {
		importConfigs := []common.ImportConfig{}
		importConfigs = append(importConfigs, common.ImportConfig{
			ResourceAddress: "duplocloud_plan_image." + resourceName,
			ResourceId:      infraName,
			WorkingDir:      workingDir,
		})
		tfContext.ImportConfigs = importConfigs
	}

	return &tfContext, nil
}

func generateImageVars(duplo []duplosdk.DuploCustomDataEx, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig, 0)
	var1 := common.VarConfig{
		Name:       prefix,
		DefaultVal: string(conf),
		TypeVal: `list(object({
			key = string
			type = string
			value = string
		  }))`,
	}
	varConfigs[prefix] = var1

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	return vars
}
