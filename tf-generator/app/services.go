package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type Services struct {
}

func (s *Services) Generate(config *common.Config, client *duplosdk.Client) {
	log.Println("[TRACE] <====== Duplo Services TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AppProject)
	list, clientErr := client.ReplicationControllerList(config.TenantId)

	if clientErr != nil {
		fmt.Println(clientErr)
		return
	}

	if list != nil {
		for _, service := range *list {
			log.Printf("[TRACE] Generating terraform config for duplo service : %s", service.Name)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "svc-"+service.Name+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()
			// Add duplocloud_aws_host resource
			svcBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_duplo_service",
					service.Name})
			svcBody := svcBlock.Body()
			// svcBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
			// 	hcl.TraverseRoot{
			// 		Name: "duplocloud_tenant.tenant",
			// 	},
			// 	hcl.TraverseAttr{
			// 		Name: "tenant_id",
			// 	},
			// })
			svcBody.SetAttributeValue("tenant_id",
				cty.StringVal(config.TenantId))
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
				svcBody.SetAttributeValue("allocation_tags",
					cty.StringVal(service.Template.AllocationTags))
				if len(service.Template.OtherDockerConfig) > 0 {
					OtherDockerConfigMap := make(map[string]interface{})
					err := json.Unmarshal([]byte(service.Template.OtherDockerConfig), &OtherDockerConfigMap)
					if err != nil {
						panic(err)
					}
					OtherDockerConfigStr, err := duplosdk.PrettyStruct(OtherDockerConfigMap)
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
					extraConfigMap := make(map[string]interface{})
					err := json.Unmarshal([]byte(service.Template.ExtraConfig), &extraConfigMap)
					if err != nil {
						panic(err)
					}
					extraConfigStr, err := duplosdk.PrettyStruct(extraConfigMap)
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
					OtherDockerHostConfigStr, err := duplosdk.PrettyStruct(OtherDockerHostConfigMap)
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
					svcBody.SetAttributeValue("docker_image",
						cty.StringVal((*service.Template.Containers)[0].Image))
				}
			}
			log.Printf("[TRACE] Terraform config is generated for duplo service : %s", service.Name)
			rootBody.AppendNewline()
			configList, clientErr := client.ReplicationControllerLbConfigurationList(config.TenantId, service.Name)
			if clientErr != nil {
				fmt.Println(clientErr)
				return
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
						return
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
			}

			tfFile.Write(hclFile.Bytes())

			// Import all created resources.
			if config.GenerateTfState {
				importer := &common.Importer{}
				importer.Import(config, &common.ImportConfig{
					ResourceAddress: "duplocloud_duplo_service." + service.Name,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/ReplicationControllerApiV2/" + service.Name,
					WorkingDir:      workingDir,
				})
				if configPresent {
					importer.Import(config, &common.ImportConfig{
						ResourceAddress: "duplocloud_duplo_service_lbconfigs." + service.Name + "-config",
						ResourceId:      "v2/subscriptions/" + config.TenantId + "/ServiceLBConfigsV2/" + service.Name,
						WorkingDir:      workingDir,
					})
				}
			}
		}

	}

	log.Println("[TRACE] <====== Duplo Services TF generation done. =====>")
}
