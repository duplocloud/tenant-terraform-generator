package admininfra

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type Infra struct{}

const INFRA_VAR_PREFIX = "infra_"

func (i Infra) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
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
	tfFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	resourceName := common.GetResourceName(infra.Name)
	varFullPrefix := INFRA_VAR_PREFIX
	inputVars := generateInfraVars(infra, varFullPrefix)
	tfContext.InputVars = append(tfContext.InputVars, inputVars...)
	fmt.Println("Admin generator tfContext.InputVars ", tfContext.InputVars)
	rootBody := hclFile.Body()

	// initialize the body of the new file object
	infraBlock := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_infrastructure", "infra"})

	infraBody := infraBlock.Body()
	infraBody.SetAttributeTraversal("infra_name", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: varFullPrefix + "name",
		},
	})
	if infra.Cloud == 2 {
		infraBody.SetAttributeTraversal("account_id", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "var",
			},
			hcl.TraverseAttr{
				Name: varFullPrefix + "account_id",
			},
		})
	}
	infraBody.SetAttributeTraversal("cloud", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: varFullPrefix + "cloud",
		},
	})
	infraBody.SetAttributeTraversal("region", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "region",
		},
	})

	infraBody.SetAttributeTraversal("azcount", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: varFullPrefix + "azcount",
		},
	})

	infraBody.SetAttributeTraversal("enable_k8_cluster", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: varFullPrefix + "enable_k8_cluster",
		},
	})

	infraBody.SetAttributeTraversal("is_serverless_kubernetes", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: varFullPrefix + "is_serverless_kubernetes",
		},
	})

	infraBody.SetAttributeTraversal("enable_ecs_cluster", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: varFullPrefix + "enable_ecs_cluster",
		},
	})

	infraBody.SetAttributeTraversal("enable_container_insights", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: varFullPrefix + "enable_container_insights",
		},
	})

	infraBody.SetAttributeTraversal("address_prefix", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: varFullPrefix + "address_prefix",
		},
	})

	infraBody.SetAttributeTraversal("subnet_cidr", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: varFullPrefix + "subnet_cidr",
		},
	})

	rootBody.AppendNewline()
	_, err = tfFile.Write(hclFile.Bytes())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	is := InfraSubnet{
		InfraName: config.DuploPlanId,
		Subnets:   *infra.Vnet.Subnets,
	}
	_, err = is.Generate(config, nil)
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

func generateInfraVars(duplo *duplosdk.DuploInfrastructureConfig, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig, 0)

	var1 := common.VarConfig{
		Name:       prefix + "account_id",
		DefaultVal: duplo.AccountId,
		TypeVal:    "string",
	}
	varConfigs["account_id"] = var1

	var2 := common.VarConfig{
		Name:       "region",
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

	var4 := common.VarConfig{
		Name:       prefix + "name",
		DefaultVal: duplo.Vnet.Name,
		TypeVal:    "string",
	}
	varConfigs["infra_name"] = var4

	var5 := common.VarConfig{
		Name:       prefix + "cloud",
		DefaultVal: strconv.Itoa(duplo.Cloud),
		TypeVal:    "number",
	}
	varConfigs["cloud"] = var5
	subnets := *duplo.Vnet.Subnets
	cidr := strings.Split(subnets[0].AddressPrefix, "/")[1]
	var7 := common.VarConfig{
		Name:       prefix + "subnet_cidr",
		DefaultVal: cidr,
		TypeVal:    "number",
	}
	varConfigs["subnet_cidr"] = var7

	var9 := common.VarConfig{
		Name:       prefix + "enable_container_insights",
		DefaultVal: strconv.FormatBool(duplo.EnableContainerInsights),
		TypeVal:    "bool",
	}
	varConfigs["enable_container_insights"] = var9

	var10 := common.VarConfig{
		Name:       prefix + "enable_ecs_cluster",
		DefaultVal: strconv.FormatBool(duplo.EnableECSCluster),
		TypeVal:    "bool",
	}
	varConfigs["enable_ecs_cluster"] = var10

	var11 := common.VarConfig{
		Name:       prefix + "enable_k8_cluster",
		DefaultVal: strconv.FormatBool(duplo.EnableK8Cluster),
		TypeVal:    "bool",
	}
	varConfigs["enable_k8_cluster"] = var11

	var12 := common.VarConfig{
		Name:       prefix + "is_serverless_kubernetes",
		DefaultVal: strconv.FormatBool(duplo.IsServerlessKubernetes),
		TypeVal:    "bool",
	}
	varConfigs["is_serverless_kubernetes"] = var12

	var13 := common.VarConfig{
		Name:       prefix + "azcount",
		DefaultVal: strconv.Itoa(duplo.AzCount),
		TypeVal:    "number",
	}
	varConfigs["azcount"] = var13

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	return vars
}
