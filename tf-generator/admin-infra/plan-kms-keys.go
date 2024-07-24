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
)

type PlanKms struct {
}

var KMSPrefix = "kmsKeys"

func (p PlanKms) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	infraName := config.DuploPlanId
	workingDir := filepath.Join(config.InfraDir, config.Infra)
	kmsKeys, clientErr := client.PlanKMSGetList(infraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	if len(*kmsKeys) == 0 {
		return nil, nil
	}
	tfContext := common.TFContext{}
	hclFile := hclwrite.NewEmptyFile()
	inputVars := generateKMSVars(*kmsKeys, KMSPrefix)
	tfContext.InputVars = append(tfContext.InputVars, inputVars...)
	// create new file on system
	path := filepath.Join(workingDir, "plan.tf")
	tfFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	resourceName := common.GetResourceName(infraName)
	rootBody := hclFile.Body()
	planBlock := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_plan_kms_v2", "kms"})

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
	if kmsKeys != nil && len(*kmsKeys) > 0 {

		content := planBody.AppendNewBlock("dynamic", []string{"kms"}).Body()
		kms := content.AppendNewBlock("content", nil).Body()
		kms.SetAttributeTraversal("name", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "kms",
			},
			hcl.TraverseAttr{
				Name: "value.name",
			},
		})
		kms.SetAttributeTraversal("id", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "kms",
			},
			hcl.TraverseAttr{
				Name: "value.id",
			},
		})
		content.SetAttributeTraversal("for_each", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "var",
			},
			hcl.TraverseAttr{
				Name: KMSPrefix,
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
			ResourceAddress: "duplocloud_plan_kms." + resourceName,
			ResourceId:      infraName + "/kms",
			WorkingDir:      workingDir,
		})
		tfContext.ImportConfigs = importConfigs
	}

	return &tfContext, nil
}

type varKms struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Arn  string `json:"arn"`
}

func setVarKms(duplo []duplosdk.DuploPlanKmsKeyInfo) []varKms {
	ar := []varKms{}
	for _, data := range duplo {
		ar = append(ar, varKms{Id: data.KeyId, Name: data.KeyName, Arn: data.KeyArn})
	}
	return ar
}

func generateKMSVars(duplo []duplosdk.DuploPlanKmsKeyInfo, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig, 0)
	value := setVarKms(duplo)
	certs, err := json.Marshal(&value)
	if err != nil {
		log.Fatal(err)
	}
	var1 := common.VarConfig{
		Name:       prefix,
		DefaultVal: string(certs),
		TypeVal: `list(object({
			id = string
			name = string
		  }))`,
	}
	varConfigs[prefix] = var1

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	return vars
}
