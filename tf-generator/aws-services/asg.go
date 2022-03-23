package awsservices

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

type ASG struct {
}

func (asg *ASG) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	log.Println("[TRACE] <====== ASG TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.AsgProfileGetList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	if list != nil {
		for _, asgProfile := range *list {
			shortName := asgProfile.FriendlyName[len("duploservices-"+config.TenantName+"-"):len(asgProfile.FriendlyName)]
			log.Printf("[TRACE] Generating terraform config for duplo ASG : %s", asgProfile.FriendlyName)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

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
					shortName})
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
			asgBody.SetAttributeValue("instance_count",
				cty.NumberIntVal(int64(asgProfile.DesiredCapacity)))
			asgBody.SetAttributeValue("min_instance_count",
				cty.NumberIntVal(int64(asgProfile.MinSize)))
			asgBody.SetAttributeValue("max_instance_count",
				cty.NumberIntVal(int64(asgProfile.MaxSize)))
			asgBody.SetAttributeValue("image_id",
				cty.StringVal(asgProfile.ImageID))
			asgBody.SetAttributeValue("capacity",
				cty.StringVal(asgProfile.Capacity))
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
			if len(asgProfile.Base64UserData) > 0 {
				asgBody.SetAttributeValue("base64_user_data",
					cty.StringVal(asgProfile.Base64UserData))
			}

			if asgProfile.MinionTags != nil {
				for _, duploObject := range *asgProfile.MinionTags {
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
			// TODO - Duplo provider doesn't handle this yet.
			// if asgProfile.Volumes != nil {
			// 	for _, duploObject := range *asgProfile.Volumes {
			// 		volumeBlock := asgBody.AppendNewBlock("volume",
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
			tfFile.Write(hclFile.Bytes())
			log.Printf("[TRACE] Terraform config is generated for duplo ASG : %s", asgProfile.FriendlyName)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs := []common.ImportConfig{}
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_asg_profile." + shortName,
					ResourceId:      config.TenantId + "/" + asgProfile.FriendlyName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
	}
	log.Println("[TRACE] <====== ASG TF generation done. =====>")
	return &tfContext, nil
}
