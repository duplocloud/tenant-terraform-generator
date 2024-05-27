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

type PlanSetting struct {
	InfraName string
}

func (p PlanSetting) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	workingDir := filepath.Join(config.AdminProjectDir, config.AdminProject)
	planSetting, clientErr := client.PlanGetSettings(p.InfraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	planDns, clientErr := client.PlanGetDnsConfig(p.InfraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	planMeta, clientErr := client.PlanMetadataGetList(p.InfraName)
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
	planBlock := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_plan_settings", "plan_setting"})

	planBody := planBlock.Body()
	planBody.SetAttributeValue("plan_id", cty.StringVal(p.InfraName))
	planBody.SetAttributeValue("unrestricted_ext_lb", cty.BoolVal(planSetting.UnrestrictedExtLB))

	dnsConfig := planBody.AppendNewBlock("dns_setting", nil).Body()
	dnsConfig.SetAttributeValue("domain_id", cty.StringVal(planDns.DomainId))
	dnsConfig.SetAttributeValue("internal_dns_suffix", cty.StringVal(planDns.InternalDnsSuffix))
	dnsConfig.SetAttributeValue("external_dns_suffix", cty.StringVal(planDns.ExternalDnsSuffix))
	dnsConfig.SetAttributeValue("ignore_global_dns", cty.BoolVal(planDns.IgnoreGlobalDNS))

	if planMeta != nil && len(*planMeta) > 0 {
		for _, v := range *planMeta {
			conf := planBody.AppendNewBlock("metadata", nil).Body()
			conf.SetAttributeValue("key", cty.StringVal(v.Key))
			conf.SetAttributeValue("value", cty.StringVal(v.Value))
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
