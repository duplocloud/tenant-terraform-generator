package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type Redis struct {
}

func (r *Redis) Generate(config *common.Config, client *duplosdk.Client) {
	log.Println("[TRACE] <====== Redis TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.EcacheInstanceList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return
	}

	if list != nil {
		for _, redis := range *list {
			shortName := redis.Identifier[len("duplo-"):len(redis.Identifier)]
			log.Printf("[TRACE] Generating terraform config for duplo Redis Instance : %s", redis.Identifier)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "redis-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_ecache_instance resource
			redisBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_ecache_instance",
					shortName})
			redisBody := redisBlock.Body()
			// redisBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
			// 	hcl.TraverseRoot{
			// 		Name: "duplocloud_tenant.tenant",
			// 	},
			// 	hcl.TraverseAttr{
			// 		Name: "tenant_id",
			// 	},
			// })
			redisBody.SetAttributeValue("tenant_id",
				cty.StringVal(config.TenantId))
			redisBody.SetAttributeValue("name",
				cty.StringVal(shortName))
			redisBody.SetAttributeValue("cache_type",
				cty.NumberIntVal(int64(0)))
			redisBody.SetAttributeValue("replicas",
				cty.NumberIntVal(int64(redis.Replicas)))

			redisBody.SetAttributeValue("size",
				cty.StringVal(redis.Size))

			redisBody.SetAttributeValue("encryption_at_rest",
				cty.BoolVal(redis.EncryptionAtRest))
			redisBody.SetAttributeValue("encryption_in_transit",
				cty.BoolVal(redis.EncryptionInTransit))
			if len(redis.AuthToken) > 0 {
				redisBody.SetAttributeValue("auth_token",
					cty.StringVal(redis.AuthToken))
			}
			if len(redis.KMSKeyID) > 0 {
				redisBody.SetAttributeValue("kms_key_id",
					cty.StringVal(redis.KMSKeyID))
			}

			//fmt.Printf("%s", hclFile.Bytes())
			tfFile.Write(hclFile.Bytes())
			log.Printf("[TRACE] Terraform config is generated for duplo redis instance : %s", redis.Identifier)

			// Import all created resources.
			if config.GenerateTfState {
				importer := &common.Importer{}
				importer.Import(config, &common.ImportConfig{
					ResourceAddress: "duplocloud_ecache_instance." + shortName,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/ECacheDBInstance/" + shortName,
					WorkingDir:      workingDir,
				})
			}
		}
	}
	log.Println("[TRACE] <====== redis TF generation done. =====>")
}
