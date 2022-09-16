package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type K8sSecretProviderClass struct {
}

func (spc *K8sSecretProviderClass) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AppProject)
	list, clientErr := client.DuploK8sSecretProviderClassList(config.TenantId)
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, nil
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== Duplo K8S Secret Provider Class TF generation started. =====>")
		for _, secretProvClass := range *list {
			log.Printf("[TRACE] Generating terraform config for duplo secret provider class : %s", secretProvClass.Name)
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "k8s-provider-class-"+secretProvClass.Name+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			resourceName := common.GetResourceName(secretProvClass.Name)
			// initialize the body of the new file object
			rootBody := hclFile.Body()
			// Add duplocloud_aws_host resource
			spcBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_k8_secret_provider_class",
					resourceName})
			spcBody := spcBlock.Body()
			spcBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			spcBody.SetAttributeValue("name",
				cty.StringVal(secretProvClass.Name))
			spcBody.SetAttributeValue("secret_provider",
				cty.StringVal(secretProvClass.Provider))

			if len(secretProvClass.Annotations) > 0 {
				newMap := make(map[string]cty.Value)
				for key, element := range secretProvClass.Annotations {
					newMap[key] = cty.StringVal(element)
				}
				spcBody.SetAttributeValue("annotations", cty.ObjectVal(newMap))
			}

			if len(secretProvClass.Labels) > 0 {
				newMap := make(map[string]cty.Value)
				for key, element := range secretProvClass.Labels {
					newMap[key] = cty.StringVal(element)
				}
				spcBody.SetAttributeValue("labels", cty.ObjectVal(newMap))
			}

			if len(*secretProvClass.SecretObjects) > 0 {
				for _, so := range *secretProvClass.SecretObjects {
					soBlock := spcBody.AppendNewBlock("secret_object", nil)
					soBody := soBlock.Body()
					soBody.SetAttributeValue("name", cty.StringVal(so.SecretName))
					soBody.SetAttributeValue("type", cty.StringVal(so.Type))
					if len(so.Annotations) > 0 {
						newMap := make(map[string]cty.Value)
						for key, element := range so.Annotations {
							newMap[key] = cty.StringVal(element)
						}
						soBody.SetAttributeValue("annotations", cty.ObjectVal(newMap))
					}

					if len(so.Labels) > 0 {
						newMap := make(map[string]cty.Value)
						for key, element := range so.Labels {
							newMap[key] = cty.StringVal(element)
						}
						soBody.SetAttributeValue("labels", cty.ObjectVal(newMap))
					}
					for _, d := range *so.Data {
						dataBlock := soBody.AppendNewBlock("data", nil)
						dataBody := dataBlock.Body()
						dataBody.SetAttributeValue("key", cty.StringVal(d.Key))
						dataBody.SetAttributeValue("object_name", cty.StringVal(d.ObjectName))
					}
				}
			}

			if secretProvClass.Parameters != nil && len(secretProvClass.Parameters.Objects) > 0 {
				ymlStr := secretProvClass.Parameters.Objects
				j2, err := yaml.YAMLToJSON([]byte(ymlStr))
				if err != nil {
					fmt.Printf("err: %v\n", err)
					return nil, nil
				}
				response := []interface{}{}
				err = json.Unmarshal(j2, &response)
				if err != nil {
					panic(err)
				}
				if len(response) > 0 {
					for _, r := range response {
						rr := r.(map[string]interface{})
						objeName := rr["objectName"]
						if strings.Contains(objeName.(string), config.TenantName) {
							rr["objectName"] = strings.Replace(objeName.(string), "-"+config.TenantName+"-", "-${local.tenant_name}-", -1)
						}
					}
					paramString, err := duplosdk.JSONMarshal(response)
					if err != nil {
						panic(err)
					}
					spcBody.SetAttributeTraversal("parameters", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "jsonencode(" + paramString + ")",
						},
					})
				}

			}

			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_k8_secret_provider_class." + resourceName,
					ResourceId:      "v3/subscriptions/" + config.TenantId + "/k8s/secretproviderclass/" + secretProvClass.Name,
					WorkingDir:      workingDir,
				})

				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== Duplo K8S Secret Provider Class TF generation done. =====>")
	}
	return &tfContext, nil
}

func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}
