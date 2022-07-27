package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const BYOH_VAR_PREFIX = "byoh_"

type BYOH struct {
}

func (byoh *BYOH) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantByohList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	if list != nil {
		log.Println("[TRACE] <====== BYOH TF generation started. =====>")
		for _, byoh := range *list {
			shortName := byoh.Name
			log.Printf("[TRACE] Generating terraform config for duplo byoh Instance : %s", shortName)

			varFullPrefix := BYOH_VAR_PREFIX + strings.ReplaceAll(shortName, "-", "_") + "_"
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "byoh-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_ecache_instance resource
			byohBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_byoh",
					shortName})
			byohBody := byohBlock.Body()
			byohBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			byohBody.SetAttributeValue("name",
				cty.StringVal(shortName))
			byohBody.SetAttributeValue("direct_address",
				cty.StringVal(byoh.DirectAddress))
			byohBody.SetAttributeValue("agent_platform",
				cty.NumberIntVal(int64(byoh.AgentPlatform)))
			if len(*byoh.Tags) > 0 {
				for _, tag := range *byoh.Tags {
					if tag.Key == "AllocationTags" {
						byohBody.SetAttributeValue("allocation_tag",
							cty.StringVal(tag.Value))
						break
					}
				}
			}
			cred, err := client.TenantHostCredentialsGet(config.TenantId, duplosdk.DuploHostOOBData{
				IPAddress: byoh.DirectAddress,
				Cloud:     4,
			})
			if err != nil {
				// TODO - Fix backend API for missing data.
				log.Printf("[TRACE] Error : %s", err)
			}

			if cred != nil {
				if len(cred.Username) > 0 {
					byohBody.SetAttributeValue("username",
						cty.StringVal(cred.Username))
				}
				if len(cred.Password) > 0 {
					byohBody.SetAttributeValue("password",
						cty.StringVal(cred.Password))
				}
				if len(cred.Privatekey) > 0 {
					byohBody.SetAttributeValue("private_key",
						cty.StringVal(cred.Privatekey))
				}
			}
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			log.Printf("[TRACE] Terraform config is generated for duplo BYOH instance : %s", shortName)

			outVars := generateBYOHOutputVars(varFullPrefix, shortName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs := []common.ImportConfig{}
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_byoh." + shortName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== BYOH TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateBYOHOutputVars(prefix, shortName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "connection_url",
		ActualVal:     "duplocloud_byoh." + shortName + ".connection_url",
		DescVal:       "The connection url for BYOH instance.",
		RootTraversal: true,
	}
	outVarConfigs["connection_url"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "network_agent_url",
		ActualVal:     "duplocloud_byoh." + shortName + ".network_agent_url",
		DescVal:       "The network agent url for BYOH instance.",
		RootTraversal: true,
	}
	outVarConfigs["network_agent_url"] = var2

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
