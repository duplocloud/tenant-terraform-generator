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

const BCE_VAR_PREFIX = "batch_ce_"

type BatchCE struct {
}

func (bce *BatchCE) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.AwsBatchComputeEnvironmentList(config.TenantId)
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
		log.Println("[TRACE] <====== AWS Batch Compute Environment TF generation started. =====>")
		for _, ce := range *list {
			shortName, _ := duplosdk.UnprefixName(prefix, ce.ComputeEnvironmentName)
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo AWS Batch Compute Environment : %s", shortName)

			varFullPrefix := BCE_VAR_PREFIX + resourceName + "_"

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "batch-ce-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_batch_compute_environment resource
			bceBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_batch_compute_environment",
					resourceName})
			bceBody := bceBlock.Body()
			bceBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			bceBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			if ce.Type != nil && len(ce.Type.Value) > 0 {
				bceBody.SetAttributeValue("type",
					cty.StringVal(ce.Type.Value))
			} else {
				bceBody.SetAttributeValue("type",
					cty.StringVal("MANAGED"))
			}

			if ce.State != nil && len(ce.State.Value) > 0 {
				bceBody.SetAttributeValue("state",
					cty.StringVal(ce.State.Value))
			} else {
				bceBody.SetAttributeValue("state",
					cty.StringVal("ENABLED"))
			}

			if ce.ComputeResources != nil {
				crBlock := bceBody.AppendNewBlock("compute_resources",
					nil)
				crBody := crBlock.Body()
				cr := ce.ComputeResources

				if cr.AllocationStrategy != nil && len(cr.AllocationStrategy.Value) > 0 {
					crBody.SetAttributeValue("allocation_strategy",
						cty.StringVal(cr.AllocationStrategy.Value))
				}
				if cr.Type != nil && len(cr.Type.Value) > 0 {
					crBody.SetAttributeValue("type",
						cty.StringVal(cr.Type.Value))
				}
				if cr.BidPercentage > 0 {
					crBody.SetAttributeValue("bid_percentage",
						cty.NumberIntVal(int64(cr.BidPercentage)))
				}
				if cr.DesiredvCpus > 0 {
					crBody.SetAttributeValue("desired_vcpus",
						cty.NumberIntVal(int64(cr.DesiredvCpus)))
				}
				if cr.MaxvCpus > 0 {
					crBody.SetAttributeValue("max_vcpus",
						cty.NumberIntVal(int64(cr.MaxvCpus)))
				}
				if cr.MinvCpus > 0 {
					crBody.SetAttributeValue("min_vcpus",
						cty.NumberIntVal(int64(cr.MinvCpus)))
				}

				if len(cr.InstanceTypes) > 0 {
					var vals []cty.Value
					for _, s := range cr.InstanceTypes {
						vals = append(vals, cty.StringVal(s))
					}
					crBody.SetAttributeValue("instance_type",
						cty.ListVal(vals))
				}

				if cr.Ec2Configuration != nil && len(*cr.Ec2Configuration) > 0 {
					imageType := "ECS_AL2"
					for _, ec2c := range *cr.Ec2Configuration {
						if ec2c.ImageType == imageType && len(*cr.Ec2Configuration) > 1 {
							continue
						}
						ec2cBlock := crBody.AppendNewBlock("ec2_configuration",
							nil)
						ec2cBody := ec2cBlock.Body()
						if len(ec2c.ImageType) > 0 {
							ec2cBody.SetAttributeValue("image_type", cty.StringVal(ec2c.ImageType))
						}
						if len(ec2c.ImageIdOverride) > 0 {
							ec2cBody.SetAttributeValue("image_id_override", cty.StringVal(ec2c.ImageIdOverride))
						}
					}
				}

				if cr.LaunchTemplate != nil {
					lt := cr.LaunchTemplate
					ltBlock := crBody.AppendNewBlock("launch_template",
						nil)
					ltBody := ltBlock.Body()
					if len(lt.LaunchTemplateId) > 0 {
						ltBody.SetAttributeValue("launch_template_id", cty.StringVal(lt.LaunchTemplateId))
					}
					if len(lt.LaunchTemplateName) > 0 {
						ltBody.SetAttributeValue("launch_template_name", cty.StringVal(lt.LaunchTemplateName))
					}
					if len(lt.Version) > 0 {
						ltBody.SetAttributeValue("version", cty.StringVal(lt.Version))
					}
				}
			}

			if len(ce.Tags) > 0 {
				newMap := make(map[string]cty.Value)
				for key, element := range ce.Tags {
					if common.Contains(common.GetDuploManagedAwsTags(), key) {
						continue
					}

					newMap[key] = cty.StringVal(element)
				}
				if len(newMap) > 0 {
					bceBody.SetAttributeValue("tags", cty.ObjectVal(newMap))
				}
			}

			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo AWS Batch Compute Environment : %s", shortName)

			outVars := generateBatchCEOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_batch_compute_environment." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== AWS Batch Compute Environment TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateBatchCEOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_batch_compute_environment." + resourceName + ".arn",
		DescVal:       "The Amazon Resource Name (ARN) of the compute environment.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var1

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
