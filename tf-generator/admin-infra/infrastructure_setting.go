package admininfra

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type InfraSetting struct{}

const INFRA_SETTING_VAR_PREFIX = "infra-setting"

func (i InfraSetting) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	//infras, _ := client.InfrastructureGetList()
	workingDir := filepath.Join(config.InfraDir, config.Infra)

	tfContext := common.TFContext{}
	//if infras != nil {
	//for _, v := range *infras {
	infra, clientErr := client.InfrastructureGetConfig(config.DuploPlanId)
	if clientErr != nil {
		log.Printf("Error while fetching infra %s : %s", config.DuploPlanId, clientErr.Error())
		return nil, fmt.Errorf(clientErr.Error())
	}
	hclFile := hclwrite.NewEmptyFile()

	// create new file on system
	path := filepath.Join(workingDir, "main.tf")
	tfFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	resourceName := common.GetResourceName(infra.Name)
	inputVars := generateInfraSettingVars(infra, INFRA_SETTING_VAR_PREFIX)
	tfContext.InputVars = append(tfContext.InputVars, inputVars...)
	rootBody := hclFile.Body()

	// initialize the body of the new file object
	infraBlock := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_infrastructure_setting", "setting"})

	infraBody := infraBlock.Body()
	infraBody.SetAttributeTraversal("infra_name", hcl.Traversal{
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

	if infra.CustomData != nil && len(*infra.CustomData) > 0 {
		content := infraBody.AppendNewBlock("dynamic", []string{"setting"}).Body()
		setting := content.AppendNewBlock("content", nil).Body()
		setting.SetAttributeTraversal("key", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "setting",
			},
			hcl.TraverseAttr{
				Name: "value.key",
			},
		})
		setting.SetAttributeTraversal("value", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "setting",
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
				Name: INFRA_SETTING_VAR_PREFIX + "_settings",
			},
		})
	}
	infraBody.SetAttributeTraversal("depends_on", hcl.Traversal{
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
			ResourceAddress: "duplocloud_infrastructure." + resourceName,
			ResourceId:      "v2/admin/InfrastructureV2/" + infra.Name,
			WorkingDir:      workingDir,
		})
		tfContext.ImportConfigs = importConfigs
	}

	//}
	//}
	return &tfContext, nil
}

func generateInfraSettingVars(duplo *duplosdk.DuploInfrastructureConfig, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig, 0)

	keyval := setKeyValueVar(*duplo.CustomData)
	arr, err := json.Marshal(&keyval)
	if err != nil {
		log.Fatal(err)
	}

	var6 := common.VarConfig{
		Name:       prefix + "_settings",
		DefaultVal: string(arr),
		TypeVal: `list(object({
			key = string
			value = string
		  }))`,
	}
	varConfigs[prefix] = var6

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	return vars
}

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func setKeyValueVar(duplo []duplosdk.DuploKeyStringValue) []keyValue {
	ar := []keyValue{}
	for _, data := range duplo {
		ar = append(ar, keyValue{Key: data.Key, Value: data.Value})
	}
	return ar
}
