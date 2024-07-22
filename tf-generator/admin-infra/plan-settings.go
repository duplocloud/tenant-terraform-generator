package admininfra

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type PlanSetting struct {
	InfraName string
}

const (
	SETTING_PREFIX = "plan_setting_"
)

func (p PlanSetting) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	infraName := config.DuploPlanId
	workingDir := filepath.Join(config.InfraDir, config.Infra)
	planSetting, clientErr := client.PlanGetSettings(infraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	planDns, clientErr := client.PlanGetDnsConfig(infraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	//planMeta, clientErr := client.PlanMetadataGetList(infraName)
	//if clientErr != nil {
	//	return nil, errors.New(clientErr.Error())
	//}
	tfContext := common.TFContext{}
	inputVars := generateSettingVars(*planDns, SETTING_PREFIX)

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
	rootBody.AppendUnstructuredTokens(hclwrite.Tokens{
		&hclwrite.Token{Type: hclsyntax.TokenComment, Bytes: []byte("# Uncomment if used to setup non default DNS configuration\n/*")},
	})

	planBlock := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_plan_settings", "plan_setting"})

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
	planBody.SetAttributeValue("unrestricted_ext_lb", cty.BoolVal(planSetting.UnrestrictedExtLB))

	dnsConfig := planBody.AppendNewBlock("dns_setting", nil).Body()
	dnsConfig.SetAttributeTraversal("domain_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "domain_id",
		},
	})
	dnsConfig.SetAttributeTraversal("internal_dns_suffix", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "internal_dns_suffix",
		},
	})

	dnsConfig.SetAttributeTraversal("external_dns_suffix", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "external_dns_suffix",
		},
	})

	dnsConfig.SetAttributeTraversal("ignore_global_dns", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "ignore_global_dns",
		},
	})

	/*if planMeta != nil && len(*planMeta) > 0 {
		for _, v := range *planMeta {
			conf := planBody.AppendNewBlock("metadata", nil).Body()
			conf.SetAttributeValue("key", cty.StringVal(v.Key))
			conf.SetAttributeValue("value", cty.StringVal(v.Value))
		}

	}*/
	planBody.SetAttributeTraversal("depends_on", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "[duplocloud_infrastructure",
		},
		hcl.TraverseAttr{
			Name: "infra]",
		},
	})
	rootBody.AppendUnstructuredTokens(hclwrite.Tokens{
		&hclwrite.Token{Type: hclsyntax.TokenComment, Bytes: []byte("*/")},
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
			ResourceAddress: "duplocloud_plan_settings." + resourceName,
			ResourceId:      infraName,
			WorkingDir:      workingDir,
		})
		tfContext.ImportConfigs = importConfigs
	}

	return &tfContext, nil
}

func generateSettingVars(duplo duplosdk.DuploPlanDnsConfig, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig, 0)
	var1 := common.VarConfig{
		Name:       prefix + "domain_id",
		DefaultVal: duplo.DomainId,
		TypeVal:    "string",
	}
	varConfigs["domain_id"] = var1

	var2 := common.VarConfig{
		Name:       prefix + "internal_dns_suffix",
		DefaultVal: duplo.DomainId,
		TypeVal:    "string",
	}
	varConfigs["internal_dns_suffix"] = var2

	var3 := common.VarConfig{
		Name:       prefix + "external_dns_suffix",
		DefaultVal: duplo.DomainId,
		TypeVal:    "string",
	}
	varConfigs["external_dns_suffix"] = var3

	var4 := common.VarConfig{
		Name:       prefix + "ignore_global_dns",
		DefaultVal: duplo.DomainId,
		TypeVal:    "bool",
	}
	varConfigs["ignore_global_dns"] = var4

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	return vars
}
