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

const BSP_VAR_PREFIX = "batch_sp_"

type BatchSP struct {
}

func (bsp *BatchSP) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.AwsBatchSchedulingPolicyList(config.TenantId)
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
		log.Println("[TRACE] <====== AWS Batch Scheduling policy TF generation started. =====>")
		for _, sp := range *list {
			shortName, _ := duplosdk.UnprefixName(prefix, sp.Name)
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo AWS Batch Scheduling policy : %s", shortName)

			varFullPrefix := BCE_VAR_PREFIX + resourceName + "_"

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "batch-sp-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_batch_scheduling_policy resource
			bspBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_batch_scheduling_policy",
					resourceName})
			bspBody := bspBlock.Body()
			bspBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			bspBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			if sp.FairsharePolicy != nil {
				fspBlock := bspBody.AppendNewBlock("fair_share_policy",
					nil)
				fspBody := fspBlock.Body()
				fsp := sp.FairsharePolicy

				if fsp.ComputeReservation > 0 {
					fspBody.SetAttributeValue("compute_reservation",
						cty.NumberIntVal(int64(fsp.ComputeReservation)))
				}
				if fsp.ShareDecaySeconds > 0 {
					fspBody.SetAttributeValue("share_decay_seconds",
						cty.NumberIntVal(int64(fsp.ShareDecaySeconds)))
				}

				if fsp.ShareDistribution != nil && len(*fsp.ShareDistribution) > 0 {
					for _, sd := range *fsp.ShareDistribution {
						sdBlock := fspBody.AppendNewBlock("share_distribution",
							nil)
						sdBody := sdBlock.Body()
						if len(sd.ShareIdentifier) > 0 {
							sdBody.SetAttributeValue("share_identifier", cty.StringVal(sd.ShareIdentifier))
						}
						if sd.WeightFactor > 0 {
							sdBody.SetAttributeValue("weight_factor", cty.NumberFloatVal(float64(sd.WeightFactor)))
						}
					}
				}
			}
			if len(sp.Tags) > 0 {
				newMap := make(map[string]cty.Value)
				for key, element := range sp.Tags {
					if common.Contains(common.GetDuploManagedAwsTags(), key) {
						continue
					}

					newMap[key] = cty.StringVal(element)
				}
				if len(newMap) > 0 {
					bspBody.SetAttributeValue("tags", cty.ObjectVal(newMap))
				}
			}
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo AWS Batch Scheduling policy : %s", shortName)

			outVars := generateBatchSPOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_batch_scheduling_policy." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== AWS Batch Scheduling policy TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateBatchSPOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_batch_scheduling_policy." + resourceName + ".arn",
		DescVal:       "The Amazon Resource Name of the scheduling policy.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var1

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
