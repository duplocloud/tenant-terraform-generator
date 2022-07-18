package app

import (
	"encoding/json"
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

const SVC_VAR_PREFIX = "svc_"
const EXCLUDE_SVC_STR = "duploinfrasvc,dockerservices-shell,system-svc-"

type Services struct {
}

func (s *Services) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	log.Println("[TRACE] <====== Duplo Services TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AppProject)
	list, clientErr := client.ReplicationControllerList(config.TenantId)
	exclude_svc_list := strings.Split(EXCLUDE_SVC_STR, ",")
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	if list != nil {
		for _, service := range *list {
			log.Printf("[TRACE] Generating terraform config for duplo service : %s", service.Name)
			skip := false
			for _, element := range exclude_svc_list {
				if strings.Contains(service.Name, element) {
					log.Printf("[TRACE] Generating terraform config for duplo service : %s skipped.", service.Name)
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			varFullPrefix := SVC_VAR_PREFIX + strings.ReplaceAll(service.Name, "-", "_") + "_"
			inputVars := generateSvcVars(service, varFullPrefix)
			tfContext.InputVars = append(tfContext.InputVars, inputVars...)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "svc-"+service.Name+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()
			// Add duplocloud_aws_host resource
			svcBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_duplo_service",
					service.Name})
			svcBody := svcBlock.Body()
			svcBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// svcBody.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			svcBody.SetAttributeValue("name",
				cty.StringVal(service.Name))

			svcBody.SetAttributeValue("replicas",
				cty.NumberIntVal(int64(service.Replicas)))
			svcBody.SetAttributeValue("lb_synced_deployment",
				cty.BoolVal(service.IsLBSyncedDeployment))
			svcBody.SetAttributeValue("cloud_creds_from_k8s_service_account",
				cty.BoolVal(service.IsCloudCredsFromK8sServiceAccount))
			svcBody.SetAttributeValue("is_daemonset",
				cty.BoolVal(service.IsDaemonset))
			if len(service.ReplicasMatchingAsgName) > 0 {
				svcBody.SetAttributeValue("replicas_matching_asg_name",
					cty.StringVal(service.ReplicasMatchingAsgName))
			}

			if service.Template != nil {
				svcBody.SetAttributeValue("agent_platform",
					cty.NumberIntVal(int64(service.Template.AgentPlatform)))
				svcBody.SetAttributeValue("cloud",
					cty.NumberIntVal(int64(service.Template.Cloud)))
				if len(service.Template.AllocationTags) > 0 {
					svcBody.SetAttributeValue("allocation_tags",
						cty.StringVal(service.Template.AllocationTags))
				}
				if len(service.Template.OtherDockerConfig) > 0 {
					OtherDockerConfigMap := make(map[string]interface{})
					err := json.Unmarshal([]byte(service.Template.OtherDockerConfig), &OtherDockerConfigMap)
					if err != nil {
						panic(err)
					}
					OtherDockerConfigStr, err := duplosdk.JSONMarshal(OtherDockerConfigMap)
					if err != nil {
						panic(err)
					}
					svcBody.SetAttributeTraversal("other_docker_config", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "jsonencode(" + OtherDockerConfigStr + ")",
						},
					})
				}
				if len(service.Template.ExtraConfig) > 0 {
					var extraConfigMap interface{}
					log.Printf("[TRACE] ExtraConfig *** : %s", service.Template.ExtraConfig)
					err := json.Unmarshal([]byte(service.Template.ExtraConfig), &extraConfigMap)
					if err != nil {
						panic(err)
					}
					extraConfigStr, err := duplosdk.JSONMarshal(extraConfigMap)
					if err != nil {
						panic(err)
					}
					svcBody.SetAttributeTraversal("extra_config", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "jsonencode(" + extraConfigStr + ")",
						},
					})
				}
				if len(service.Template.OtherDockerHostConfig) > 0 {
					OtherDockerHostConfigMap := make(map[string]interface{})
					err := json.Unmarshal([]byte(service.Template.OtherDockerHostConfig), &OtherDockerHostConfigMap)
					if err != nil {
						panic(err)
					}
					OtherDockerHostConfigStr, err := duplosdk.JSONMarshal(OtherDockerHostConfigMap)
					if err != nil {
						panic(err)
					}
					svcBody.SetAttributeTraversal("other_docker_host_config", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "jsonencode(" + OtherDockerHostConfigStr + ")",
						},
					})
				}

				if service.Template.Commands != nil && len(service.Template.Commands) > 0 {
					var vals []cty.Value
					for _, cmd := range service.Template.Commands {
						vals = append(vals, cty.StringVal(cmd))
					}
					svcBody.SetAttributeValue("commands",
						cty.ListVal(vals))
				}

				// If there is at least one container, get the first docker image from it.
				if service.Template.Containers != nil && len(*service.Template.Containers) > 0 {
					svcBody.SetAttributeTraversal("docker_image",
						hcl.Traversal{
							hcl.TraverseRoot{
								Name: "var",
							},
							hcl.TraverseAttr{
								Name: varFullPrefix + "docker_image",
							},
						})
				}

				if len(service.Template.Volumes) > 0 {
					//log.Printf("[TRACE] Volume : %s", service.Template.Volumes)
					//volConfigMap := make(map[string]interface{})
					var volConfigMapList interface{}
					// log.Printf("[TRACE] Vol *** : %s", service.Template.Volumes)
					err := json.Unmarshal([]byte(service.Template.Volumes), &volConfigMapList)
					if err != nil {
						panic(err)
					}
					volConfigMapStr, err := duplosdk.JSONMarshal(volConfigMapList)
					if err != nil {
						panic(err)
					}
					svcBody.SetAttributeTraversal("volumes", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "jsonencode(" + volConfigMapStr + ")",
						},
					})
				}
			}
			log.Printf("[TRACE] Terraform config is generated for duplo service : %s", service.Name)
			rootBody.AppendNewline()
			configList, clientErr := client.ReplicationControllerLbConfigurationList(config.TenantId, service.Name)
			if clientErr != nil {
				fmt.Println(clientErr)
				return nil, clientErr
			}
			configPresent := false
			if configList != nil && len(*configList) > 0 {
				configPresent = true
				svcConfigBlock := rootBody.AppendNewBlock("resource",
					[]string{"duplocloud_duplo_service_lbconfigs",
						service.Name + "-config"})
				svcConfigBody := svcConfigBlock.Body()
				svcConfigBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "duplocloud_duplo_service." + service.Name,
					},
					hcl.TraverseAttr{
						Name: "tenant_id",
					},
				})
				svcConfigBody.SetAttributeTraversal("replication_controller_name", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "duplocloud_duplo_service." + service.Name,
					},
					hcl.TraverseAttr{
						Name: "name",
					},
				})
				for _, serviceConfig := range *configList {
					lbConfigBlock := svcConfigBody.AppendNewBlock("lbconfigs",
						nil)
					lbConfigBlockBody := lbConfigBlock.Body()
					lbConfigBlockBody.SetAttributeValue("lb_type",
						cty.NumberIntVal(int64(serviceConfig.LbType)))
					lbConfigBlockBody.SetAttributeValue("is_native",
						cty.BoolVal(serviceConfig.IsNative))
					lbConfigBlockBody.SetAttributeValue("is_internal",
						cty.BoolVal(serviceConfig.IsInternal))
					port, err := strconv.Atoi(serviceConfig.Port)
					if err != nil {
						fmt.Println(err)
						return nil, err
					}
					lbConfigBlockBody.SetAttributeValue("port",
						cty.NumberIntVal(int64(port)))
					lbConfigBlockBody.SetAttributeValue("external_port",
						cty.NumberIntVal(int64(serviceConfig.ExternalPort)))
					lbConfigBlockBody.SetAttributeValue("protocol",
						cty.StringVal(serviceConfig.Protocol))
					lbConfigBlockBody.SetAttributeValue("health_check_url",
						cty.StringVal(serviceConfig.HealthCheckURL))
					lbConfigBlockBody.SetAttributeValue("certificate_arn",
						cty.StringVal(serviceConfig.CertificateArn))
					svcConfigBody.AppendNewline()
				}

				// TODO Add duplocloud_duplo_service_params resource here.

				svcParamBlock := rootBody.AppendNewBlock("resource",
					[]string{"duplocloud_duplo_service_params",
						service.Name + "-params"})
				svcParamBody := svcParamBlock.Body()
				svcParamBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "duplocloud_duplo_service_lbconfigs." + service.Name + "-config",
					},
					hcl.TraverseAttr{
						Name: "tenant_id",
					},
				})
				svcParamBody.SetAttributeTraversal("replication_controller_name", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "duplocloud_duplo_service_lbconfigs." + service.Name + "-config",
					},
					hcl.TraverseAttr{
						Name: "replication_controller_name",
					},
				})
				svcParamBody.SetAttributeValue("dns_prfx",
					cty.StringVal(service.DnsPrfx))
				if doesReplicationControllerHaveAlb(&service) {
					webAclId, clientError := client.ReplicationControllerLbWafGet(config.TenantId, service.Name)
					if clientError != nil {
						if clientError.Status() == 500 && service.Template.Cloud != 0 {
							log.Printf("[TRACE] Ignoring error %s for non AWS cloud.", clientError)
						}
						webAclId = ""
					}
					if len(webAclId) > 0 {
						svcParamBody.SetAttributeValue("webaclid",
							cty.StringVal(webAclId))
					}
				}
				isError := false
				if doesReplicationControllerHaveAlbOrNlb(&service) {
					details, err := getDuploServiceAwsLbSettings(config.TenantId, &service, client)
					if details == nil || err != nil {
						isError = true
					}
					settings, err := client.TenantGetApplicationLbSettings(config.TenantId, details.LoadBalancerArn)
					if err != nil {
						isError = true
					}
					if settings != nil && settings.LoadBalancerArn != "" {
						svcParamBody.SetAttributeValue("enable_access_logs",
							cty.BoolVal(settings.EnableAccessLogs))
						svcParamBody.SetAttributeValue("drop_invalid_headers",
							cty.BoolVal(settings.DropInvalidHeaders))
						svcParamBody.SetAttributeValue("http_to_https_redirect",
							cty.BoolVal(settings.HttpToHttpsRedirect))
					} else if isError {
						svcParamBody.SetAttributeValue("enable_access_logs",
							cty.BoolVal(false))
						svcParamBody.SetAttributeValue("drop_invalid_headers",
							cty.BoolVal(false))
						svcParamBody.SetAttributeValue("http_to_https_redirect",
							cty.BoolVal(false))
					}
				}
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
					ResourceAddress: "duplocloud_duplo_service." + service.Name,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/ReplicationControllerApiV2/" + service.Name,
					WorkingDir:      workingDir,
				})

				if configPresent {
					importConfigs = append(importConfigs, common.ImportConfig{
						ResourceAddress: "duplocloud_duplo_service_lbconfigs." + service.Name + "-config",
						ResourceId:      "v2/subscriptions/" + config.TenantId + "/ServiceLBConfigsV2/" + service.Name,
						WorkingDir:      workingDir,
					})
					importConfigs = append(importConfigs, common.ImportConfig{
						ResourceAddress: "duplocloud_duplo_service_params." + service.Name + "-params",
						ResourceId:      "v2/subscriptions/" + config.TenantId + "/ReplicationControllerParamsV2/" + service.Name,
						WorkingDir:      workingDir,
					})

				}
				tfContext.ImportConfigs = importConfigs
			}
		}

	}

	log.Println("[TRACE] <====== Duplo Services TF generation done. =====>")
	return &tfContext, nil
}

func generateSvcVars(duplo duplosdk.DuploReplicationController, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	imageIdVar := common.VarConfig{
		Name:       prefix + "docker_image",
		DefaultVal: (*duplo.Template.Containers)[0].Image,
		TypeVal:    "string",
	}
	varConfigs["docker_image"] = imageIdVar

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}
	return vars
}

func doesReplicationControllerHaveAlb(duplo *duplosdk.DuploReplicationController) bool {
	if duplo != nil && duplo.Template != nil {
		for _, lb := range duplo.Template.LBConfigurations {
			if lb.LbType == 1 || lb.LbType == 2 { // ALB or Healthcheck only
				return true
			}
		}
	}
	return false
}

func getDuploServiceAwsLbSettings(tenantID string, rpc *duplosdk.DuploReplicationController, c *duplosdk.Client) (*duplosdk.DuploAwsLbDetailsInService, error) {

	if rpc.Template != nil && rpc.Template.Cloud == 0 {

		// Look for load balancer settings.
		details, err := c.TenantGetLbDetailsInService(tenantID, rpc.Name)
		if err != nil {
			return nil, err
		}
		if details != nil && details.LoadBalancerArn != "" {
			return details, nil
		}
	}

	// Nothing found.
	return nil, nil
}

func doesReplicationControllerHaveAlbOrNlb(duplo *duplosdk.DuploReplicationController) bool {
	if duplo != nil && duplo.Template != nil {
		for _, lb := range duplo.Template.LBConfigurations {
			if lb.LbType == 1 || lb.LbType == 2 || lb.LbType == 6 { // ALB, Healthcheck only or NLB
				return true
			}
		}
	}
	return false
}
