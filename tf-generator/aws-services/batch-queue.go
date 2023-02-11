package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const BJQ_VAR_PREFIX = "batch_q_"

type BatchQ struct {
}

func (bjq *BatchQ) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.AwsBatchJobQueueList(config.TenantId)

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
		spList, _ := client.AwsBatchSchedulingPolicyList(config.TenantId)
		ceList, _ := client.AwsBatchComputeEnvironmentList(config.TenantId)
		for _, q := range *list {
			shortName, _ := duplosdk.UnprefixName(prefix, q.JobQueueName)
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo AWS Batch Job Queue: %s", shortName)

			varFullPrefix := BJQ_VAR_PREFIX + resourceName + "_"

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "batch-q-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_batch_job_queue resource
			jqBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_batch_job_queue",
					resourceName})
			jqBody := jqBlock.Body()
			jqBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			jqBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			if q.State != nil && len(q.State.Value) > 0 {
				jqBody.SetAttributeValue("state",
					cty.StringVal(q.State.Value))
			} else {
				jqBody.SetAttributeValue("state",
					cty.StringVal("ENABLED"))
			}

			if q.Priority > 0 {
				jqBody.SetAttributeValue("priority",
					cty.NumberIntVal(int64(q.Priority)))
			}

			if len(q.SchedulingPolicyArn) > 0 {
				for _, sp := range *spList {
					fullname := common.ExtractResourceNameFromArn(sp.Arn)
					if sp.Name == fullname {
						shortName, _ := duplosdk.UnprefixName(prefix, fullname)
						spResourceName := common.GetResourceName(shortName)
						jqBody.SetAttributeTraversal("scheduling_policy_arn", hcl.Traversal{
							hcl.TraverseRoot{
								Name: "duplocloud_aws_batch_scheduling_policy." + spResourceName,
							},
							hcl.TraverseAttr{
								Name: "arn",
							},
						})
						break
					}
				}
			}

			if q.ComputeEnvironmentOrder != nil && len(*q.ComputeEnvironmentOrder) > 0 {
				sort.Slice(*q.ComputeEnvironmentOrder, func(i, j int) bool {
					return (*q.ComputeEnvironmentOrder)[i].Order < (*q.ComputeEnvironmentOrder)[j].Order
				})
				tokens := hclwrite.Tokens{
					{Type: hclsyntax.TokenOQuote, Bytes: []byte(`[`)},
				}
				// orderedTokens := make([]*hclwrite.Token, len(*q.ComputeEnvironmentOrder))
				for _, ceo := range *q.ComputeEnvironmentOrder {
					for i, ce := range *ceList {
						fullname := common.ExtractResourceNameFromArn(ceo.ComputeEnvironment)
						if ce.ComputeEnvironmentName == fullname {
							shortName, _ := duplosdk.UnprefixName(prefix, fullname)
							ceResourceName := common.GetResourceName(shortName)
							arn := "duplocloud_aws_batch_compute_environment." + ceResourceName + ".arn"
							// orderedTokens[ceo.Order] = &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(arn)}
							tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(arn)})
							if i < len(*q.ComputeEnvironmentOrder)-1 {
								tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(",")})
							}
							break
						}
					}
				}
				// for i, t := range orderedTokens {
				// 	tokens = append(tokens, t)
				// 	if i < len(*q.ComputeEnvironmentOrder)-1 {
				// 		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(",")})
				// 	}
				// }
				tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenEOF, Bytes: []byte(`]`)})
				jqBody.SetAttributeRaw("compute_environments", tokens)
			}

			log.Printf("[TRACE] Generating tf file for resource : %s", shortName)
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo AWS Batch Job Queue: %s", shortName)

			outVars := generateBatchJQOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_batch_job_queue." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== AWS Batch Job Queue TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateBatchJQOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_batch_job_queue." + resourceName + ".arn",
		DescVal:       "The Amazon Resource Name of the job queue.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var1

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
