package app

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

const EXCLUDE_K8S_CONFIG_STR = "kube-root-ca.crt"

type K8sConfig struct {
}

func (k8sConfig *K8sConfig) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AppProject)
	list, clientErr := client.K8ConfigMapGetList(config.TenantId)
	exclude_k8s_config_list := strings.Split(EXCLUDE_K8S_CONFIG_STR, ",")
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, nil
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== Duplo K8S Config Map TF generation started. =====>")
		for _, k8sConfig := range *list {
			log.Printf("[TRACE] Generating terraform config for duplo k8s config map : %s", k8sConfig.Name)
			skip := false
			for _, element := range exclude_k8s_config_list {
				if strings.Contains(k8sConfig.Name, element) {
					log.Printf("[TRACE] Generating terraform config for duplo k8s config map : %s skipped.", k8sConfig.Name)
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "k8s-cm-"+k8sConfig.Name+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			resourceName := common.GetResourceName(k8sConfig.Name)
			// initialize the body of the new file object
			rootBody := hclFile.Body()
			// Add duplocloud_aws_host resource
			k8sConfigBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_k8_config_map",
					resourceName})
			k8sConfigBody := k8sConfigBlock.Body()
			k8sConfigBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			k8sConfigBody.SetAttributeValue("name",
				cty.StringVal(k8sConfig.Name))

			if len(k8sConfig.Data) > 0 {
				configDataStr, err := duplosdk.JSONMarshal(EscapeDollarEscapes(k8sConfig.Data))
				if err != nil {
					panic(err)
				}
				k8sConfigBody.SetAttributeTraversal("data", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + configDataStr + ")",
					},
				})
			}

			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// Import all created resources.
			if config.GenerateTfState {

				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_k8_config_map." + resourceName,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/K8ConfigMapApiV2/" + k8sConfig.Name,
					WorkingDir:      workingDir,
				})

				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== Duplo K8S Config Map TF generation done. =====>")
	}

	return &tfContext, nil
}

func EscapeDollarEscapes(data map[string]interface{}) map[string]interface{} {
	var processMap func(map[string]interface{}) map[string]interface{}
	processMap = func(input map[string]interface{}) map[string]interface{} {
		result := make(map[string]interface{})
		for key, value := range input {
			switch v := value.(type) {
			case string:
				if strings.Contains(v, "${") && !strings.Contains(v, "$${") {
					result[key] = strings.ReplaceAll(v, "${", "$${")
				} else {
					result[key] = v
				}
			case map[string]interface{}:
				// Process nested maps recursively
				result[key] = processMap(v)
			default:
				// Leave non-string values unchanged
				result[key] = v
			}
		}
		return result
	}

	// Process the top-level map
	return processMap(data)

}
