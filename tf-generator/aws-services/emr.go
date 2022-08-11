package awsservices

import (
	"encoding/json"
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

const EMR_VAR_PREFIX = "emr_"

type EMR struct {
}

func (emr *EMR) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.DuploEmrClusterGetList(config.TenantId)
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== EMR TF generation started. =====>")
		for _, emr := range *list {
			shortName, _ := extractEMRShortName(client, config.TenantId, emr.Name)
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo EMR Instance : %s", shortName)

			emrInfo, clientErr := client.DuploEmrClusterGet(config.TenantId, shortName)
			if clientErr != nil {
				fmt.Println(clientErr)
				return nil, clientErr
			}
			varFullPrefix := EMR_VAR_PREFIX + resourceName + "_"
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "emr-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_ecache_instance resource
			emrBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_emr_cluster",
					resourceName})
			emrBody := emrBlock.Body()
			emrBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			emrBody.SetAttributeValue("name",
				cty.StringVal(shortName))
			emrBody.SetAttributeValue("release_label", cty.StringVal(emrInfo.ReleaseLabel))

			if len(emrInfo.CustomAmiId) > 0 {
				emrBody.SetAttributeValue("custom_ami_id", cty.StringVal(emrInfo.CustomAmiId))
			}
			if len(emrInfo.LogUri) > 0 {
				emrBody.SetAttributeValue("log_uri", cty.StringVal(emrInfo.LogUri))
			}
			emrBody.SetAttributeValue("termination_protection", cty.BoolVal(emrInfo.TerminationProtection))
			emrBody.SetAttributeValue("keep_job_flow_alive_when_no_steps", cty.BoolVal(emrInfo.KeepJobFlowAliveWhenNoSteps))
			emrBody.SetAttributeValue("visible_to_all_users", cty.BoolVal(emrInfo.VisibleToAllUsers))

			if emrInfo.EbsRootVolumeSize > 0 {
				emrBody.SetAttributeValue("ebs_root_volume_size", cty.NumberIntVal(int64(emrInfo.EbsRootVolumeSize)))
			}
			if emrInfo.StepConcurrencyLevel > 0 {
				emrBody.SetAttributeValue("step_concurrency_level", cty.NumberIntVal(int64(emrInfo.StepConcurrencyLevel)))
			}
			if len(emrInfo.MasterInstanceType) > 0 {
				emrBody.SetAttributeValue("master_instance_type", cty.StringVal(emrInfo.MasterInstanceType))
			}
			if len(emrInfo.SlaveInstanceType) > 0 {
				emrBody.SetAttributeValue("slave_instance_type", cty.StringVal(emrInfo.SlaveInstanceType))
			}
			if emrInfo.InstanceCount > 0 {
				emrBody.SetAttributeValue("instance_count", cty.NumberIntVal(int64(emrInfo.InstanceCount)))
			}
			if len(emrInfo.ScaleDownBehavior) > 0 {
				emrBody.SetAttributeValue("scale_down_behavior", cty.StringVal(emrInfo.ScaleDownBehavior))
			}
			if len(emrInfo.Applications) > 0 {
				var appsMap interface{}
				err := json.Unmarshal([]byte(emrInfo.Applications), &appsMap)
				if err != nil {
					panic(err)
				}
				appsStr, err := duplosdk.JSONMarshal(appsMap)
				if err != nil {
					panic(err)
				}
				emrBody.SetAttributeTraversal("applications", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + appsStr + ")",
					},
				})
			}
			if len(emrInfo.BootstrapActions) > 0 {
				var bootstrapActionsMap interface{}
				err := json.Unmarshal([]byte(emrInfo.BootstrapActions), &bootstrapActionsMap)
				if err != nil {
					panic(err)
				}
				bootstrapActionsMapStr, err := duplosdk.JSONMarshal(bootstrapActionsMap)
				if err != nil {
					panic(err)
				}
				emrBody.SetAttributeTraversal("bootstrap_actions", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + bootstrapActionsMapStr + ")",
					},
				})
			}
			if len(emrInfo.Configurations) > 0 {
				var configurationsMap interface{}
				err := json.Unmarshal([]byte(emrInfo.Configurations), &configurationsMap)
				if err != nil {
					panic(err)
				}
				configurationsMapStr, err := duplosdk.JSONMarshal(configurationsMap)
				if err != nil {
					panic(err)
				}
				emrBody.SetAttributeTraversal("configurations", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + configurationsMapStr + ")",
					},
				})
			}
			if len(emrInfo.Steps) > 0 {
				var stepsMap interface{}
				err := json.Unmarshal([]byte(emrInfo.Steps), &stepsMap)
				if err != nil {
					panic(err)
				}
				stepsMapStr, err := duplosdk.JSONMarshal(stepsMap)
				if err != nil {
					panic(err)
				}
				emrBody.SetAttributeTraversal("steps", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + stepsMapStr + ")",
					},
				})
			}
			if len(emrInfo.AdditionalInfo) > 0 {
				var additionalInfoMap interface{}
				err := json.Unmarshal([]byte(emrInfo.AdditionalInfo), &additionalInfoMap)
				if err != nil {
					panic(err)
				}
				additionalInfoMapStr, err := duplosdk.JSONMarshal(additionalInfoMap)
				if err != nil {
					panic(err)
				}
				emrBody.SetAttributeTraversal("additional_info", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + additionalInfoMapStr + ")",
					},
				})
			}
			if len(emrInfo.ManagedScalingPolicy) > 0 {
				var managedScalingPolicyMap interface{}
				err := json.Unmarshal([]byte(emrInfo.ManagedScalingPolicy), &managedScalingPolicyMap)
				if err != nil {
					panic(err)
				}
				managedScalingPolicyMapStr, err := duplosdk.JSONMarshal(managedScalingPolicyMap)
				if err != nil {
					panic(err)
				}
				emrBody.SetAttributeTraversal("managed_scaling_policy", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + managedScalingPolicyMapStr + ")",
					},
				})
			}
			if len(emrInfo.InstanceFleets) > 0 {
				var instanceFleetsMap interface{}
				err := json.Unmarshal([]byte(emrInfo.InstanceFleets), &instanceFleetsMap)
				if err != nil {
					panic(err)
				}
				instanceFleetsMapStr, err := duplosdk.JSONMarshal(instanceFleetsMap)
				if err != nil {
					panic(err)
				}
				emrBody.SetAttributeTraversal("instance_fleets", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + instanceFleetsMapStr + ")",
					},
				})
			}
			if len(emrInfo.InstanceGroups) > 0 {
				var instanceGroupsMap interface{}
				err := json.Unmarshal([]byte(emrInfo.InstanceGroups), &instanceGroupsMap)
				if err != nil {
					panic(err)
				}
				instanceGroupsMapStr, err := duplosdk.JSONMarshal(instanceGroupsMap)
				if err != nil {
					panic(err)
				}
				emrBody.SetAttributeTraversal("instance_groups", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + instanceGroupsMapStr + ")",
					},
				})
			}

			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			log.Printf("[TRACE] Terraform config is generated for duplo EMR : %s", shortName)

			outVars := generateEMROutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_emr_cluster." + resourceName,
					ResourceId:      config.TenantId + "/" + emr.JobFlowId,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== EMR TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateEMROutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "fullname",
		ActualVal:     "duplocloud_emr_cluster." + resourceName + ".fullname",
		DescVal:       "The full name of the EMR cluster.",
		RootTraversal: true,
	}
	outVarConfigs["fullname"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_emr_cluster." + resourceName + ".arn",
		DescVal:       "The ARN of the EMR cluster.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "job_flow_id",
		ActualVal:     "duplocloud_emr_cluster." + resourceName + ".job_flow_id",
		DescVal:       "job flow id.",
		RootTraversal: true,
	}
	outVarConfigs["job_flow_id"] = var3

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}

func extractEMRShortName(client *duplosdk.Client, tenantID string, fullName string) (string, error) {
	prefix, err := client.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return "", err
	}
	name, _ := duplosdk.UnprefixName(prefix, fullName)
	return name, nil
}
