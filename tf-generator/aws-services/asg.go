package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const ASG_VAR_PREFIX = "asg_"

type ASG struct {
}

func (asg *ASG) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.AsgProfileGetList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== ASG TF generation started. =====>")
		for _, asgProfile := range *list {
			shortName := asgProfile.FriendlyName[len("duploservices-"+config.TenantName+"-"):len(asgProfile.FriendlyName)]
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo ASG : %s", asgProfile.FriendlyName)
			varFullPrefix := ASG_VAR_PREFIX + resourceName + "_"

			hclFile := hclwrite.NewEmptyFile()
			inputVars := generateAsgVars(asgProfile, varFullPrefix)
			tfContext.InputVars = append(tfContext.InputVars, inputVars...)
			// create new file on system

			path := filepath.Join(workingDir, "asg-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			asgBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_asg_profile",
					resourceName})
			asgBody := asgBlock.Body()
			asgBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// asgBody.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			asgBody.SetAttributeValue("friendly_name",
				cty.StringVal(shortName))
			asgBody.SetAttributeTraversal("instance_count", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "instance_count",
				},
			})
			asgBody.SetAttributeTraversal("min_instance_count", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "min_instance_count",
				},
			})
			asgBody.SetAttributeTraversal("max_instance_count", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "max_instance_count",
				},
			})
			asgBody.SetAttributeTraversal("image_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "image_id",
				},
			})
			asgBody.SetAttributeTraversal("capacity", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "capacity",
				},
			})
			asgBody.SetAttributeValue("agent_platform",
				cty.NumberIntVal(int64(asgProfile.AgentPlatform)))
			asgBody.SetAttributeValue("zone",
				cty.NumberIntVal(int64(asgProfile.Zone)))
			asgBody.SetAttributeValue("is_minion",
				cty.BoolVal(asgProfile.IsMinion))
			asgBody.SetAttributeValue("is_ebs_optimized",
				cty.BoolVal(asgProfile.IsEbsOptimized))
			asgBody.SetAttributeValue("encrypt_disk",
				cty.BoolVal(asgProfile.EncryptDisk))
			asgBody.SetAttributeValue("allocated_public_ip",
				cty.BoolVal(asgProfile.AllocatedPublicIP))
			asgBody.SetAttributeValue("cloud",
				cty.NumberIntVal(int64(asgProfile.Cloud)))
			asgBody.SetAttributeValue("keypair_type",
				cty.NumberIntVal(int64(asgProfile.KeyPairType)))
			if len(asgProfile.Base64UserData) > 0 {
				asgBody.SetAttributeValue("base64_user_data",
					cty.StringVal(asgProfile.Base64UserData))
			}

			if asgProfile.CustomDataTags != nil {
				for _, duploObject := range *asgProfile.CustomDataTags {
					minionTagsBlock := asgBody.AppendNewBlock("minion_tags",
						nil)
					minionTagsBody := minionTagsBlock.Body()
					minionTagsBody.SetAttributeValue("key",
						cty.StringVal(duploObject.Key))
					minionTagsBody.SetAttributeValue("value",
						cty.StringVal(duploObject.Value))
					rootBody.AppendNewline()
				}
			}
			//TODO - Duplo provider doesn't handle this yet.
			// if asgProfile.MetaData != nil {
			// 	for _, duploObject := range *asgProfile.MetaData {
			// 		mdBlock := asgBody.AppendNewBlock("metadata",
			// 			nil)
			// 		mdBody := mdBlock.Body()
			// 		mdBody.SetAttributeValue("key",
			// 			cty.StringVal(duploObject.Key))
			// 		mdBody.SetAttributeValue("value",
			// 			cty.StringVal(duploObject.Value))
			// 		rootBody.AppendNewline()
			// 	}
			// }
			if asgProfile.Volumes != nil && len(*asgProfile.Volumes) > 0 {
				for _, duploObject := range *asgProfile.Volumes {
					volumeBlock := asgBody.AppendNewBlock("volume",
						nil)
					volumeBody := volumeBlock.Body()
					volumeBody.SetAttributeValue("iops",
						cty.NumberIntVal(int64(duploObject.Iops)))
					volumeBody.SetAttributeValue("name",
						cty.StringVal(duploObject.Name))
					volumeBody.SetAttributeValue("size",
						cty.NumberIntVal(int64(duploObject.Size)))
					volumeBody.SetAttributeValue("volume_id",
						cty.StringVal(duploObject.VolumeID))
					volumeBody.SetAttributeValue("volume_type",
						cty.StringVal(duploObject.VolumeType))
					rootBody.AppendNewline()
				}
			}

			asgBody.SetAttributeValue("use_spot_instances", cty.BoolVal(asgProfile.UseSpotInstances))
			if asgProfile.MaxSpotPrice != "" {
				asgBody.SetAttributeValue("max_spot_price", cty.StringVal(asgProfile.MaxSpotPrice))
			}
			// TODO - Handle tags, network_interface
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo ASG : %s", asgProfile.FriendlyName)

			outVars := generateAsgOutputVars(asgProfile, varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_asg_profile." + resourceName,
					ResourceId:      config.TenantId + "/" + asgProfile.FriendlyName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== ASG TF generation done. =====>")
	}
	return &tfContext, nil
}

func generateAsgVars(duplo duplosdk.DuploAsgProfile, prefix string) []common.VarConfig {
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

	instanceCountVar := common.VarConfig{
		Name:       prefix + "instance_count",
		DefaultVal: strconv.Itoa(duplo.DesiredCapacity),
		TypeVal:    "number",
	}
	varConfigs["instance_count"] = instanceCountVar

	minCountVar := common.VarConfig{
		Name:       prefix + "min_instance_count",
		DefaultVal: strconv.Itoa(duplo.MinSize),
		TypeVal:    "number",
	}
	varConfigs["min_instance_count"] = minCountVar

	maxCountVar := common.VarConfig{
		Name:       prefix + "max_instance_count",
		DefaultVal: strconv.Itoa(duplo.MaxSize),
		TypeVal:    "number",
	}
	varConfigs["max_instance_count"] = maxCountVar

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}

func generateAsgOutputVars(duplo duplosdk.DuploAsgProfile, prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	fullNameVar := common.OutputVarConfig{
		Name:          prefix + "fullname",
		ActualVal:     "duplocloud_asg_profile." + resourceName + ".fullname",
		DescVal:       "The full name of the ASG.",
		RootTraversal: true,
	}
	outVarConfigs["fullname"] = fullNameVar

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
