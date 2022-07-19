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

const HOST_VAR_PREFIX = "host_"

type Hosts struct {
}

func (h *Hosts) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	log.Println("[TRACE] <====== Hosts TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.NativeHostGetList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	if list != nil {
		for _, host := range *list {
			shortName := host.FriendlyName[len("duploservices-"+config.TenantName+"-"):len(host.FriendlyName)]
			log.Printf("[TRACE] Generating terraform config for duplo host : %s", host.FriendlyName)
			if isPartOfAsg(host) {
				continue
			}
			varFullPrefix := HOST_VAR_PREFIX + strings.ReplaceAll(shortName, "-", "_") + "_"
			inputVars := generateHostVars(host, varFullPrefix)
			tfContext.InputVars = append(tfContext.InputVars, inputVars...)
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "host-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_host resource
			hostBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_host",
					shortName})
			hostBody := hostBlock.Body()
			hostBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// hostBody.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			hostBody.SetAttributeValue("friendly_name",
				cty.StringVal(shortName))
			hostBody.SetAttributeTraversal("image_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "image_id",
				},
			})
			hostBody.SetAttributeTraversal("capacity", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "capacity",
				},
			})
			hostBody.SetAttributeValue("agent_platform",
				cty.NumberIntVal(int64(host.AgentPlatform)))
			hostBody.SetAttributeValue("zone",
				cty.NumberIntVal(int64(host.Zone)))
			if len(host.UserAccount) > 0 {
				hostBody.SetAttributeValue("user_account",
					cty.StringVal(host.UserAccount))
			}
			hostBody.SetAttributeValue("is_minion",
				cty.BoolVal(host.IsMinion))
			hostBody.SetAttributeValue("is_ebs_optimized",
				cty.BoolVal(host.IsEbsOptimized))
			hostBody.SetAttributeValue("encrypt_disk",
				cty.BoolVal(host.EncryptDisk))
			hostBody.SetAttributeValue("allocated_public_ip",
				cty.BoolVal(host.AllocatedPublicIP))
			hostBody.SetAttributeValue("cloud",
				cty.NumberIntVal(int64(host.Cloud)))
			if len(host.Base64UserData) > 0 {
				hostBody.SetAttributeValue("base64_user_data",
					cty.StringVal(host.Base64UserData))
			}

			if host.MinionTags != nil {
				for _, duploObject := range *host.MinionTags {
					if len(duploObject.Value) > 0 {
						minionTagsBlock := hostBody.AppendNewBlock("minion_tags",
							nil)
						minionTagsBody := minionTagsBlock.Body()
						minionTagsBody.SetAttributeValue("key",
							cty.StringVal(duploObject.Key))
						minionTagsBody.SetAttributeValue("value",
							cty.StringVal(duploObject.Value))
						rootBody.AppendNewline()
					}
				}
			}
			//TODO - Duplo provider doesn't handle this yet.
			// if host.MetaData != nil {
			// 	for _, duploObject := range *host.MetaData {
			// 		mdBlock := hostBody.AppendNewBlock("metadata",
			// 			nil)
			// 		mdBody := mdBlock.Body()
			// 		mdBody.SetAttributeValue("key",
			// 			cty.StringVal(duploObject.Key))
			// 		mdBody.SetAttributeValue("value",
			// 			cty.StringVal(duploObject.Value))
			// 		rootBody.AppendNewline()
			// 	}
			// }
			// TODO - Duplo provider doesn't handle this yet.
			// if host.Volumes != nil {
			// 	for _, duploObject := range *host.Volumes {
			// 		volumeBlock := hostBody.AppendNewBlock("volume",
			// 			nil)
			// 		volumeBody := volumeBlock.Body()
			// 		volumeBody.SetAttributeValue("iops",
			// 			cty.NumberIntVal(int64(duploObject.Iops)))
			// 		volumeBody.SetAttributeValue("name",
			// 			cty.StringVal(duploObject.Name))
			// 		volumeBody.SetAttributeValue("size",
			// 			cty.NumberIntVal(int64(duploObject.Size)))
			// 		volumeBody.SetAttributeValue("volume_id",
			// 			cty.StringVal(duploObject.VolumeID))
			// 		volumeBody.SetAttributeValue("volume_type",
			// 			cty.StringVal(duploObject.VolumeType))
			// 		rootBody.AppendNewline()
			// 	}
			// }
			// TODO - Handle tags, network_interface
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo host : %s", host.FriendlyName)

			outVars := generateHostOutputVars(host, varFullPrefix, shortName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs := []common.ImportConfig{}
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_host." + shortName,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/NativeHostV2/" + host.InstanceID,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
	}
	log.Println("[TRACE] <====== Hosts TF generation done. =====>")
	return &tfContext, nil
}

func isPartOfAsg(host duplosdk.DuploNativeHost) bool {
	asgTagKey := []string{"aws:autoscaling:groupName"}
	if host.Tags != nil && len(*host.Tags) > 0 {
		asgTag := duplosdk.SelectKeyValues(host.Tags, asgTagKey)
		if asgTag != nil && len(*asgTag) > 0 {
			return true
		}
	}
	return false
}

func generateHostVars(duplo duplosdk.DuploNativeHost, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	imageIdVar := common.VarConfig{
		Name:       prefix + "image_id",
		DefaultVal: duplo.ImageID,
		TypeVal:    "string",
	}
	varConfigs["image_id"] = imageIdVar

	capacityVar := common.VarConfig{
		Name:       prefix + "capacity",
		DefaultVal: duplo.Capacity,
		TypeVal:    "string",
	}
	varConfigs["capacity"] = capacityVar

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}

func generateHostOutputVars(duplo duplosdk.DuploNativeHost, prefix, shortName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "instance_id",
		ActualVal:     "duplocloud_aws_host." + shortName + ".instance_id",
		DescVal:       "The AWS EC2 instance ID of the host.",
		RootTraversal: true,
	}
	outVarConfigs["instance_id"] = var1
	var2 := common.OutputVarConfig{
		Name:          prefix + "private_ip_address",
		ActualVal:     "duplocloud_aws_host." + shortName + ".private_ip_address",
		DescVal:       "The primary private IP address assigned to the host.",
		RootTraversal: true,
	}
	outVarConfigs["private_ip_address"] = var2

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
