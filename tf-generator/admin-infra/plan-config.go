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

type PlanConfig struct {
}

func (p PlanConfig) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	infraName := config.DuploPlanId
	workingDir := filepath.Join(config.InfraDir, config.Infra)
	planConfig, clientErr := client.PlanConfigGetList(infraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	if len(*planConfig) == 0 {
		return nil, nil
	}
	tfContext := common.TFContext{}
	hclFile := hclwrite.NewEmptyFile()
	inputVars := generateConfigVars(*planConfig, "configs")

	// create new file on system
	path := filepath.Join(workingDir, "plan.tf")
	tfFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resourceName := common.GetResourceName(infraName)
	rootBody := hclFile.Body()
	tfContext.InputVars = append(tfContext.InputVars, inputVars...)
	// initialize the body of the new file object
	planBlock := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_plan_configs", "plan_config"})

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

	planBody.SetAttributeValue("delete_unspecified_configs", cty.BoolVal(false))

	if planConfig != nil && len(*planConfig) > 0 {
		content := planBody.AppendNewBlock("dynamic", []string{"config"}).Body()
		conf := content.AppendNewBlock("content", nil).Body()
		conf.SetAttributeTraversal("config", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "config",
			},
			hcl.TraverseAttr{
				Name: "value.key",
			},
		})
		conf.SetAttributeTraversal("type", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "config",
			},
			hcl.TraverseAttr{
				Name: "value.type",
			},
		})
		conf.SetAttributeTraversal("value", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "config",
			},
			hcl.TraverseAttr{
				Name: "value.value",
			},
		})
		content.SetAttributeTraversal("for_each", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "var",
			},
			hcl.TraverseAttr{
				Name: "configs",
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
			ResourceAddress: "duplocloud_plan_configs." + resourceName,
			ResourceId:      infraName,
			WorkingDir:      workingDir,
		})
		tfContext.ImportConfigs = importConfigs
	}

	return &tfContext, nil
}

type varConf struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func setVarConfigs(duplo []duplosdk.DuploCustomDataEx) []varConf {
	ar := []varConf{}
	for _, data := range duplo {
		ar = append(ar, varConf{Key: data.Key, Type: data.Type, Value: data.Value})
	}
	return ar
}

func generateConfigVars(duplo []duplosdk.DuploCustomDataEx, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig, 0)
	value := setVarConfigs(duplo)
	conf, err := json.Marshal(&value)
	if err != nil {
		log.Fatal(err)
	}
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
