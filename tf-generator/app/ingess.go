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
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type K8sIngress struct {
}

func (k8sIngress *K8sIngress) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AppProject)
	list, clientErr := client.DuploK8sIngressGetList(config.TenantId)
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, nil
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== Duplo K8S Ingress TF generation started. =====>")
		for _, k8sIngress := range *list {
			log.Printf("[TRACE] Generating terraform config for duplo k8s ingress : %s", k8sIngress.Name)
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "k8s-ingress-"+k8sIngress.Name+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			resourceName := common.GetResourceName(k8sIngress.Name)
			// initialize the body of the new file object
			rootBody := hclFile.Body()
			// Add duplocloud_aws_host resource
			k8sIngressBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_k8_ingress",
					resourceName})
			k8sIngressBody := k8sIngressBlock.Body()
			k8sIngressBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			k8sIngressBody.SetAttributeValue("name",
				cty.StringVal(k8sIngress.Name))
			k8sIngressBody.SetAttributeValue("ingress_class_name",
				cty.StringVal(k8sIngress.IngressClassName))

			if len(k8sIngress.Annotations) > 0 {
				newMap := make(map[string]cty.Value)
				for key, element := range k8sIngress.Annotations {
					newMap[key] = cty.StringVal(element)
				}
				k8sIngressBody.SetAttributeValue("annotations", cty.ObjectVal(newMap))
			}

			if len(k8sIngress.Labels) > 0 {
				newMap := make(map[string]cty.Value)
				for key, element := range k8sIngress.Labels {
					newMap[key] = cty.StringVal(element)
				}
				k8sIngressBody.SetAttributeValue("labels", cty.ObjectVal(newMap))
			}

			if k8sIngress.LbConfig != nil {
				lbcBlock := k8sIngressBody.AppendNewBlock("lbconfig",
					nil)
				lbcBody := lbcBlock.Body()

				if len(k8sIngress.LbConfig.DnsPrefix) > 0 {
					dnsPrefix := ""
					if strings.Contains(k8sIngress.LbConfig.DnsPrefix, config.TenantName) {
						dnsPrefix = strings.Replace(k8sIngress.LbConfig.DnsPrefix, config.TenantName, "${local.tenant_name}", -1)
					} else {
						dnsPrefix = dnsPrefix + "-${local.tenant_name}"
					}
					//dnsPrefix = dnsPrefix + "-${local.tenant_name}"
					dnsPrefixTokens := hclwrite.Tokens{
						{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
						{Type: hclsyntax.TokenIdent, Bytes: []byte(dnsPrefix)},
						{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
					}
					lbcBody.SetAttributeRaw("dns_prefix", dnsPrefixTokens)
				}
				lbcBody.SetAttributeValue("is_internal", cty.BoolVal(!k8sIngress.LbConfig.IsPublic))
				if len(k8sIngress.LbConfig.CertArn) > 0 {
					lbcBody.SetAttributeTraversal("certificate_arn", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "local",
						},
						hcl.TraverseAttr{
							Name: "cert_arn",
						},
					})
				}
				if k8sIngress.LbConfig.Listeners != nil && len(k8sIngress.LbConfig.Listeners.Https) > 0 {
					lbcBody.SetAttributeValue("https_port", cty.NumberIntVal(int64(k8sIngress.LbConfig.Listeners.Https[0])))
				}
				if k8sIngress.LbConfig.Listeners != nil && len(k8sIngress.LbConfig.Listeners.Http) > 0 {
					lbcBody.SetAttributeValue("http_port", cty.NumberIntVal(int64(k8sIngress.LbConfig.Listeners.Http[0])))
				}
			}

			if len(*k8sIngress.Rules) > 0 {
				for _, rule := range *k8sIngress.Rules {
					ruleBlock := k8sIngressBody.AppendNewBlock("rule",
						nil)
					ruleBody := ruleBlock.Body()
					ruleBody.SetAttributeValue("path", cty.StringVal(rule.Path))
					ruleBody.SetAttributeValue("path_type", cty.StringVal(rule.PathType))
					ruleBody.SetAttributeValue("port", cty.NumberIntVal(int64(rule.Port)))
					if len(rule.Host) > 0 {
						ruleBody.SetAttributeValue("host", cty.StringVal(rule.Host))
					}
					ruleBody.SetAttributeTraversal("service_name", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "duplocloud_duplo_service." + common.GetResourceName(rule.ServiceName),
						},
						hcl.TraverseAttr{
							Name: "name",
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
					ResourceAddress: "duplocloud_k8_ingress." + resourceName,
					ResourceId:      "v3/subscriptions/" + config.TenantId + "/k8s/ingress/" + k8sIngress.Name,
					WorkingDir:      workingDir,
				})

				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== Duplo K8S Ingress TF generation done. =====>")
	}
	return &tfContext, nil
}
