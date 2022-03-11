package tfgenerator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type Hosts struct {
}

func (h *Hosts) Generate(config *Config, client *duplosdk.Client) {
	log.Println("[TRACE] <====== Hosts TF generation started. =====>")
	workingDir := filepath.Join("target", config.CustomerName, config.AwsServicesProject)
	list, clientErr := client.NativeHostGetList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return
	}

	if list != nil {
		for _, host := range *list {
			shortName := host.FriendlyName[len("duploservices-"+host.UserAccount+"-"):len(host.FriendlyName)]
			log.Printf("[TRACE] Generating terraform config for duplo host : %s", host.FriendlyName)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "host-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_host resource
			hostBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_host",
					shortName})
			hostBody := hostBlock.Body()
			// hostBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
			// 	hcl.TraverseRoot{
			// 		Name: "duplocloud_tenant.tenant",
			// 	},
			// 	hcl.TraverseAttr{
			// 		Name: "tenant_id",
			// 	},
			// })
			hostBody.SetAttributeValue("tenant_id",
				cty.StringVal(config.TenantId))
			hostBody.SetAttributeValue("friendly_name",
				cty.StringVal(shortName))
			hostBody.SetAttributeValue("image_id",
				cty.StringVal(host.ImageID))
			hostBody.SetAttributeValue("capacity",
				cty.StringVal(host.Capacity))
			hostBody.SetAttributeValue("agent_platform",
				cty.NumberIntVal(int64(host.AgentPlatform)))
			hostBody.SetAttributeValue("zone",
				cty.NumberIntVal(int64(host.Zone)))
			hostBody.SetAttributeValue("user_account",
				cty.StringVal(host.UserAccount))
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
			hostBody.SetAttributeValue("base64_user_data",
				cty.StringVal(host.Base64UserData))

			if host.MinionTags != nil {
				for _, duploObject := range *host.MinionTags {
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
			tfFile.Write(hclFile.Bytes())
			log.Printf("[TRACE] Terraform config is generated for duplo host : %s", host.FriendlyName)

			// Import all created resources.
			if config.GenerateTfState {
				importer := &Importer{}
				importer.Import(config, &ImportConfig{
					resourceAddress: "duplocloud_aws_host." + shortName,
					resourceId:      "v2/subscriptions/" + config.TenantId + "/NativeHostV2/" + host.InstanceID,
					workingDir:      workingDir,
				})
			}
		}
	}
	log.Println("[TRACE] <====== Hosts TF generation done. =====>")
}
