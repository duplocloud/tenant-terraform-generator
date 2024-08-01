package admininfra

import (
	"encoding/json"
	"errors"
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

type PlanImage struct {
}

var varFullPrefix = "plan_image"

func (p PlanImage) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	infraName := config.DuploPlanId
	workingDir := filepath.Join(config.InfraDir, config.Infra)
	planImg, clientErr := client.PlanImageGetList(infraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	if len(*planImg) == 0 {
		return nil, nil
	}
	tfContext := common.TFContext{}
	inputVars := generateImageVars(*planImg, varFullPrefix)
	tfContext.InputVars = append(tfContext.InputVars, inputVars...)

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
	planBlock := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_plan_images", varFullPrefix})
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

	if planImg != nil && len(*planImg) > 0 {
		content := planBody.AppendNewBlock("dynamic", []string{"image"}).Body()
		imgs := content.AppendNewBlock("content", nil).Body()
		imgs.SetAttributeTraversal("name", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "image",
			},
			hcl.TraverseAttr{
				Name: "value.name",
			},
		})
		imgs.SetAttributeTraversal("image_id", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "image",
			},
			hcl.TraverseAttr{
				Name: "value.image_id",
			},
		})
		imgs.SetAttributeTraversal("os", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "os",
			},
			hcl.TraverseAttr{
				Name: "value.os",
			},
		})
		imgs.SetAttributeTraversal("username", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "username",
			},
			hcl.TraverseAttr{
				Name: "value.username",
			},
		})
		content.SetAttributeTraversal("for_each", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "var",
			},
			hcl.TraverseAttr{
				Name: "images",
			},
		})
	}
	planBody.SetAttributeTraversal("depends_on", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "[duplocloud_infrastructure",
		},
		hcl.TraverseAttr{
			Name: "infra]",
		},
	})

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

type varImage struct {
	Name     string `json:"name"`
	ImageId  string `json:"image_id"`
	OS       string `json:"os"`
	UserName string `json:"username"`
}

func setVarImage(duplo []duplosdk.DuploPlanImage) []varImage {
	imageVars := []varImage{}
	for _, v := range duplo {
		imageVars = append(imageVars, varImage{
			Name:     v.Name,
			ImageId:  v.ImageId,
			OS:       v.OS,
			UserName: v.Username,
		})
	}
	return imageVars
}
func generateImageVars(duplo []duplosdk.DuploPlanImage, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig, 0)
	value := setVarImage(duplo)
	image, err := json.Marshal(&value)
	if err != nil {
		log.Fatal(err)
	}

	var1 := common.VarConfig{
		Name:       "images",
		DefaultVal: string(image),
		TypeVal: `list(object({
			name = string
			image_id = string
			os = string
			username=string
		  }))`,
	}
	varConfigs["images"] = var1

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	return vars
}
