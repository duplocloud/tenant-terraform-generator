package awsservices

import (
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

const LB_VAR_PREFIX = "lb_"

type LoadBalancer struct {
}

func (lb *LoadBalancer) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.TenantGetApplicationLBList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		log.Println("[TRACE] <====== Load balancer TF generation started. =====>")
		for _, lb := range *list {
			shortName, err := extractLbShortName(client, config.TenantId, lb.Name)
			resourceName := common.GetResourceName(shortName)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			settings, err := client.TenantGetApplicationLbSettings(config.TenantId, lb.Arn)
			if err != nil {
				fmt.Println(err)
				settings = nil
			}
			log.Printf("[TRACE] Generating terraform config for duplo aws load balancer : %s", shortName)

			varFullPrefix := LB_VAR_PREFIX + resourceName + "_"

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "lb-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			rootBody := hclFile.Body()

			lbBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_load_balancer",
					resourceName})
			lbBody := lbBlock.Body()
			lbBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			lbBody.SetAttributeValue("name",
				cty.StringVal(shortName))
			lBType := "Application"
			if len(lb.LbType.Value) > 0 {
				lBType = lb.LbType.Value
			}
			lbBody.SetAttributeValue("load_balancer_type",
				cty.StringVal(lBType))

			lbBody.SetAttributeValue("enable_access_logs",
				cty.BoolVal(lb.EnableAccessLogs))
			lbBody.SetAttributeValue("is_internal",
				cty.BoolVal(lb.IsInternal))

			if lb.LbType != nil {
				lbBody.SetAttributeValue("load_balancer_type",
					cty.StringVal(lb.LbType.Value))
			}

			if settings != nil {
				lbBody.SetAttributeValue("drop_invalid_headers",
					cty.BoolVal(settings.DropInvalidHeaders))
				if len(settings.WebACLID) > 0 {
					lbBody.SetAttributeValue("web_acl_id",
						cty.StringVal(settings.WebACLID))
				}
			}

			// Fetch all listeners
			listeners, clientErr := client.TenantListApplicationLbListeners(config.TenantId, shortName)
			if clientErr != nil {
				fmt.Println(err)
				listeners = nil
			}
			rootBody.AppendNewline()
			log.Printf("[TRACE] Terraform config is generation started for duplo aws load balancer listener : %s", shortName)
			if listeners != nil {
				for _, listener := range *listeners {
					listenerResourceName := resourceName + "_listener_" + strconv.Itoa(listener.Port)
					listenerBlock := rootBody.AppendNewBlock("resource",
						[]string{"duplocloud_aws_load_balancer_listener",
							listenerResourceName})
					listenerBody := listenerBlock.Body()

					listenerBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "local",
						},
						hcl.TraverseAttr{
							Name: "tenant_id",
						},
					})
					listenerBody.SetAttributeTraversal("load_balancer_name", hcl.Traversal{
						hcl.TraverseRoot{
							Name: "duplocloud_aws_load_balancer",
						},
						hcl.TraverseAttr{
							Name: resourceName + ".name",
						},
					})

					listenerBody.SetAttributeValue("protocol",
						cty.StringVal(listener.Protocol.Value))
					listenerBody.SetAttributeValue("port",
						cty.NumberIntVal(int64(listener.Port)))

					if len(listener.DefaultActions) > 0 {
						listenerBody.SetAttributeValue("target_group_arn",
							cty.StringVal(listener.DefaultActions[0].TargetGroupArn))
					}
					rootBody.AppendNewline()

					importConfigs = append(importConfigs, common.ImportConfig{
						ResourceAddress: "duplocloud_aws_load_balancer_listener." + listenerResourceName,
						ResourceId:      config.TenantId + "/" + shortName + "/" + listener.ListenerArn,
						WorkingDir:      workingDir,
					})

					getReq := duplosdk.DuploTargetGroupAttributesGetReq{
						TargetGroupArn: listener.DefaultActions[0].TargetGroupArn,
					}
					targetGrpAttrs, _ := client.DuploAwsTargetGroupAttributesGet(config.TenantId, getReq)
					if targetGrpAttrs != nil && len(*targetGrpAttrs) > 0 {
						tgAttrBlock := rootBody.AppendNewBlock("resource",
							[]string{"duplocloud_aws_target_group_attributes",
								resourceName + "_listener_" + strconv.Itoa(listener.Port) + "_tg_attributes"})
						tgAttrBody := tgAttrBlock.Body()
						tgAttrBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
							hcl.TraverseRoot{
								Name: "local",
							},
							hcl.TraverseAttr{
								Name: "tenant_id",
							},
						})
						tgAttrBody.SetAttributeTraversal("target_group_arn", hcl.Traversal{
							hcl.TraverseRoot{
								Name: "duplocloud_aws_load_balancer_listener." + resourceName + "_listener_" + strconv.Itoa(listener.Port),
							},
							hcl.TraverseAttr{
								Name: "target_group_arn",
							},
						})
						for _, tgAttr := range *targetGrpAttrs {
							if len(tgAttr.Key) > 0 && len(tgAttr.Value) > 0 {
								attrBlock := tgAttrBody.AppendNewBlock("dimension",
									nil)
								attrBody := attrBlock.Body()
								attrBody.SetAttributeValue("key", cty.StringVal(tgAttr.Key))
								attrBody.SetAttributeValue("value", cty.StringVal(tgAttr.Value))
							}
						}
						importConfigs = append(importConfigs, common.ImportConfig{
							ResourceAddress: "duplocloud_aws_target_group_attributes." + resourceName + "_listener_" + strconv.Itoa(listener.Port) + "_tg_attributes",
							ResourceId:      config.TenantId + "/" + listener.DefaultActions[0].TargetGroupArn,
							WorkingDir:      workingDir,
						})
					}

					// Add all listener Rules
					appendListenerRuleResources(listener.ListenerArn, listenerResourceName, rootBody, config, client, importConfigs, workingDir)
				}
			}

			log.Printf("[TRACE] Terraform config is generated for duplo aws load balancer listener.: %s", shortName)
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo aws load balancer : %s", shortName)

			outVars := generateLBOutputVars(varFullPrefix, resourceName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_load_balancer." + resourceName,
					ResourceId:      config.TenantId + "/" + shortName,
					WorkingDir:      workingDir,
				})
			}
		}
		tfContext.ImportConfigs = importConfigs
		log.Println("[TRACE] <====== Load balancer TF generation done. =====>")
	}

	return &tfContext, nil
}

func extractLbShortName(client *duplosdk.Client, tenantID string, fullName string) (string, error) {
	prefix, err := client.GetResourcePrefix("duplo3", tenantID)
	if err != nil {
		return "", err
	}
	name, _ := duplosdk.UnprefixName(prefix, fullName)
	return name, nil
}

func generateLBOutputVars(prefix, resourceName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "fullname",
		ActualVal:     "duplocloud_aws_load_balancer." + resourceName + ".fullname",
		DescVal:       "The full name of the load balancer.",
		RootTraversal: true,
	}
	outVarConfigs["fullname"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_load_balancer." + resourceName + ".arn",
		DescVal:       "The ARN of the load balancer.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "dns_name",
		ActualVal:     "duplocloud_aws_load_balancer." + resourceName + ".dns_name",
		DescVal:       "The DNS name of the load balancer.",
		RootTraversal: true,
	}
	outVarConfigs["dns_name"] = var3

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}

func appendListenerRuleResources(listenerArn string, listenerResourceName string, body *hclwrite.Body, config *common.Config, client *duplosdk.Client, importConfigs []common.ImportConfig, workingDir string) {
	listenerRules, clientErr := client.DuploAwsLbListenerRuleList(config.TenantId, listenerArn)
	if clientErr != nil {
		fmt.Println(clientErr)
		listenerRules = nil
	}
	if listenerRules != nil {
		for i, listenerRule := range *listenerRules {
			listenerRuleResourceName := listenerResourceName + "_rule_" + strconv.Itoa(i+1)
			listenerRuleBlock := body.AppendNewBlock("resource",
				[]string{"duplocloud_aws_lb_listener_rule", listenerRuleResourceName})
			listenerRuleBody := listenerRuleBlock.Body()

			listenerRuleBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})
			listenerRuleBody.SetAttributeTraversal("listener_arn", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "duplocloud_aws_load_balancer_listener",
				},
				hcl.TraverseAttr{
					Name: listenerResourceName + ".name",
				},
			})
			priority, _ := strconv.Atoi(listenerRule.Priority)
			if priority > 0 {
				listenerRuleBody.SetAttributeValue("priority",
					cty.NumberIntVal(int64(priority)))
			}

			if listenerRule.Actions != nil && len(*listenerRule.Actions) > 0 {
				for _, action := range *listenerRule.Actions {
					actionBlock := listenerRuleBody.AppendNewBlock("action", nil)
					actionBody := actionBlock.Body()
					actionBody.SetAttributeValue("type", cty.StringVal(action.Type.Value))
					if action.Order > 0 {
						actionBody.SetAttributeValue("order",
							cty.NumberIntVal(int64(action.Order)))
					}

					switch action.Type.Value {
					case "forward":
						if action.Type.Value == "forward" && len(action.TargetGroupArn) > 0 {
							actionBody.SetAttributeValue("target_group_arn", cty.StringVal(action.TargetGroupArn))
						} else if action.ForwardConfig != nil {
							forwardBlock := actionBody.AppendNewBlock("forward", nil)
							forwardBody := forwardBlock.Body()
							if action.ForwardConfig.TargetGroups != nil && len(*action.ForwardConfig.TargetGroups) > 0 {
								for _, tg := range *action.ForwardConfig.TargetGroups {
									tgBlock := forwardBody.AppendNewBlock("target_group", nil)
									tgBody := tgBlock.Body()
									tgBody.SetAttributeValue("arn", cty.StringVal(tg.TargetGroupArn))
									tgBody.SetAttributeValue("weight",
										cty.NumberIntVal(int64(tg.Weight)))
								}
							}
							if action.ForwardConfig.TargetGroupStickinessConfig != nil {
								tgscBlock := forwardBody.AppendNewBlock("stickiness", nil)
								tgscBody := tgscBlock.Body()
								tgscBody.SetAttributeValue("enabled", cty.BoolVal(action.ForwardConfig.TargetGroupStickinessConfig.Enabled))
								tgscBody.SetAttributeValue("duration",
									cty.NumberIntVal(int64(action.ForwardConfig.TargetGroupStickinessConfig.DurationSeconds)))
							}
						}
					case "redirect":
						if action.RedirectConfig != nil {
							redirectBlock := actionBody.AppendNewBlock("redirect", nil)
							redirectBody := redirectBlock.Body()
							redirectBody.SetAttributeValue("status_code", cty.StringVal(action.RedirectConfig.StatusCode.Value))
							if len(action.RedirectConfig.Host) > 0 {
								redirectBody.SetAttributeValue("host", cty.StringVal(action.RedirectConfig.Host))
							}
							if len(action.RedirectConfig.Path) > 0 {
								redirectBody.SetAttributeValue("path", cty.StringVal(action.RedirectConfig.Path))
							}
							if len(action.RedirectConfig.Port) > 0 {
								redirectBody.SetAttributeValue("port", cty.StringVal(action.RedirectConfig.Port))
							}
							if len(action.RedirectConfig.Protocol) > 0 {
								redirectBody.SetAttributeValue("protocol", cty.StringVal(action.RedirectConfig.Protocol))
							}
							if len(action.RedirectConfig.Query) > 0 {
								redirectBody.SetAttributeValue("query", cty.StringVal(action.RedirectConfig.Query))
							}
						}
					case "fixed-response":
						if action.FixedResponseConfig != nil {
							frcBlock := actionBody.AppendNewBlock("fixed_response", nil)
							frcBody := frcBlock.Body()
							frcBody.SetAttributeValue("content_type", cty.StringVal(action.FixedResponseConfig.ContentType))
							if len(action.FixedResponseConfig.MessageBody) > 0 {
								frcBody.SetAttributeValue("message_body", cty.StringVal(action.FixedResponseConfig.MessageBody))
							}
							if len(action.FixedResponseConfig.StatusCode) > 0 {
								frcBody.SetAttributeValue("status_code", cty.StringVal(action.FixedResponseConfig.StatusCode))
							}
						}
					case "authenticate-cognito":
						if action.AuthenticateCognitoConfig != nil {
							acBlock := actionBody.AppendNewBlock("authenticate_cognito", nil)
							acBody := acBlock.Body()
							acBody.SetAttributeValue("user_pool_arn", cty.StringVal(action.AuthenticateCognitoConfig.UserPoolArn))
							acBody.SetAttributeValue("user_pool_client_id", cty.StringVal(action.AuthenticateCognitoConfig.UserPoolClientId))
							acBody.SetAttributeValue("user_pool_domain", cty.StringVal(action.AuthenticateCognitoConfig.UserPoolDomain))

							if len(action.AuthenticateCognitoConfig.AuthenticationRequestExtraParams) > 0 {
								newMap := make(map[string]cty.Value)
								for key, element := range action.AuthenticateCognitoConfig.AuthenticationRequestExtraParams {
									newMap[key] = cty.StringVal(element)
								}
								acBody.SetAttributeValue("authentication_request_extra_params", cty.MapVal(newMap))
							}
							if action.AuthenticateCognitoConfig.OnUnauthenticatedRequest != nil && len(action.AuthenticateCognitoConfig.OnUnauthenticatedRequest.Value) > 0 {
								acBody.SetAttributeValue("on_unauthenticated_request", cty.StringVal(action.AuthenticateCognitoConfig.OnUnauthenticatedRequest.Value))
							}
							if len(action.AuthenticateCognitoConfig.Scope) > 0 {
								acBody.SetAttributeValue("scope", cty.StringVal(action.AuthenticateCognitoConfig.Scope))
							}
							if len(action.AuthenticateCognitoConfig.SessionCookieName) > 0 {
								acBody.SetAttributeValue("session_cookie_name", cty.StringVal(action.AuthenticateCognitoConfig.SessionCookieName))
							}
							if action.AuthenticateCognitoConfig.SessionTimeout > 0 {
								acBody.SetAttributeValue("session_timeout", cty.NumberIntVal(int64(action.AuthenticateCognitoConfig.SessionTimeout)))
							}
						}
					case "authenticate-oidc":
						if action.AuthenticateOidcConfig != nil {
							oidcBlock := actionBody.AppendNewBlock("authenticate_oidc", nil)
							oidcBody := oidcBlock.Body()
							oidcBody.SetAttributeValue("authorization_endpoint", cty.StringVal(action.AuthenticateOidcConfig.AuthorizationEndpoint))
							oidcBody.SetAttributeValue("client_id", cty.StringVal(action.AuthenticateOidcConfig.ClientId))
							oidcBody.SetAttributeValue("client_secret", cty.StringVal(action.AuthenticateOidcConfig.ClientSecret))
							oidcBody.SetAttributeValue("issuer", cty.StringVal(action.AuthenticateOidcConfig.Issuer))
							oidcBody.SetAttributeValue("token_endpoint", cty.StringVal(action.AuthenticateOidcConfig.TokenEndpoint))
							oidcBody.SetAttributeValue("user_info_endpoint", cty.StringVal(action.AuthenticateOidcConfig.UserInfoEndpoint))
							if len(action.AuthenticateOidcConfig.AuthenticationRequestExtraParams) > 0 {
								newMap := make(map[string]cty.Value)
								for key, element := range action.AuthenticateOidcConfig.AuthenticationRequestExtraParams {
									newMap[key] = cty.StringVal(element)
								}
								oidcBody.SetAttributeValue("authentication_request_extra_params", cty.MapVal(newMap))
							}
							if action.AuthenticateOidcConfig.OnUnauthenticatedRequest != nil && len(action.AuthenticateOidcConfig.OnUnauthenticatedRequest.Value) > 0 {
								oidcBody.SetAttributeValue("on_unauthenticated_request", cty.StringVal(action.AuthenticateOidcConfig.OnUnauthenticatedRequest.Value))
							}
							if len(action.AuthenticateOidcConfig.Scope) > 0 {
								oidcBody.SetAttributeValue("scope", cty.StringVal(action.AuthenticateOidcConfig.Scope))
							}
							if len(action.AuthenticateOidcConfig.SessionCookieName) > 0 {
								oidcBody.SetAttributeValue("session_cookie_name", cty.StringVal(action.AuthenticateOidcConfig.SessionCookieName))
							}
							if action.AuthenticateOidcConfig.SessionTimeout > 0 {
								oidcBody.SetAttributeValue("session_timeout", cty.NumberIntVal(int64(action.AuthenticateOidcConfig.SessionTimeout)))
							}
						}
					}
				}
			}
			if listenerRule.Conditions != nil && len(*listenerRule.Conditions) > 0 {
				for _, condition := range *listenerRule.Conditions {
					conditionBlock := listenerRuleBody.AppendNewBlock("condition", nil)
					conditionBody := conditionBlock.Body()
					switch condition.Field {
					case "host-header":
						if condition.HostHeaderConfig != nil {
							hostHeaderBlock := conditionBody.AppendNewBlock("host_header", nil)
							hostHeaderBody := hostHeaderBlock.Body()
							var vals []cty.Value
							for _, s := range condition.HostHeaderConfig.Values {
								vals = append(vals, cty.StringVal(s))
							}
							hostHeaderBody.SetAttributeValue("values",
								cty.ListVal(vals))
						}
					case "http-header":
						if condition.HttpHeaderConfig != nil {
							httpHeaderBlock := conditionBody.AppendNewBlock("http_header", nil)
							httpHeaderBody := httpHeaderBlock.Body()
							var vals []cty.Value
							for _, s := range condition.HttpHeaderConfig.Values {
								vals = append(vals, cty.StringVal(s))
							}
							httpHeaderBody.SetAttributeValue("http_header_name",
								cty.StringVal(condition.HttpHeaderConfig.HttpHeaderName))
							httpHeaderBody.SetAttributeValue("values",
								cty.ListVal(vals))
						}
					case "http-request-method":
						if condition.HttpRequestMethodConfig != nil {
							httpReqHeaderBlock := conditionBody.AppendNewBlock("http_request_method", nil)
							httpReqHeaderBody := httpReqHeaderBlock.Body()
							var vals []cty.Value
							for _, s := range condition.HttpRequestMethodConfig.Values {
								vals = append(vals, cty.StringVal(s))
							}
							httpReqHeaderBody.SetAttributeValue("values",
								cty.ListVal(vals))
						}

					case "path-pattern":
						if condition.PathPatternConfig != nil {
							ppBlock := conditionBody.AppendNewBlock("path_pattern", nil)
							ppBody := ppBlock.Body()
							var vals []cty.Value
							for _, s := range condition.PathPatternConfig.Values {
								vals = append(vals, cty.StringVal(s))
							}
							ppBody.SetAttributeValue("values",
								cty.ListVal(vals))
						}

					case "query-string":
						if condition.QueryStringConfig != nil {
							for _, s := range *condition.QueryStringConfig.Values {
								qsBlock := conditionBody.AppendNewBlock("query_string", nil)
								qsBody := qsBlock.Body()
								qsBody.SetAttributeValue("key",
									cty.StringVal(s.Key))
								qsBody.SetAttributeValue("value",
									cty.StringVal(s.Value))
							}

						}

					case "source-ip":
						if condition.SourceIpConfig != nil {
							sipBlock := conditionBody.AppendNewBlock("source_ip", nil)
							sipBody := sipBlock.Body()
							var vals []cty.Value
							for _, s := range condition.SourceIpConfig.Values {
								vals = append(vals, cty.StringVal(s))
							}
							sipBody.SetAttributeValue("values",
								cty.ListVal(vals))
						}

					}
				}
			}
			importConfigs = append(importConfigs, common.ImportConfig{
				ResourceAddress: "duplocloud_aws_lb_listener_rule." + listenerRuleResourceName,
				ResourceId:      config.TenantId + "/" + listenerRule.RuleArn,
				WorkingDir:      workingDir,
			})
		}
	}
}
