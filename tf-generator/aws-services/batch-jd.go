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

const BJD_VAR_PREFIX = "batch_jd_"

type BatchJD struct {
}

func (bjd *BatchJD) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.AwsBatchJobDefinitionList(config.TenantId)

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
		log.Println("[TRACE] <====== AWS Batch Job Definition TF generation started. =====>")
		for _, jd := range *list {
			shortName, _ := duplosdk.UnprefixName(prefix, jd.JobDefinitionName)
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo AWS Batch Job Definition: %s", shortName)

			varFullPrefix := BJD_VAR_PREFIX + resourceName + "_"

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "batch-jd-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_batch_job_definition resource
			jdBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_batch_job_definition",
					resourceName})
			jdBody := jdBlock.Body()
			jdBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			jdBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			if len(jd.Type) > 0 {
				jdBody.SetAttributeValue("type",
					cty.StringVal(jd.Type))
			} else {
				jdBody.SetAttributeValue("type",
					cty.StringVal("container"))
			}

			if len(jd.PlatformCapabilities) > 0 {
				var vals []cty.Value
				for _, s := range jd.PlatformCapabilities {
					vals = append(vals, cty.StringVal(s))
				}
				jdBody.SetAttributeValue("platform_capabilities",
					cty.ListVal(vals))
			}

			if jd.Timeout != nil {
				toBlock := jdBody.AppendNewBlock("timeout",
					nil)
				toBody := toBlock.Body()
				toBody.SetAttributeValue("attempt_duration_seconds",
					cty.NumberIntVal(int64(jd.Timeout.AttemptDurationSeconds)))

			}

			if jd.RetryStrategy != nil {
				rsBlock := jdBody.AppendNewBlock("retry_strategy",
					nil)
				rsBody := rsBlock.Body()
				if jd.RetryStrategy.Attempts > 0 {
					rsBody.SetAttributeValue("attempts",
						cty.NumberIntVal(int64(jd.RetryStrategy.Attempts)))
				}
				if jd.RetryStrategy.EvaluateOnExit != nil && len(*jd.RetryStrategy.EvaluateOnExit) > 0 {
					for _, eoe := range *jd.RetryStrategy.EvaluateOnExit {
						eoeBlock := rsBody.AppendNewBlock("evaluate_on_exit",
							nil)
						eoeBody := eoeBlock.Body()
						if eoe.Action != nil && len(eoe.Action.Value) > 0 {
							eoeBody.SetAttributeValue("action",
								cty.StringVal(eoe.Action.Value))
						}
						if len(eoe.OnExitCode) > 0 {
							eoeBody.SetAttributeValue("on_exit_code",
								cty.StringVal(eoe.OnExitCode))
						}
						if len(eoe.OnReason) > 0 {
							eoeBody.SetAttributeValue("on_reason",
								cty.StringVal(eoe.OnReason))
						}
						if len(eoe.OnStatusReason) > 0 {
							eoeBody.SetAttributeValue("on_status_reason",
								cty.StringVal(eoe.OnStatusReason))
						}
					}
				}
			}

			if len(jd.Parameters) > 0 {
				newMap := make(map[string]cty.Value)
				for key, element := range jd.Parameters {
					newMap[key] = cty.StringVal(element)
				}
				jdBody.SetAttributeValue("parameters", cty.ObjectVal(newMap))
			}

			if len(jd.ContainerProperties) > 0 {
				reduceJobContainerProperties(jd.ContainerProperties)
				heredocstr := "<<CONTAINER_PROPERTIES"
				containerPropsTokens := hclwrite.Tokens{
					{Type: hclsyntax.TokenOQuote, Bytes: []byte(heredocstr)},
					// {Type: hclsyntax.TokenMinus, Bytes: []byte(`-`)},
					// {Type: hclsyntax.TokenEOF, Bytes: []byte(`EOT`)},
					{Type: hclsyntax.TokenIdent, Bytes: []byte("\n")},
				}
				containerPropsStr, err := duplosdk.JSONMarshal(jd.ContainerProperties)
				if err != nil {
					panic(err)
				}
				containerPropsTokens = append(containerPropsTokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(containerPropsStr)})
				containerPropsTokens = append(containerPropsTokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte("\n")})
				containerPropsTokens = append(containerPropsTokens, &hclwrite.Token{Type: hclsyntax.TokenEOF, Bytes: []byte(`CONTAINER_PROPERTIES`)})
				jdBody.SetAttributeRaw("container_properties", containerPropsTokens)
			}
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo AWS Batch Job Definition: %s", shortName)

			outVars := generateBatchJDOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_batch_job_definition." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== AWS Batch Job Definition TF generation done. =====>")
	}

	return &tfContext, nil
}

func generateBatchJDOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_batch_job_definition." + resourceName + ".arn",
		DescVal:       "The Amazon Resource Name of the job Definition.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var1

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}

func reduceJobContainerProperties(props map[string]interface{}) error {
	common.MakeMapUpperCamelCase(props)

	reorderJobContainerPropertiesEnvironmentVariables(props)

	common.ReduceNilOrEmptyMapEntries(props)

	// Handle fields that have defaults.
	if v, ok := props["ReadonlyRootFilesystem"]; ok || v != nil && !v.(bool) {
		delete(props, "ReadonlyRootFilesystem")
	}

	if v, ok := props["Privileged"]; ok || v != nil && !v.(bool) {
		delete(props, "Privileged")
	}

	if v, ok := props["Memory"]; ok || v != nil && v.(int) == 0 {
		delete(props, "Memory")
	}
	if v, ok := props["Vcpus"]; ok || v != nil && v.(int) == 0 {
		delete(props, "Vcpus")
	}
	delete(props, "JobRoleArn")

	return nil
}

func reorderJobContainerPropertiesEnvironmentVariables(defn map[string]interface{}) {

	// Re-order environment variables to a canonical order.
	if v, ok := defn["Environment"]; ok && v != nil {
		if env, ok := v.([]interface{}); ok && env != nil {
			sort.SliceStable(env, func(i, j int) bool {

				// Get both maps, ensure we are using upper camel-case.
				mi := env[i].(map[string]interface{})
				mj := env[j].(map[string]interface{})
				common.MakeMapUpperCamelCase(mi)
				common.MakeMapUpperCamelCase(mj)

				// Get both name keys, fall back on an empty string.
				si := ""
				sj := ""
				if v, ok = mi["Name"]; ok && !common.IsInterfaceNil(v) {
					si = v.(string)
				}
				if v, ok = mj["Name"]; ok && !common.IsInterfaceNil(v) {
					sj = v.(string)
				}

				// Compare the two.
				return si < sj
			})
		}
	}
}
