package adminproject

import (
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

type Infra struct{}

const INFRA_VAR_PREFIX = "infra_"

func (i Infra) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	infras, _ := client.InfrastructureGetList()
	workingDir := filepath.Join(config.AdminProjectDir, config.AdminProject)

	tfContext := common.TFContext{}
	if infras != nil {
		for _, v := range *infras {
			infra, clientErr := client.InfrastructureGetConfig(v.Name)
			if clientErr != nil {
				log.Printf("Error while fetching infra %s : %s", v.Name, clientErr.Error())
				continue
			}

			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, infra.Name+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			resourceName := common.GetResourceName(infra.Name)
			varFullPrefix := INFRA_VAR_PREFIX + resourceName + "_"
			inputVars := generateInfraVars(infra, varFullPrefix)
			tfContext.InputVars = append(tfContext.InputVars, inputVars...)
			fmt.Println("Admin generator tfContext.InputVars ", tfContext.InputVars)
			rootBody := hclFile.Body()

			// initialize the body of the new file object
			infraBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_infrastructure", "infra"})

			infraBody := infraBlock.Body()
			infraBody.SetAttributeValue("infra_name", cty.StringVal(infra.Name))
			infraBody.SetAttributeTraversal("account_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "account_id",
				},
			})
			infraBody.SetAttributeValue("cloud", cty.NumberIntVal(int64(infra.Cloud)))
			infraBody.SetAttributeTraversal("region", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "region",
				},
			})

			infraBody.SetAttributeValue("azcount", cty.NumberIntVal(int64(infra.AzCount)))
			infraBody.SetAttributeValue("enable_k8_cluster", cty.BoolVal(infra.EnableK8Cluster))
			infraBody.SetAttributeValue("is_serverless_kubernetes", cty.BoolVal(infra.IsServerlessKubernetes))
			infraBody.SetAttributeValue("enable_ecs_cluster", cty.BoolVal(infra.EnableECSCluster))
			infraBody.SetAttributeValue("enable_container_insights", cty.BoolVal(infra.EnableContainerInsights))
			infraBody.SetAttributeTraversal("address_prefix", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "address_prefix",
				},
			})

			infraBody.SetAttributeValue("subnet_cidr", cty.NumberIntVal(int64(infra.Vnet.SubnetCidr)))
			for _, cd := range *infra.CustomData {
				allsettingBody := infraBody.AppendNewBlock("custom_data", nil).Body()
				allsettingBody.SetAttributeValue("key", cty.StringVal(cd.Key))
				allsettingBody.SetAttributeValue("value", cty.StringVal(cd.Value))

			}

			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			is := InfraSubnet{
				InfraName: v.Name,
				Subnets:   *infra.Vnet.Subnets,
			}
			_, err = is.Generate(config, nil)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			p := PlanConfig{
				InfraName: v.Name,
			}
			_, err = p.Generate(config, client)
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

		}
	}
	return &tfContext, nil
}

func generateInfraVars(duplo *duplosdk.DuploInfrastructureConfig, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	var1 := common.VarConfig{
		Name:       prefix + "account_id",
		DefaultVal: duplo.AccountId,
		TypeVal:    "string",
	}
	varConfigs["account_id"] = var1

	var2 := common.VarConfig{
		Name:       prefix + "region",
		DefaultVal: duplo.Region,
		TypeVal:    "string",
	}
	varConfigs["region"] = var2

	var3 := common.VarConfig{
		Name:       prefix + "address_prefix",
		DefaultVal: duplo.Vnet.AddressPrefix,
		TypeVal:    "string",
	}
	varConfigs["address_prefix"] = var3

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}
