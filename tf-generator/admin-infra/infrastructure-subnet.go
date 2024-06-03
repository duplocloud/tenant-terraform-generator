package admininfra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type InfraSubnet struct {
	InfraName string
	Subnets   []duplosdk.DuploInfrastructureVnetSubnet
}

const INFRASUBNET_VAR_PREFIX = "infra_subnet_"

func (i InfraSubnet) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	workingDir := filepath.Join(config.AdminInfraDir, config.AdminInfra)

	tfContext := common.TFContext{}
	if i.Subnets != nil {
		for _, v := range i.Subnets {
			hclFile := hclwrite.NewEmptyFile()
			visiblity := "public"
			if strings.Contains(v.Name, "private") {
				visiblity = "private"
			}
			// create new file on system
			path := filepath.Join(workingDir, i.InfraName+".tf")
			tfFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			resourceName := common.GetResourceName(v.Name)
			//varFullPrefix := INFRASUBNET_VAR_PREFIX + resourceName + "_"
			//inputVars := generateInfraVars(v, varFullPrefix)
			//tfContext.InputVars = append(tfContext.InputVars, inputVars...)
			//fmt.Println("Admin generator tfContext.InputVars ", tfContext.InputVars)
			rootBody := hclFile.Body()

			// initialize the body of the new file object
			infraBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_infrastructure_subnet", "infra-" + v.Zone + "-" + visiblity})

			infraBody := infraBlock.Body()
			infraBody.SetAttributeValue("name", cty.StringVal(v.Name))
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
			infraBody.SetAttributeValue("cidr_block", cty.StringVal(v.AddressPrefix))
			infraBody.SetAttributeValue("type", cty.StringVal(visiblity))
			infraBody.SetAttributeValue("zone", cty.StringVal(v.Zone))
			infraBody.SetAttributeValue("isolated_network", cty.BoolVal(v.IsolatedNetwork))
			tagMp := make(map[string]string)
			for _, t := range *v.Tags {
				tagMp[t.Key] = t.Value
			}
			if len(v.ServiceEndpoints) > 0 {
				infraBody.SetAttributeValue("service_endpoints", cty.SetVal(common.StringSliceToListVal(v.ServiceEndpoints)))
			}
			infraBody.SetAttributeValue("tags", cty.MapVal(common.MapStringToMapVal(tagMp)))

			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			if config.GenerateTfState {
				importConfigs := []common.ImportConfig{}
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_infrastructure_subnet." + resourceName,
					ResourceId:      i.InfraName + "/" + v.Name + "/" + v.AddressPrefix,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}

		}
	}
	return &tfContext, nil
}
