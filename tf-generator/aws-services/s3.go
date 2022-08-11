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

const S3_VAR_PREFIX = "s3_"

type S3Bucket struct {
}

func (s3 *S3Bucket) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantListS3Buckets(config.TenantId)

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== S3 bucket TF generation started. =====>")
		for _, s3 := range *list {
			shortName := s3.Name
			if strings.HasPrefix(s3.Name, "duploservices-") {
				shortName = s3.Name[len("duploservices-"+config.TenantName+"-"):len(s3.Name)]
				parts := strings.Split(shortName, "-")
				if len(parts) > 0 {
					parts = parts[:len(parts)-1]
				}
				shortName = strings.Join(parts, "-")
			}
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo s3 bucket : %s", shortName)
			varFullPrefix := S3_VAR_PREFIX + resourceName + "_"
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			inputVars := generateS3Vars(s3, varFullPrefix)
			tfContext.InputVars = append(tfContext.InputVars, inputVars...)

			// create new file on system
			path := filepath.Join(workingDir, "s3-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_s3_bucket resource
			s3Block := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_s3_bucket",
					resourceName})
			s3Body := s3Block.Body()
			s3Body.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// s3Body.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			s3Body.SetAttributeValue("name",
				cty.StringVal(shortName))

			s3Body.SetAttributeTraversal("allow_public_access", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "allow_public_access",
				},
			})

			s3Body.SetAttributeTraversal("enable_access_logs", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "enable_access_logs",
				},
			})

			s3Body.SetAttributeTraversal("enable_versioning", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "var",
				},
				hcl.TraverseAttr{
					Name: varFullPrefix + "enable_versioning",
				},
			})

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
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo s3 bucket : %s", shortName)

			outVars := generateS3OutputVars(s3, varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_s3_bucket." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== S3 Bucket TF generation done. =====>")
	}
	return &tfContext, nil
}

func generateS3Vars(duplo duplosdk.DuploS3Bucket, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	var1 := common.VarConfig{
		Name:       prefix + "allow_public_access",
		DefaultVal: strconv.FormatBool(duplo.AllowPublicAccess),
		TypeVal:    "bool",
	}
	varConfigs["allow_public_access"] = var1

	var2 := common.VarConfig{
		Name:       prefix + "enable_access_logs",
		DefaultVal: strconv.FormatBool(duplo.EnableAccessLogs),
		TypeVal:    "bool",
	}
	varConfigs["enable_access_logs"] = var2

	var3 := common.VarConfig{
		Name:       prefix + "enable_versioning",
		DefaultVal: strconv.FormatBool(duplo.EnableVersioning),
		TypeVal:    "bool",
	}
	varConfigs["enable_versioning"] = var3

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}

func generateS3OutputVars(duplo duplosdk.DuploS3Bucket, prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	fullNameVar := common.OutputVarConfig{
		Name:          prefix + "fullname",
		ActualVal:     "duplocloud_s3_bucket." + resourceName + ".fullname",
		DescVal:       "The full name of the S3 bucket.",
		RootTraversal: true,
	}
	outVarConfigs["fullname"] = fullNameVar

	arnVar := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_s3_bucket." + resourceName + ".arn",
		DescVal:       "The ARN of the S3 bucket.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = arnVar

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
