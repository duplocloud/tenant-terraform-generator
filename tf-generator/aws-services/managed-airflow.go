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

const MWAA_VAR_PREFIX = "mwaa_"

type MWAA struct {
}

func (mwaa *MWAA) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.MwaaAirflowList(config.TenantId)
	//Get tenant from duplo

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
		log.Println("[TRACE] <====== AWS Apache Airflow TF generation started. =====>")
		kms, kmsClientErr := client.TenantGetTenantKmsKey(config.TenantId)
		for _, mwaa := range *list {
			shortName, _ := duplosdk.UnprefixName(prefix, mwaa.Name)
			log.Printf("[TRACE] Generating terraform config for duplo AWS Apache Airflow : %s", mwaa.Name)

			varFullPrefix := MWAA_VAR_PREFIX + strings.ReplaceAll(shortName, "-", "_") + "_"
			mwaaDetails, clientErr := client.MwaaAirflowDetailsGet(config.TenantId, mwaa.Name)
			if clientErr != nil {
				fmt.Println(clientErr)
				return nil, clientErr
			}

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "mwaa-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_mwaa_environment resource
			mwaalock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_mwaa_environment",
					shortName})
			mwaaBody := mwaalock.Body()
			mwaaBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			mwaaBody.SetAttributeValue("name", cty.StringVal(shortName))

			if len(mwaaDetails.SourceBucketArn) > 0 {
				resourceName := duplosdk.UnwrapResoureNameFromAwsArn(mwaaDetails.SourceBucketArn)
				s3, clientErr := client.TenantGetS3BucketSettings(config.TenantId, resourceName)
				if s3 != nil && clientErr == nil && s3.Arn == mwaaDetails.SourceBucketArn {
					s3ShortName := s3.Name[len("duploservices-"+config.TenantName+"-"):len(s3.Name)]
					parts := strings.Split(s3ShortName, "-")
					if len(parts) > 0 {
						parts = parts[:len(parts)-1]
					}
					s3ShortName = strings.Join(parts, "-")
					mwaaBody.SetAttributeTraversal("source_bucket_arn", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "duplocloud_s3_bucket." + s3ShortName,
						},
						hcl.TraverseAttr{
							Name: "arn",
						},
					})
				} else {
					mwaaBody.SetAttributeValue("source_bucket_arn", cty.StringVal(mwaaDetails.SourceBucketArn))
				}

			}

			if len(mwaaDetails.DagS3Path) > 0 {
				mwaaBody.SetAttributeValue("dag_s3_path", cty.StringVal(mwaaDetails.DagS3Path))
			}

			if len(mwaaDetails.KmsKey) > 0 {
				if kms != nil && kmsClientErr == nil && (mwaaDetails.KmsKey == kms.KeyArn || mwaaDetails.KmsKey == kms.KeyID) {
					mwaaBody.SetAttributeTraversal("kms_key", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "data.duplocloud_tenant_aws_kms_key.tenant_kms",
						},
						hcl.TraverseAttr{
							Name: "key_arn",
						},
					})
				} else {
					mwaaBody.SetAttributeValue("kms_key", cty.StringVal(mwaaDetails.KmsKey))
				}
			}

			mwaaBody.SetAttributeValue("schedulers",
				cty.NumberIntVal(int64(mwaaDetails.Schedulers)))

			mwaaBody.SetAttributeValue("max_workers",
				cty.NumberIntVal(int64(mwaaDetails.MinWorkers)))

			mwaaBody.SetAttributeValue("max_workers",
				cty.NumberIntVal(int64(mwaaDetails.MaxWorkers)))
			if len(mwaaDetails.AirflowVersion) > 0 {
				mwaaBody.SetAttributeValue("airflow_version", cty.StringVal(mwaaDetails.AirflowVersion))
			}

			if len(mwaaDetails.WeeklyMaintenanceWindowStart) > 0 {
				mwaaBody.SetAttributeValue("weekly_maintenance_window_start", cty.StringVal(mwaaDetails.WeeklyMaintenanceWindowStart))
			}

			if len(mwaaDetails.PluginsS3Path) > 0 {
				mwaaBody.SetAttributeValue("plugins_s3_path", cty.StringVal(mwaaDetails.PluginsS3Path))
			}

			if len(mwaaDetails.PluginsS3ObjectVersion) > 0 {
				mwaaBody.SetAttributeValue("plugins_s3_object_version", cty.StringVal(mwaaDetails.PluginsS3ObjectVersion))
			}

			if len(mwaaDetails.RequirementsS3Path) > 0 {
				mwaaBody.SetAttributeValue("requirements_s3_path", cty.StringVal(mwaaDetails.RequirementsS3Path))
			}

			if len(mwaaDetails.RequirementsS3ObjectVersion) > 0 {
				mwaaBody.SetAttributeValue("requirements_s3_object_version", cty.StringVal(mwaaDetails.RequirementsS3ObjectVersion))
			}

			if mwaaDetails.WebserverAccessMode != nil && len(mwaaDetails.WebserverAccessMode.Value) > 0 {
				mwaaBody.SetAttributeValue("webserver_access_mode", cty.StringVal(mwaaDetails.WebserverAccessMode.Value))
			}

			// if len(mwaaDetails.ExecutionRoleArn) > 0 {
			// 	mwaaBody.SetAttributeValue("execution_role_arn", cty.StringVal(mwaaDetails.ExecutionRoleArn))
			// }

			if len(mwaaDetails.EnvironmentClass) > 0 {
				mwaaBody.SetAttributeValue("environment_class", cty.StringVal(mwaaDetails.EnvironmentClass))
			}
			if len(mwaaDetails.AirflowConfigurationOptions) > 0 {
				newMap := make(map[string]cty.Value)
				for key, element := range mwaaDetails.AirflowConfigurationOptions {
					newMap[key] = cty.StringVal(element)
				}
				mwaaBody.SetAttributeValue("airflow_configuration_options", cty.ObjectVal(newMap))
			}

			if mwaaDetails.LoggingConfiguration != nil {
				logConfigBlock := mwaaBody.AppendNewBlock("logging_configuration",
					nil)
				logConfigBody := logConfigBlock.Body()
				if mwaaDetails.LoggingConfiguration.DagProcessingLogs != nil {
					dagConfigBlock := logConfigBody.AppendNewBlock("dag_processing_logs",
						nil)
					dagConfigBody := dagConfigBlock.Body()
					dagConfigBody.SetAttributeValue("enabled",
						cty.BoolVal(mwaaDetails.LoggingConfiguration.DagProcessingLogs.Enabled))
					dagConfigBody.SetAttributeValue("log_level",
						cty.StringVal(mwaaDetails.LoggingConfiguration.DagProcessingLogs.LogLevel.Value))
					logConfigBody.AppendNewline()
				}
				if mwaaDetails.LoggingConfiguration.SchedulerLogs != nil {
					schConfigBlock := logConfigBody.AppendNewBlock("scheduler_logs",
						nil)
					schConfigBody := schConfigBlock.Body()
					schConfigBody.SetAttributeValue("enabled",
						cty.BoolVal(mwaaDetails.LoggingConfiguration.SchedulerLogs.Enabled))
					schConfigBody.SetAttributeValue("log_level",
						cty.StringVal(mwaaDetails.LoggingConfiguration.SchedulerLogs.LogLevel.Value))
					logConfigBody.AppendNewline()
				}
				if mwaaDetails.LoggingConfiguration.TaskLogs != nil {
					taskConfigBlock := logConfigBody.AppendNewBlock("task_logs",
						nil)
					taskConfigBody := taskConfigBlock.Body()
					taskConfigBody.SetAttributeValue("enabled",
						cty.BoolVal(mwaaDetails.LoggingConfiguration.TaskLogs.Enabled))
					taskConfigBody.SetAttributeValue("log_level",
						cty.StringVal(mwaaDetails.LoggingConfiguration.TaskLogs.LogLevel.Value))
					logConfigBody.AppendNewline()
				}
				if mwaaDetails.LoggingConfiguration.WebserverLogs != nil {
					webConfigBlock := logConfigBody.AppendNewBlock("webserver_logs",
						nil)
					webConfigBody := webConfigBlock.Body()
					webConfigBody.SetAttributeValue("enabled",
						cty.BoolVal(mwaaDetails.LoggingConfiguration.WebserverLogs.Enabled))
					webConfigBody.SetAttributeValue("log_level",
						cty.StringVal(mwaaDetails.LoggingConfiguration.WebserverLogs.LogLevel.Value))
					logConfigBody.AppendNewline()
				}
				if mwaaDetails.LoggingConfiguration.WorkerLogs != nil {
					workerConfigBlock := logConfigBody.AppendNewBlock("worker_logs",
						nil)
					workerConfigBody := workerConfigBlock.Body()
					workerConfigBody.SetAttributeValue("enabled",
						cty.BoolVal(mwaaDetails.LoggingConfiguration.WorkerLogs.Enabled))
					workerConfigBody.SetAttributeValue("log_level",
						cty.StringVal(mwaaDetails.LoggingConfiguration.WorkerLogs.LogLevel.Value))
					logConfigBody.AppendNewline()
				}

			}

			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo AWS Apache Airflow : %s", mwaa.Name)

			outVars := generateMWAAOutputVars(varFullPrefix, shortName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_mwaa_environment." + shortName,
					ResourceId:      config.TenantId + "/" + mwaa.Name,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== AWS Apache Airflow TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateMWAAOutputVars(prefix, shortName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "webserver_url",
		ActualVal:     "duplocloud_aws_mwaa_environment." + shortName + ".webserver_url",
		DescVal:       "The webserver URL of the MWAA Environment.",
		RootTraversal: true,
	}
	outVarConfigs["webserver_url"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_mwaa_environment." + shortName + ".arn",
		DescVal:       "The ARN of the Managed Workflows Apache Airflow.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
