package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const LF_VAR_PREFIX = "lf_"

type LambdaFunction struct {
}

func (lf *LambdaFunction) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.LambdaFunctionGetList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== Lambda Function TF generation started. =====>")
		for _, lf := range *list {
			shortName := lf.Name
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for lammbda funtion : %s", shortName)

			lfDetails, clientErr := client.LambdaFunctionGet(config.TenantId, lf.FunctionName)
			if clientErr != nil {
				fmt.Println(clientErr)
				continue
			}
			varFullPrefix := LF_VAR_PREFIX + resourceName + "_"
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "lf-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_ecache_instance resource
			lfBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_lambda_function",
					resourceName})
			lfBody := lfBlock.Body()
			lfBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// kafkaBody.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			lfBody.SetAttributeValue("name",
				cty.StringVal(shortName))

			if len(lf.Description) > 0 {
				lfBody.SetAttributeValue("description",
					cty.StringVal(lf.Description))
			}
			if lf.PackageType != nil && len(lf.PackageType.Value) > 0 {
				lfBody.SetAttributeValue("package_type",
					cty.StringVal(lf.PackageType.Value))
				if strings.ToLower(lf.PackageType.Value) == strings.ToLower("Zip") {
					if len(lfDetails.Code.S3Bucket) > 0 {
						lfBody.SetAttributeValue("s3_bucket",
							cty.StringVal(lfDetails.Code.S3Bucket))
					}
					if len(lfDetails.Code.S3Key) > 0 {
						lfBody.SetAttributeValue("s3_key",
							cty.StringVal(lfDetails.Code.S3Key))
					}
				} else if strings.ToLower(lf.PackageType.Value) == strings.ToLower("Image") {
					lfBody.SetAttributeValue("image_uri",
						cty.StringVal(lfDetails.Code.ImageURI))
				}
			}
			lfBody.SetAttributeValue("memory_size",
				cty.NumberIntVal(int64(lf.MemorySize)))
			lfBody.SetAttributeValue("timeout",
				cty.NumberIntVal(int64(lf.Timeout)))
			if len(lf.Handler) > 0 {
				lfBody.SetAttributeValue("handler",
					cty.StringVal(lf.Handler))
			}
			if lf.Runtime != nil && len(lf.Runtime.Value) > 0 {
				lfBody.SetAttributeValue("runtime",
					cty.StringVal(lf.Runtime.Value))
			}
			if len(*lf.Layers) > 0 {
				var layers []cty.Value
				for _, l := range *lf.Layers {
					layers = append(layers, cty.StringVal(l.Arn))
				}
				lfBody.SetAttributeValue("layers", cty.ListVal(layers))

			}
			if lf.Environment != nil && len(lf.Environment.Variables) > 0 {
				envBlock := lfBody.AppendNewBlock("environment", nil)
				envBody := envBlock.Body()
				newMap := make(map[string]cty.Value)
				for key, element := range lf.Environment.Variables {
					newMap[key] = cty.StringVal(element)
				}
				envBody.SetAttributeValue("variables", cty.MapVal(newMap))
			}

			// Lambda Permission Resource
			lfPermission, _ := client.LambdaPermissionGet(config.TenantId, lf.FunctionName)
			if lfPermission != nil && len(*lfPermission) > 0 {
				for i, lfPerm := range *lfPermission {
					index := strconv.Itoa(i)
					lfPermBlock := rootBody.AppendNewBlock("resource",
						[]string{"duplocloud_aws_lambda_permission",
							shortName + "-permission" + index})
					lfPermBody := lfPermBlock.Body()
					lfPermBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "local",
						},
						hcl.TraverseAttr{
							Name: "tenant_id",
						},
					})
					lfPermBody.SetAttributeTraversal("function_name", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "duplocloud_aws_lambda_function." + resourceName,
						},
						hcl.TraverseAttr{
							Name: "name",
						},
					})
					lfPermBody.SetAttributeValue("action",
						cty.StringVal(lfPerm.Action))
					lfPermBody.SetAttributeValue("principal",
						cty.StringVal(lfPerm.Principal.Service))
					lfPermBody.SetAttributeValue("statement_id",
						cty.StringVal(lfPerm.Sid))
				}
			}
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			log.Printf("[TRACE] Terraform config is generated for lambda function : %s", shortName)

			outVars := generateLFOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_lambda_function." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== Lambda Function TF generation done. =====>")
	}
	return &tfContext, nil
}

func generateLFOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "fullname",
		ActualVal:     "duplocloud_aws_lambda_function." + resourceName + ".fullname",
		DescVal:       "The full name of the lambda function.",
		RootTraversal: true,
	}
	outVarConfigs["fullname"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_lambda_function." + resourceName + ".arn",
		DescVal:       "The ARN of the lambda function.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "version",
		ActualVal:     "duplocloud_aws_lambda_function." + resourceName + ".version",
		DescVal:       "The version of the lambda function.",
		RootTraversal: true,
	}
	outVarConfigs["version"] = var3

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
