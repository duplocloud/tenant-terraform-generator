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

const TTDB_VAR_PREFIX = "timestream_db"
const TTDB_TABLE_VAR_PREFIX = "timestream_tbl"

type TimestreamDB struct {
}

func (tdb *TimestreamDB) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.DuploTimestreamDBGetList(config.TenantId)

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, nil
	}
	prefix, clientErr := client.GetDuploServicesPrefix(config.TenantId)
	if clientErr != nil {
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== AWS Batch Job Queue TF generation started. =====>")
		for _, tsdb := range *list {
			shortName, _ := duplosdk.UnprefixName(prefix, tsdb.DatabaseName)
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo AWS Timestream DB: %s", shortName)

			varFullPrefix := TTDB_VAR_PREFIX + resourceName + "_"

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "timestream-db-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_timestreamwrite_database resource
			tsdbBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_timestreamwrite_database",
					resourceName})
			tsdbBody := tsdbBlock.Body()
			tsdbBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			tsdbBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			if len(tsdb.KmsKeyId) > 0 {
				tsdbBody.SetAttributeValue("kms_key_id",
					cty.StringVal(tsdb.KmsKeyId))
			}
			if tsdb.Tags != nil && len(*tsdb.Tags) > 0 {
				for _, duploObject := range *tsdb.Tags {
					if len(duploObject.Value) > 0 {
						if common.Contains(common.GetDuploManagedAwsTags(), duploObject.Key) {
							continue
						}
						tagsBlock := tsdbBody.AppendNewBlock("tags",
							nil)
						tagsBody := tagsBlock.Body()
						tagsBody.SetAttributeValue("key",
							cty.StringVal(duploObject.Key))
						tagsBody.SetAttributeValue("value",
							cty.StringVal(duploObject.Value))
						rootBody.AppendNewline()
					}
				}
			}

			// Add duplocloud_aws_timestreamwrite_table

			tblList, clientErr := client.DuploTimestreamDBTableGetList(config.TenantId, tsdb.DatabaseName)
			if clientErr != nil {
				fmt.Println(clientErr)
				return nil, nil
			}
			if tblList != nil && len(*tblList) > 0 {
				for _, tstbl := range *tblList {
					tblResourceName := common.GetResourceName(tstbl.TableName)
					tstblBlock := rootBody.AppendNewBlock("resource",
						[]string{"duplocloud_aws_timestreamwrite_table",
							tblResourceName})
					tstblBody := tstblBlock.Body()
					tstblBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "local",
						},
						hcl.TraverseAttr{
							Name: "tenant_id",
						},
					})

					tstblBody.SetAttributeValue("name",
						cty.StringVal(tstbl.TableName))

					tstblBody.SetAttributeTraversal("database_name", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "duplocloud_aws_timestreamwrite_database." + resourceName,
						},
						hcl.TraverseAttr{
							Name: "fullname",
						},
					})

					if tstbl.RetentionProperties != nil {
						rpBlock := tstblBody.AppendNewBlock("retention_properties",
							nil)
						rpBody := rpBlock.Body()
						rpBody.SetAttributeValue("magnetic_store_retention_period_in_days",
							cty.NumberIntVal(int64(tstbl.RetentionProperties.MagneticStoreRetentionPeriodInDays)))
						rpBody.SetAttributeValue("memory_store_retention_period_in_hours",
							cty.NumberIntVal(int64(tstbl.RetentionProperties.MemoryStoreRetentionPeriodInHours)))

					}

					if tstbl.MagneticStoreWriteProperties != nil && tstbl.MagneticStoreWriteProperties.EnableMagneticStoreWrites {
						mswBlock := tstblBody.AppendNewBlock("magnetic_store_write_properties",
							nil)
						mswBody := mswBlock.Body()
						if tstbl.MagneticStoreWriteProperties.EnableMagneticStoreWrites {
							mswBody.SetAttributeValue("enable_magnetic_store_writes",
								cty.BoolVal(tstbl.MagneticStoreWriteProperties.EnableMagneticStoreWrites))
						}
						if tstbl.MagneticStoreWriteProperties.MagneticStoreRejectedDataLocation != nil && tstbl.MagneticStoreWriteProperties.MagneticStoreRejectedDataLocation.S3Configuration != nil {
							msrBlock := mswBody.AppendNewBlock("magnetic_store_rejected_data_location",
								nil)
							msrBody := msrBlock.Body()
							s3cBlock := msrBody.AppendNewBlock("s3_configuration",
								nil)
							s3cBody := s3cBlock.Body()
							s3c := tstbl.MagneticStoreWriteProperties.MagneticStoreRejectedDataLocation.S3Configuration
							if len(s3c.BucketName) > 0 {
								s3cBody.SetAttributeValue("bucket_name",
									cty.StringVal(s3c.BucketName))
							}
							if len(s3c.KmsKeyId) > 0 {
								s3cBody.SetAttributeValue("kms_key_id",
									cty.StringVal(s3c.KmsKeyId))
							}
							if s3c.EncryptionOption != nil && len(s3c.EncryptionOption.Value) > 0 {
								s3cBody.SetAttributeValue("encryption_option",
									cty.StringVal(s3c.EncryptionOption.Value))
							}
							if len(s3c.ObjectKeyPrefix) > 0 {
								s3cBody.SetAttributeValue("object_key_prefix",
									cty.StringVal(s3c.ObjectKeyPrefix))
							}
						}
					}

					if tstbl.Tags != nil && len(*tstbl.Tags) > 0 {
						for _, duploObject := range *tstbl.Tags {
							if len(duploObject.Value) > 0 {
								if common.Contains(common.GetDuploManagedAwsTags(), duploObject.Key) {
									continue
								}
								tagsBlock := tstblBody.AppendNewBlock("tags",
									nil)
								tagsBody := tagsBlock.Body()
								tagsBody.SetAttributeValue("key",
									cty.StringVal(duploObject.Key))
								tagsBody.SetAttributeValue("value",
									cty.StringVal(duploObject.Value))
								rootBody.AppendNewline()
							}
						}
					}

					if config.GenerateTfState {
						importConfigs = append(importConfigs, common.ImportConfig{
							ResourceAddress: "duplocloud_aws_timestreamwrite_table." + tblResourceName,
							ResourceId:      config.TenantId + "/" + tstbl.TableName,
							WorkingDir:      workingDir,
						})
						tfContext.ImportConfigs = importConfigs
					}
				}

			}

			log.Printf("[TRACE] Generating tf file for resource : %s", shortName)
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo AWS Timestream Database: %s", shortName)

			outVars := generateTimestreamDBOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_timestreamwrite_database." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== AWS Timestream Database TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateTimestreamDBOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_timestreamwrite_database." + resourceName + ".arn",
		DescVal:       "The ARN that uniquely identifies this database.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var1

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
