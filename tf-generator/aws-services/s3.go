package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type S3Bucket struct {
}

func (s3 *S3Bucket) Generate(config *common.Config, client *duplosdk.Client) {
	log.Println("[TRACE] <====== S3 bucket TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantListS3Buckets(config.TenantId)

	//duplo, clientErr := client.TenantGetS3BucketSettings(config.TenantId)
	// if duplo == nil {
	// 	d.SetId("") // object missing
	// 	return nil
	// }
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return
	}

	if list != nil {
		for _, s3 := range *list {
			shortName := s3.Name[len("duploservices-"+config.TenantName+"-"):len(s3.Name)]
			parts := strings.Split(shortName, "-")
			if len(parts) > 0 {
				parts = parts[:len(parts)-1]
			}
			shortName = strings.Join(parts, "-")
			log.Printf("[TRACE] Generating terraform config for duplo s3 bucket : %s", shortName)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "s3-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_s3_bucket resource
			s3Block := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_s3_bucket",
					shortName})
			s3Body := s3Block.Body()
			// s3Body.SetAttributeTraversal("tenant_id", hcl.Traversal{
			// 	hcl.TraverseRoot{
			// 		Name: "duplocloud_tenant.tenant",
			// 	},
			// 	hcl.TraverseAttr{
			// 		Name: "tenant_id",
			// 	},
			// })
			s3Body.SetAttributeValue("tenant_id",
				cty.StringVal(config.TenantId))
			s3Body.SetAttributeValue("name",
				cty.StringVal(shortName))

			s3Body.SetAttributeValue("allow_public_access",
				cty.BoolVal(s3.AllowPublicAccess))
			s3Body.SetAttributeValue("enable_access_logs",
				cty.BoolVal(s3.EnableAccessLogs))
			s3Body.SetAttributeValue("enable_versioning",
				cty.BoolVal(s3.EnableVersioning))
			var encryptionMethod string
			if len(s3.DefaultEncryption) > 0 {
				encryptionMethod = s3.DefaultEncryption
			} else {
				encryptionMethod = "Sse"
			}
			defaultEncrBlock := s3Body.AppendNewBlock("default_encryption",
				nil)
			defaultEncrBody := defaultEncrBlock.Body()
			defaultEncrBody.SetAttributeValue("method",
				cty.StringVal(encryptionMethod))
			//fmt.Printf("%s", hclFile.Bytes())
			tfFile.Write(hclFile.Bytes())
			log.Printf("[TRACE] Terraform config is generated for duplo s3 bucket : %s", shortName)

			// Import all created resources.
			if config.GenerateTfState {
				importer := &common.Importer{}
				importer.Import(config, &common.ImportConfig{
					ResourceAddress: "duplocloud_s3_bucket." + shortName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
			}
		}
	}
	log.Println("[TRACE] <====== S3 Bucket TF generation done. =====>")
}
