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

type PlanWaf struct {
}

var WAF_PREFIX = "wafs"

func (p PlanWaf) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	infraName := config.DuploPlanId
	workingDir := filepath.Join(config.InfraDir, config.Infra)
	planWAF, clientErr := client.PlanWAFGetList(infraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	if len(*planWAF) == 0 {
		return nil, nil
	}
	tfContext := common.TFContext{}
	hclFile := hclwrite.NewEmptyFile()
	inputVars := generateWAFVars(*planWAF, WAF_PREFIX)
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
		[]string{"duplocloud_plan_waf_v2", "waf"})

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

	if planWAF != nil && len(*planWAF) > 0 {
		content := planBody.AppendNewBlock("dynamic", []string{"waf"}).Body()
		w := content.AppendNewBlock("content", nil).Body()
		w.SetAttributeTraversal("name", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "waf",
			},
			hcl.TraverseAttr{
				Name: "value.name",
			},
		})
		w.SetAttributeTraversal("arn", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "waf",
			},
			hcl.TraverseAttr{
				Name: "value.arn",
			},
		})
		w.SetAttributeTraversal("dashboard_url", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "waf",
			},
			hcl.TraverseAttr{
				Name: "value.dashboard_url",
			},
		})
		content.SetAttributeTraversal("for_each", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "var",
			},
			hcl.TraverseAttr{
				Name: WAF_PREFIX,
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
			ResourceAddress: "duplocloud_plan_waf." + resourceName,
			ResourceId:      infraName + "/waf",
			WorkingDir:      workingDir,
		})
		tfContext.ImportConfigs = importConfigs
	}

	return &tfContext, nil
}

type varWaf struct {
	Id           string `json:"arn"`
	Name         string `json:"name"`
	DashboardUrl string `json:"dashboard_url"`
}

func setVarWaf(duplo []duplosdk.PlanWAF) []varWaf {
	ar := []varWaf{}
	for _, data := range duplo {
		ar = append(ar, varWaf{Id: data.WebAclId, Name: data.WebAclName, DashboardUrl: data.DashboardUrl})
	}
	return ar
}

func generateWAFVars(duplo []duplosdk.PlanWAF, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig, 0)
	value := setVarWaf(duplo)
	certs, err := json.Marshal(&value)
	if err != nil {
		log.Fatal(err)
	}
	var1 := common.VarConfig{
		Name:       prefix,
		DefaultVal: string(certs),
		TypeVal: `list(object({
			arn = string
			name = string
			dashboard_url=string
		  }))`,
	}
	varConfigs[prefix+"_list"] = var1

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	return vars
}
