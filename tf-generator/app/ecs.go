package app

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
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type ECS struct {
}

func (ecs *ECS) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AppProject)

	taskDefnList, clientErr := client.EcsTaskDefinitionArnssGet(config.TenantId)
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	list, clientErr := client.EcsServiceList(config.TenantId)

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	taskDefn := []string{}
	if list != nil {
		log.Println("[TRACE] <====== Duplo ECS TF generation started. =====>")
		for _, ecs := range *list {

			taskDefObj, clientErr := client.EcsTaskDefinitionGet(config.TenantId, ecs.TaskDefinition)
			if clientErr != nil {
				fmt.Println(clientErr)
				return nil, clientErr
			}
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "ecs-"+ecs.Name+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			resourceName := common.GetResourceName(ecs.Name)
			// initialize the body of the new file object
			rootBody := hclFile.Body()
			log.Printf("[TRACE] Generating terraform config for duplo task definition : %s", taskDefObj.Family)
			// Add duplocloud_aws_host resource
			tdBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_ecs_task_definition",
					resourceName})
			tdBody := tdBlock.Body()
			tdBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			// tdBody.SetAttributeValue("tenant_id",
			// 	cty.StringVal(config.TenantId))
			taskDefnName, err := extractTaskDefnName(client, config.TenantId, taskDefObj.Family)
			taskDefn = append(taskDefn, taskDefnName)
			if err != nil {
				return nil, err
			}
			name := "duploservices-${local.tenant_name}-" + taskDefnName
			tdNameTokens := hclwrite.Tokens{
				{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenIdent, Bytes: []byte(name)},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
			}
			tdBody.SetAttributeRaw("family", tdNameTokens)

			// tdBody.SetAttributeValue("family",
			// 	cty.StringVal(taskDefObj.Family))
			tdBody.SetAttributeValue("cpu",
				cty.StringVal(taskDefObj.CPU))
			tdBody.SetAttributeValue("memory",
				cty.StringVal(taskDefObj.Memory))
			tdBody.SetAttributeValue("network_mode",
				cty.StringVal(taskDefObj.NetworkMode.Value))
			tdBody.SetAttributeValue("prevent_tf_destroy",
				cty.BoolVal(false))

			if taskDefObj.RequiresCompatibilities != nil && len(taskDefObj.RequiresCompatibilities) > 0 {
				var vals []cty.Value
				for _, s := range taskDefObj.RequiresCompatibilities {
					vals = append(vals, cty.StringVal(s))
				}
				tdBody.SetAttributeValue("requires_compatibilities",
					cty.ListVal(vals))
			}

			if taskDefObj.Volumes != nil && len(taskDefObj.Volumes) > 0 {
				volString, err := duplosdk.JSONMarshal(taskDefObj.Volumes)
				if err != nil {
					panic(err)
				}
				tdBody.SetAttributeTraversal("volumes", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + volString + ")",
					},
				})
			}
			if taskDefObj.ContainerDefinitions != nil && len(taskDefObj.ContainerDefinitions) > 0 {
				containerString, err := duplosdk.JSONMarshal(taskDefObj.ContainerDefinitions)
				if err != nil {
					panic(err)
				}
				containerString = strings.Replace(containerString, config.TenantName, "${local.tenant_name}", -1)
				tdBody.SetAttributeTraversal("container_definitions", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + containerString + ")",
					},
				})
			}
			rootBody.AppendNewline()
			log.Printf("[TRACE] Terraform config generated for duplo task definition : %s", taskDefObj.Family)

			log.Printf("[TRACE] Generating terraform config for duplo ECS service : %s", ecs.Name)

			ecsBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_ecs_service",
					resourceName})
			ecsBody := ecsBlock.Body()
			ecsBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			ecsBody.SetAttributeValue("name",
				cty.StringVal(ecs.Name))
			ecsBody.SetAttributeTraversal("task_definition", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "duplocloud_ecs_task_definition." + resourceName,
				},
				hcl.TraverseAttr{
					Name: "arn",
				},
			})
			ecsBody.SetAttributeValue("replicas",
				cty.NumberIntVal(int64(ecs.Replicas)))
			if ecs.HealthCheckGracePeriodSeconds > 0 {
				ecsBody.SetAttributeValue("health_check_grace_period_seconds",
					cty.NumberIntVal(int64(ecs.HealthCheckGracePeriodSeconds)))
			}
			ecsBody.SetAttributeValue("old_task_definition_buffer_size",
				cty.NumberIntVal(int64(ecs.OldTaskDefinitionBufferSize)))
			ecsBody.SetAttributeValue("is_target_group_only",
				cty.BoolVal(ecs.IsTargetGroupOnly))

			if len(ecs.DNSPrfx) > 0 {
				dnsPrefix := strings.Replace(ecs.DNSPrfx, "-"+config.TenantName, "", -1)
				dnsPrefix = dnsPrefix + "-${local.tenant_name}"
				dnsPrefixTokens := hclwrite.Tokens{
					{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
					{Type: hclsyntax.TokenIdent, Bytes: []byte(dnsPrefix)},
					{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
				}
				ecsBody.SetAttributeRaw("dns_prfx", dnsPrefixTokens)
			}

			for _, capacityProvider := range *ecs.CapacityProviderStrategy {
				cpConfigBlock := ecsBody.AppendNewBlock("capacity_provider_strategy",
					nil)
				cpConfigBlockBody := cpConfigBlock.Body()
				cpConfigBlockBody.SetAttributeValue("base",
					cty.NumberIntVal(int64(capacityProvider.Base)))
				cpConfigBlockBody.SetAttributeValue("weight",
					cty.NumberIntVal(int64(capacityProvider.Weight)))
				cpConfigBlockBody.SetAttributeValue("capacity_provider",
					cty.StringVal(capacityProvider.CapacityProvider))
			}
			for _, serviceConfig := range *ecs.LBConfigurations {
				lbConfigBlock := ecsBody.AppendNewBlock("load_balancer",
					nil)
				lbConfigBlockBody := lbConfigBlock.Body()

				lbConfigBlockBody.SetAttributeValue("target_group_count",
					cty.NumberIntVal(int64(serviceConfig.TgCount)))
				lbConfigBlockBody.SetAttributeValue("lb_type",
					cty.NumberIntVal(int64(serviceConfig.LbType)))
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
				if len(serviceConfig.BackendProtocol) > 0 {
					lbConfigBlockBody.SetAttributeValue("backend_protocol",
						cty.StringVal(serviceConfig.BackendProtocol))
				}
				if len(serviceConfig.HealthCheckURL) > 0 {
					lbConfigBlockBody.SetAttributeValue("health_check_url",
						cty.StringVal(serviceConfig.HealthCheckURL))
				}
				if len(serviceConfig.CertificateArn) > 0 {
					lbConfigBlockBody.SetAttributeTraversal("certificate_arn", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "local",
						},
						hcl.TraverseAttr{
							Name: "cert_arn",
						},
					})
				}

				// TODO - Add health_check_config block
				if serviceConfig.HealthCheckConfig != nil && (serviceConfig.HealthCheckConfig.HealthyThresholdCount != 0 || serviceConfig.HealthCheckConfig.UnhealthyThresholdCount != 0 || serviceConfig.HealthCheckConfig.HealthCheckIntervalSeconds != 0 || serviceConfig.HealthCheckConfig.HealthCheckTimeoutSeconds != 0) {
					lbConfigBlockBody.AppendNewline()
					hccBlock := lbConfigBlockBody.AppendNewBlock("health_check_config",
						nil)
					hccBlockBody := hccBlock.Body()
					hccBlockBody.SetAttributeValue("healthy_threshold_count",
						cty.NumberIntVal(int64(serviceConfig.HealthCheckConfig.HealthyThresholdCount)))
					hccBlockBody.SetAttributeValue("unhealthy_threshold_count",
						cty.NumberIntVal(int64(serviceConfig.HealthCheckConfig.UnhealthyThresholdCount)))
					hccBlockBody.SetAttributeValue("health_check_interval_seconds",
						cty.NumberIntVal(int64(serviceConfig.HealthCheckConfig.HealthCheckIntervalSeconds)))
					hccBlockBody.SetAttributeValue("health_check_timeout_seconds",
						cty.NumberIntVal(int64(serviceConfig.HealthCheckConfig.HealthCheckTimeoutSeconds)))
					if len(serviceConfig.HealthCheckConfig.HttpSuccessCode) > 0 {
						hccBlockBody.SetAttributeValue("http_success_code",
							cty.StringVal(serviceConfig.HealthCheckConfig.HttpSuccessCode))
					}
					if len(serviceConfig.HealthCheckConfig.GrpcSuccessCode) > 0 {
						hccBlockBody.SetAttributeValue("grpc_success_code",
							cty.StringVal(serviceConfig.HealthCheckConfig.GrpcSuccessCode))
					}
				}

				ecsBody.AppendNewline()
			}
			//}

			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_ecs_task_definition." + resourceName,
					ResourceId:      "subscriptions/" + config.TenantId + "/EcsTaskDefinition/" + ecs.TaskDefinition,
					WorkingDir:      workingDir,
				}, common.ImportConfig{
					ResourceAddress: "duplocloud_ecs_service." + resourceName,
					ResourceId:      "v2/subscriptions/" + config.TenantId + "/EcsServiceApiV2/" + ecs.Name,
					WorkingDir:      workingDir,
				},
				)
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== Duplo ECS TF generation done. =====>")
	}
	if taskDefnList != nil {
		for _, td := range *taskDefnList {
			tdObj, clientErr := client.EcsTaskDefinitionGet(config.TenantId, td)
			if clientErr != nil {
				fmt.Println(clientErr)
				return nil, nil
			}
			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			shortName, err := extractTaskDefnName(client, config.TenantId, tdObj.Family)
			if err != nil {
				return nil, err
			}
			if common.Contains(taskDefn, shortName) {
				continue
			}
			path := filepath.Join(workingDir, "td-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			resourceName := common.GetResourceName(shortName)
			rootBody := hclFile.Body()
			log.Printf("[TRACE] Generating terraform config for duplo task definition : %s", tdObj.Family)
			// Add duplocloud_aws_host resource
			tdBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_ecs_task_definition",
					resourceName})
			tdBody := tdBlock.Body()
			tdBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			name := "duploservices-${local.tenant_name}-" + shortName
			tdNameTokens := hclwrite.Tokens{
				{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
				{Type: hclsyntax.TokenIdent, Bytes: []byte(name)},
				{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
			}
			tdBody.SetAttributeRaw("family", tdNameTokens)

			// tdBody.SetAttributeValue("family",
			// 	cty.StringVal(taskDefObj.Family))
			tdBody.SetAttributeValue("cpu",
				cty.StringVal(tdObj.CPU))
			tdBody.SetAttributeValue("memory",
				cty.StringVal(tdObj.Memory))
			tdBody.SetAttributeValue("network_mode",
				cty.StringVal(tdObj.NetworkMode.Value))
			tdBody.SetAttributeValue("prevent_tf_destroy",
				cty.BoolVal(false))

			if tdObj.RequiresCompatibilities != nil && len(tdObj.RequiresCompatibilities) > 0 {
				var vals []cty.Value
				for _, s := range tdObj.RequiresCompatibilities {
					vals = append(vals, cty.StringVal(s))
				}
				tdBody.SetAttributeValue("requires_compatibilities",
					cty.ListVal(vals))
			}
			if tdObj.Volumes != nil && len(tdObj.Volumes) > 0 {
				volString, err := duplosdk.JSONMarshal(tdObj.Volumes)
				if err != nil {
					panic(err)
				}
				tdBody.SetAttributeTraversal("volumes", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + volString + ")",
					},
				})
			}
			if tdObj.ContainerDefinitions != nil && len(tdObj.ContainerDefinitions) > 0 {
				containerString, err := duplosdk.JSONMarshal(tdObj.ContainerDefinitions)
				if err != nil {
					panic(err)
				}
				containerString = strings.Replace(containerString, config.TenantName, "${local.tenant_name}", -1)
				tdBody.SetAttributeTraversal("container_definitions", hcl.Traversal{
					hcl.TraverseRoot{
						Name: "jsonencode(" + containerString + ")",
					},
				})
			}

			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_ecs_task_definition." + resourceName,
					ResourceId:      "subscriptions/" + config.TenantId + "/EcsTaskDefinition/" + td,
					WorkingDir:      workingDir,
				},
				)
				tfContext.ImportConfigs = importConfigs
			}
		}
	}

	return &tfContext, nil
}

func extractTaskDefnName(client *duplosdk.Client, tenantID string, family string) (string, error) {
	prefix, err := client.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return "", err
	}
	name, _ := duplosdk.UnprefixName(prefix, family)
	return name, nil
}
