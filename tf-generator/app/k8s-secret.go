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

const EXCLUDE_K8S_SECRET_STR = "default-token,duploservices-,filebeat-token-"

type K8sSecret struct {
}

func (k8sSecret *K8sSecret) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	log.Println("[TRACE] <====== Duplo K8S Secret TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AppProject)
	list, clientErr := client.K8SecretGetList(config.TenantId)
	exclude_k8s_secret_list := strings.Split(EXCLUDE_K8S_SECRET_STR, ",")
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, nil
	}
	tfContext := common.TFContext{}
	if list != nil {
		for _, k8sSecret := range *list {
			log.Printf("[TRACE] Generating terraform config for duplo k8s secret : %s", k8sSecret.SecretName)
			skip := false
			for _, element := range exclude_k8s_secret_list {
				if strings.Contains(k8sSecret.SecretName, element) {
					log.Printf("[TRACE] Generating terraform config for duplo k8s secret : %s skipped.", k8sSecret.SecretName)
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
			path := filepath.Join(workingDir, "k8s-secret-"+k8sSecret.SecretName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			resourceName := strings.ReplaceAll(k8sSecret.SecretName, ".", "_")
			// initialize the body of the new file object
			rootBody := hclFile.Body()
			// Add duplocloud_aws_host resource
			k8sSecretBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_k8_secret",
					resourceName})
			k8sSecretBody := k8sSecretBlock.Body()
			k8sSecretBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			k8sSecretBody.SetAttributeValue("secret_name",
				cty.StringVal(k8sSecret.SecretName))
			k8sSecretBody.SetAttributeValue("secret_type",
				cty.StringVal(k8sSecret.SecretType))

			if len(k8sSecret.SecretAnnotations) > 0 {
				newMap := make(map[string]cty.Value)
				for key, element := range k8sSecret.SecretAnnotations {
					newMap[key] = cty.StringVal(element)
				}
				k8sSecretBody.SetAttributeValue("secret_annotations", cty.ObjectVal(newMap))
			}

			if len(k8sSecret.SecretData) > 0 {
				secretDataStr, err := duplosdk.JSONMarshal(k8sSecret.SecretData)
				if err != nil {
					panic(err)
				}
				k8sSecretBody.SetAttributeTraversal("secret_data", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + secretDataStr + ")",
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
				importConfigs := []common.ImportConfig{}

				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_k8_secret." + resourceName,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/K8SecretApiV2/" + k8sSecret.SecretName,
					WorkingDir:      workingDir,
				})

				tfContext.ImportConfigs = importConfigs
			}
		}

	}

	log.Println("[TRACE] <====== Duplo K8S Secret TF generation done. =====>")
	return &tfContext, nil
}
