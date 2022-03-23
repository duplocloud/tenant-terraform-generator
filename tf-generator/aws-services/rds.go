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

type Rds struct {
}

func (r *Rds) Generate(config *common.Config, client *duplosdk.Client) {
	log.Println("[TRACE] <====== RDS TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.RdsInstanceList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return
	}

	if list != nil {
		for _, rds := range *list {
			shortName := rds.Identifier[len("duplo"):len(rds.Identifier)]
			log.Printf("[TRACE] Generating terraform config for duplo RDS Instance : %s", rds.Identifier)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "rds-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_rds_instance resource
			rdsBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_rds_instance",
					shortName})
			rdsBody := rdsBlock.Body()
			rdsBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// rdsBody.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			rdsBody.SetAttributeValue("name",
				cty.StringVal(shortName))
			rdsBody.SetAttributeValue("engine",
				cty.NumberIntVal(int64(rds.Engine)))
			rdsBody.SetAttributeValue("engine_version",
				cty.StringVal(rds.EngineVersion))
			rdsBody.SetAttributeValue("size",
				cty.StringVal(rds.SizeEx))

			if len(rds.SnapshotID) > 0 {
				rdsBody.SetAttributeValue("snapshot_id",
					cty.StringVal(rds.SnapshotID))
			} else {
				rdsBody.SetAttributeValue("master_username",
					cty.StringVal(rds.MasterUsername))

			}
			rdsBody.SetAttributeValue("master_password",
				cty.StringVal(rds.MasterPassword))
			// if len(rds.DBParameterGroupName) > 0 {
			// 	rdsBody.SetAttributeValue("parameter_group_name",
			// 		cty.StringVal(rds.DBParameterGroupName))
			// }
			// rdsBody.SetAttributeValue("store_details_in_secret_manager",
			// 	cty.BoolVal(rds.StoreDetailsInSecretManager))

			rdsBody.SetAttributeValue("encrypt_storage",
				cty.BoolVal(rds.EncryptStorage))
			rdsBody.SetAttributeValue("enable_logging",
				cty.BoolVal(rds.EnableLogging))
			rdsBody.SetAttributeValue("multi_az",
				cty.BoolVal(rds.MultiAZ))
			//fmt.Printf("%s", hclFile.Bytes())
			tfFile.Write(hclFile.Bytes())
			log.Printf("[TRACE] Terraform config is generated for duplo RDS instance : %s", rds.Identifier)

			// Import all created resources.
			if config.GenerateTfState {
				importer := &common.Importer{}
				importer.Import(config, &common.ImportConfig{
					ResourceAddress: "duplocloud_rds_instance." + shortName,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/RDSDBInstance/" + shortName,
					WorkingDir:      workingDir,
				})
			}
		}
	}
	log.Println("[TRACE] <====== RDS TF generation done. =====>")
}
