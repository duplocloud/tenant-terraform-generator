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

	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const CFD_VAR_PREFIX = "cfd_"

type CFD struct {
}

func (cfd *CFD) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	log.Println("[TRACE] <====== AWS Cloudfront Distribution TF generation started. =====>")
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.AwsCloudfrontDistributionList(config.TenantId)
	//Get tenant from duplo

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, nil
	}
	prefix, clientErr := client.GetDuploServicesPrefix(config.TenantId)
	if clientErr != nil {
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	if list != nil {
		s3List, _ := client.TenantListS3Buckets(config.TenantId)
		for _, cfd := range *list {
			shortName, _ := duplosdk.UnprefixName(prefix, cfd.Comment)
			log.Printf("[TRACE] Generating terraform config for duplo AWS Cloudfront Distribution : %s", shortName)

			varFullPrefix := CFD_VAR_PREFIX + strings.ReplaceAll(shortName, "-", "_") + "_"

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "cfd-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			// initialize the body of the new file object
			rootBody := hclFile.Body()

			// Add duplocloud_aws_cloudfront_distribution resource
			cfdBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_cloudfront_distribution",
					shortName})
			cfdBody := cfdBlock.Body()
			cfdBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			if len(cfd.Comment) > 0 {
				cfdBody.SetAttributeValue("comment", cty.StringVal(shortName))
			}

			if cfd.Aliases != nil && len(cfd.Aliases.Items) > 0 {
				var vals []cty.Value
				for _, s := range cfd.Aliases.Items {
					vals = append(vals, cty.StringVal(s))
				}
				cfdBody.SetAttributeValue("aliases", cty.SetVal(vals))
			}

			if len(cfd.DefaultRootObject) > 0 {
				cfdBody.SetAttributeValue("default_root_object", cty.StringVal(cfd.DefaultRootObject))
			}

			cfdBody.SetAttributeValue("enabled", cty.BoolVal(cfd.Enabled))

			if cfd.HttpVersion != nil && len(cfd.HttpVersion.Value) > 0 {
				cfdBody.SetAttributeValue("http_version", cty.StringVal(cfd.HttpVersion.Value))
			}

			if cfd.PriceClass != nil && len(cfd.PriceClass.Value) > 0 {
				cfdBody.SetAttributeValue("price_class", cty.StringVal(cfd.PriceClass.Value))
			}

			if cfd.IsIPV6Enabled {
				cfdBody.SetAttributeValue("is_ipv6_enabled", cty.BoolVal(cfd.IsIPV6Enabled))
			}

			if len(cfd.WebACLId) > 0 {
				cfdBody.SetAttributeValue("web_acl_id", cty.StringVal(cfd.WebACLId))
			}

			if cfd.CustomErrorResponses != nil && cfd.CustomErrorResponses.Quantity > 0 {
				for _, cer := range *cfd.CustomErrorResponses.Items {
					cerBlock := cfdBody.AppendNewBlock("custom_error_response",
						nil)
					cerBody := cerBlock.Body()
					cerBody.SetAttributeValue("error_code", cty.NumberIntVal(int64(cer.ErrorCode)))
					if len(cer.ResponseCode) > 0 {
						val, _ := strconv.Atoi(cer.ResponseCode)
						cerBody.SetAttributeValue("response_code", cty.NumberIntVal(int64(val)))
					}
					if len(cer.ResponsePagePath) > 0 {
						cerBody.SetAttributeValue("response_page_path", cty.StringVal(cer.ResponsePagePath))
					}
					if cer.ErrorCachingMinTTL > 0 {
						cerBody.SetAttributeValue("error_caching_min_ttl", cty.NumberIntVal(int64(cer.ErrorCachingMinTTL)))
					}
				}
			}
			if cfd.ViewerCertificate != nil {
				vcBlock := cfdBody.AppendNewBlock("viewer_certificate",
					nil)
				vcBody := vcBlock.Body()
				if len(cfd.ViewerCertificate.IAMCertificateId) > 0 {
					vcBody.SetAttributeValue("iam_certificate_id", cty.StringVal(cfd.ViewerCertificate.IAMCertificateId))
				} else if len(cfd.ViewerCertificate.ACMCertificateArn) > 0 {
					vcBody.SetAttributeValue("acm_certificate_arn", cty.StringVal(cfd.ViewerCertificate.ACMCertificateArn))
				} else {
					vcBody.SetAttributeValue("cloudfront_default_certificate", cty.BoolVal(cfd.ViewerCertificate.CloudFrontDefaultCertificate))
				}
				if cfd.ViewerCertificate.MinimumProtocolVersion != nil && len(cfd.ViewerCertificate.MinimumProtocolVersion.Value) > 0 {
					vcBody.SetAttributeValue("minimum_protocol_version", cty.StringVal(cfd.ViewerCertificate.MinimumProtocolVersion.Value))
				}
				if cfd.ViewerCertificate.SSLSupportMethod != nil && len(cfd.ViewerCertificate.SSLSupportMethod.Value) > 0 {
					vcBody.SetAttributeValue("ssl_support_method", cty.StringVal(cfd.ViewerCertificate.SSLSupportMethod.Value))
				}
			}
			if cfd.Restrictions != nil && cfd.Restrictions.GeoRestriction != nil && cfd.Restrictions.GeoRestriction.Quantity > 0 {
				resBlock := cfdBody.AppendNewBlock("restrictions", nil)
				resBody := resBlock.Body()
				gresBlock := resBody.AppendNewBlock("restrictions", nil)
				gresBody := gresBlock.Body()

				gresBody.SetAttributeValue("restriction_type", cty.StringVal(cfd.Restrictions.GeoRestriction.RestrictionType.Value))
			}
			if cfd.Logging != nil {
				logBlock := cfdBody.AppendNewBlock("logging_config", nil)
				logBody := logBlock.Body()
				logBody.SetAttributeValue("bucket", cty.StringVal(cfd.Logging.Bucket))
				if len(cfd.Logging.Prefix) > 0 {
					logBody.SetAttributeValue("prefix", cty.StringVal(cfd.Logging.Prefix))
				}
				if cfd.Logging.IncludeCookies {
					logBody.SetAttributeValue("prefix", cty.BoolVal(cfd.Logging.IncludeCookies))
				}
			}

			if cfd.OriginGroups != nil && cfd.OriginGroups.Quantity > 0 {
				for _, og := range *cfd.OriginGroups.Items {
					ogBlock := cfdBody.AppendNewBlock("origin_group", nil)
					ogBody := ogBlock.Body()
					ogBody.SetAttributeValue("origin_id", cty.StringVal(og.Id))
					if og.FailoverCriteria != nil && og.FailoverCriteria.StatusCodes != nil && og.FailoverCriteria.StatusCodes.Quantity > 0 {
						focBlock := ogBody.AppendNewBlock("failover_criteria", nil)
						focBody := focBlock.Body()
						var vals []cty.Value
						for _, s := range og.FailoverCriteria.StatusCodes.Items {
							vals = append(vals, cty.NumberIntVal(int64(s)))
						}
						focBody.SetAttributeValue("status_codes", cty.SetVal(vals))
					}
					if og.Members != nil && og.Members.Quantity > 0 {
						for _, member := range *og.Members.Items {
							memberBlock := ogBody.AppendNewBlock("member", nil)
							memberBody := memberBlock.Body()
							memberBody.SetAttributeValue("origin_id", cty.StringVal(member.OriginId))
						}
					}
				}
			}

			if cfd.Origins != nil && cfd.Origins.Quantity > 0 {
				for _, origin := range *cfd.Origins.Items {
					originBlock := cfdBody.AppendNewBlock("origin", nil)
					originBody := originBlock.Body()
					originBody.SetAttributeValue("connection_attempts", cty.NumberIntVal(int64(origin.ConnectionAttempts)))
					originBody.SetAttributeValue("connection_timeout", cty.NumberIntVal(int64(origin.ConnectionTimeout)))
					orginAdded := false
					for _, s3 := range *s3List {
						if strings.HasPrefix(origin.DomainName, s3.Name) {
							s3ShortName := s3.Name[len("duploservices-"+config.TenantName+"-"):len(s3.Name)]
							parts := strings.Split(s3ShortName, "-")
							if len(parts) > 0 {
								parts = parts[:len(parts)-1]
							}
							s3ShortName = strings.Join(parts, "-")
							str := "${duplocloud_s3_bucket." + s3ShortName + ".fullname}.s3.${local.region}.amazonaws.com"
							tokens := hclwrite.Tokens{
								{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
								{Type: hclsyntax.TokenIdent, Bytes: []byte(str)},
								{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
							}
							originBody.SetAttributeRaw("domain_name", tokens)
							originBody.SetAttributeRaw("origin_id", tokens)
							orginAdded = true
							break
						}
					}
					if !orginAdded {
						originBody.SetAttributeValue("domain_name", cty.StringVal(origin.DomainName))
						originBody.SetAttributeValue("origin_id", cty.StringVal(origin.Id))
					}

					if len(origin.OriginPath) > 0 {
						originBody.SetAttributeValue("origin_path", cty.StringVal(origin.OriginPath))
					}
					if origin.CustomOriginConfig != nil {
						cocBlock := originBody.AppendNewBlock("custom_origin_config", nil)
						cocBody := cocBlock.Body()
						cocBody.SetAttributeValue("http_port", cty.NumberIntVal(int64(origin.CustomOriginConfig.HTTPPort)))
						cocBody.SetAttributeValue("https_port", cty.NumberIntVal(int64(origin.CustomOriginConfig.HTTPSPort)))
						cocBody.SetAttributeValue("origin_keepalive_timeout", cty.NumberIntVal(int64(origin.CustomOriginConfig.OriginKeepaliveTimeout)))
						cocBody.SetAttributeValue("origin_read_timeout", cty.NumberIntVal(int64(origin.CustomOriginConfig.OriginReadTimeout)))
						cocBody.SetAttributeValue("origin_protocol_policy", cty.StringVal(origin.CustomOriginConfig.OriginProtocolPolicy.Value))
						var vals []cty.Value
						for _, s := range origin.CustomOriginConfig.OriginSslProtocols.Items {
							vals = append(vals, cty.StringVal(s))
						}
						cocBody.SetAttributeValue("origin_ssl_protocols", cty.SetVal(vals))
					}
					if origin.CustomHeaders != nil && origin.CustomHeaders.Quantity > 0 {
						for _, header := range *origin.CustomHeaders.Items {
							headerBlock := originBody.AppendNewBlock("custom_header", nil)
							headerBody := headerBlock.Body()
							headerBody.SetAttributeValue("name", cty.StringVal(header.HeaderName))
							headerBody.SetAttributeValue("value", cty.StringVal(header.HeaderValue))
						}
					}
					if origin.OriginShield != nil && origin.OriginShield.Enabled {
						originShieldBlock := originBody.AppendNewBlock("origin_shield", nil)
						originShieldBody := originShieldBlock.Body()
						originShieldBody.SetAttributeValue("enabled", cty.BoolVal(origin.OriginShield.Enabled))
						originShieldBody.SetAttributeValue("origin_shield_region", cty.StringVal(origin.OriginShield.OriginShieldRegion))

					}
					// s3_origin_config --> origin_access_identity duplo handles at backend.
				}
			}
			if cfd.DefaultCacheBehavior != nil {
				dcbBlock := cfdBody.AppendNewBlock("default_cache_behavior", nil)
				dcbBody := dcbBlock.Body()
				targetOrginAdded := false
				for _, s3 := range *s3List {
					if strings.HasPrefix(cfd.DefaultCacheBehavior.TargetOriginId, s3.Name) {
						s3ShortName := s3.Name[len("duploservices-"+config.TenantName+"-"):len(s3.Name)]
						parts := strings.Split(s3ShortName, "-")
						if len(parts) > 0 {
							parts = parts[:len(parts)-1]
						}
						s3ShortName = strings.Join(parts, "-")
						str := "${duplocloud_s3_bucket." + s3ShortName + ".fullname}.s3.${local.region}.amazonaws.com"
						tokens := hclwrite.Tokens{
							{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
							{Type: hclsyntax.TokenIdent, Bytes: []byte(str)},
							{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
						}
						dcbBody.SetAttributeRaw("target_origin_id", tokens)
						targetOrginAdded = true
						break
					}
				}
				if !targetOrginAdded {
					dcbBody.SetAttributeValue("target_origin_id", cty.StringVal(cfd.DefaultCacheBehavior.TargetOriginId))
				}

				var allowedMethods []cty.Value
				for _, s := range cfd.DefaultCacheBehavior.AllowedMethods.Items {
					allowedMethods = append(allowedMethods, cty.StringVal(s))
				}
				dcbBody.SetAttributeValue("allowed_methods", cty.SetVal(allowedMethods))
				var cachedMethods []cty.Value
				for _, s := range cfd.DefaultCacheBehavior.AllowedMethods.CachedMethods.Items {
					cachedMethods = append(cachedMethods, cty.StringVal(s))
				}
				dcbBody.SetAttributeValue("cached_methods", cty.SetVal(cachedMethods))
				if len(cfd.DefaultCacheBehavior.CachePolicyId) > 0 {
					dcbBody.SetAttributeValue("cache_policy_id", cty.StringVal(cfd.DefaultCacheBehavior.CachePolicyId))
				}
				if cfd.DefaultCacheBehavior.Compress {
					dcbBody.SetAttributeValue("compress", cty.BoolVal(cfd.DefaultCacheBehavior.Compress))
				}
				if cfd.DefaultCacheBehavior.DefaultTTL > 0 {
					dcbBody.SetAttributeValue("default_ttl", cty.NumberIntVal(int64(cfd.DefaultCacheBehavior.DefaultTTL)))
				}
				if len(cfd.DefaultCacheBehavior.FieldLevelEncryptionId) > 0 {
					dcbBody.SetAttributeValue("field_level_encryption_id", cty.StringVal(cfd.DefaultCacheBehavior.FieldLevelEncryptionId))
				}
				if cfd.DefaultCacheBehavior.MaxTTL > 0 {
					dcbBody.SetAttributeValue("max_ttl", cty.NumberIntVal(int64(cfd.DefaultCacheBehavior.MaxTTL)))
				}
				if cfd.DefaultCacheBehavior.MinTTL > 0 {
					dcbBody.SetAttributeValue("min_ttl", cty.NumberIntVal(int64(cfd.DefaultCacheBehavior.MinTTL)))
				}
				if len(cfd.DefaultCacheBehavior.OriginRequestPolicyId) > 0 {
					dcbBody.SetAttributeValue("origin_request_policy_id", cty.StringVal(cfd.DefaultCacheBehavior.OriginRequestPolicyId))
				}
				if cfd.DefaultCacheBehavior.SmoothStreaming {
					dcbBody.SetAttributeValue("smooth_streaming", cty.BoolVal(cfd.DefaultCacheBehavior.SmoothStreaming))
				}
				if cfd.DefaultCacheBehavior.TrustedSigners != nil && cfd.DefaultCacheBehavior.TrustedSigners.Quantity > 0 {
					var trustedSigners []cty.Value
					for _, s := range cfd.DefaultCacheBehavior.TrustedSigners.Items {
						trustedSigners = append(trustedSigners, cty.StringVal(s))
					}
					dcbBody.SetAttributeValue("trusted_signers", cty.ListVal(trustedSigners))
				}
				dcbBody.SetAttributeValue("viewer_protocol_policy", cty.StringVal(cfd.DefaultCacheBehavior.ViewerProtocolPolicy.Value))
			}
			if cfd.CacheBehaviors != nil && cfd.CacheBehaviors.Quantity > 0 {
				for _, ocb := range *cfd.CacheBehaviors.Items {
					targetOrginAdded := false
					ocbBlock := cfdBody.AppendNewBlock("ordered_cache_behavior", nil)
					ocbBody := ocbBlock.Body()
					for _, s3 := range *s3List {
						if strings.HasPrefix(ocb.TargetOriginId, s3.Name) {
							s3ShortName := s3.Name[len("duploservices-"+config.TenantName+"-"):len(s3.Name)]
							parts := strings.Split(s3ShortName, "-")
							if len(parts) > 0 {
								parts = parts[:len(parts)-1]
							}
							s3ShortName = strings.Join(parts, "-")
							str := "${duplocloud_s3_bucket." + s3ShortName + ".fullname}.s3.${local.region}.amazonaws.com"
							tokens := hclwrite.Tokens{
								{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
								{Type: hclsyntax.TokenIdent, Bytes: []byte(str)},
								{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
							}
							ocbBody.SetAttributeRaw("target_origin_id", tokens)
							targetOrginAdded = true
							break
						}
					}
					if !targetOrginAdded {
						ocbBody.SetAttributeValue("target_origin_id", cty.StringVal(ocb.TargetOriginId))
					}
					var allowedMethods []cty.Value
					for _, s := range ocb.AllowedMethods.Items {
						allowedMethods = append(allowedMethods, cty.StringVal(s))
					}
					ocbBody.SetAttributeValue("allowed_methods", cty.SetVal(allowedMethods))
					var cachedMethods []cty.Value
					for _, s := range ocb.AllowedMethods.CachedMethods.Items {
						cachedMethods = append(cachedMethods, cty.StringVal(s))
					}
					ocbBody.SetAttributeValue("cached_methods", cty.SetVal(cachedMethods))
					if len(ocb.CachePolicyId) > 0 {
						ocbBody.SetAttributeValue("cache_policy_id", cty.StringVal(ocb.CachePolicyId))
					}
					if ocb.Compress {
						ocbBody.SetAttributeValue("compress", cty.BoolVal(ocb.Compress))
					}
					if ocb.DefaultTTL > 0 {
						ocbBody.SetAttributeValue("default_ttl", cty.NumberIntVal(int64(ocb.DefaultTTL)))
					}
					if len(ocb.FieldLevelEncryptionId) > 0 {
						ocbBody.SetAttributeValue("field_level_encryption_id", cty.StringVal(ocb.FieldLevelEncryptionId))
					}
					if ocb.MaxTTL > 0 {
						ocbBody.SetAttributeValue("max_ttl", cty.NumberIntVal(int64(ocb.MaxTTL)))
					}
					if ocb.MinTTL > 0 {
						ocbBody.SetAttributeValue("min_ttl", cty.NumberIntVal(int64(ocb.MinTTL)))
					}
					if len(ocb.OriginRequestPolicyId) > 0 {
						ocbBody.SetAttributeValue("origin_request_policy_id", cty.StringVal(ocb.OriginRequestPolicyId))
					}
					if ocb.SmoothStreaming {
						ocbBody.SetAttributeValue("smooth_streaming", cty.BoolVal(ocb.SmoothStreaming))
					}
					if ocb.TrustedSigners != nil && ocb.TrustedSigners.Quantity > 0 {
						var trustedSigners []cty.Value
						for _, s := range ocb.TrustedSigners.Items {
							trustedSigners = append(trustedSigners, cty.StringVal(s))
						}
						ocbBody.SetAttributeValue("trusted_signers", cty.ListVal(trustedSigners))
					}

					ocbBody.SetAttributeValue("path_pattern", cty.StringVal(ocb.PathPattern))
					ocbBody.SetAttributeValue("viewer_protocol_policy", cty.StringVal(ocb.ViewerProtocolPolicy.Value))
				}
			}
			//fmt.Printf("%s", hclFile.Bytes())
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo AWS Cloudfront Distribution : %s", shortName)

			outVars := generateCFDOutputVars(varFullPrefix, shortName)
			tfContext.OutputVars = append(tfContext.OutputVars, outVars...)
			// Import all created resources.
			if config.GenerateTfState {
				importConfigs := []common.ImportConfig{}
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_cloudfront_distribution." + shortName,
					ResourceId:      config.TenantId + "/" + cfd.Id,
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
	}
	log.Println("[TRACE] <====== AWS Cloudfront Distribution TF generation done. =====>")

	return &tfContext, nil
}

func generateCFDOutputVars(prefix, shortName string) []common.OutputVarConfig {
	outVarConfigs := make(map[string]common.OutputVarConfig)

	var1 := common.OutputVarConfig{
		Name:          prefix + "arn",
		ActualVal:     "duplocloud_aws_cloudfront_distribution." + shortName + ".arn",
		DescVal:       "The ARN for the distribution.",
		RootTraversal: true,
	}
	outVarConfigs["arn"] = var1

	var2 := common.OutputVarConfig{
		Name:          prefix + "id",
		ActualVal:     "duplocloud_aws_cloudfront_distribution." + shortName + ".id",
		DescVal:       "The identifier for the distribution.",
		RootTraversal: true,
	}
	outVarConfigs["id"] = var2

	var3 := common.OutputVarConfig{
		Name:          prefix + "domain_name",
		ActualVal:     "duplocloud_aws_cloudfront_distribution." + shortName + ".domain_name",
		DescVal:       "The domain name corresponding to the distribution.",
		RootTraversal: true,
	}
	outVarConfigs["domain_name"] = var3

	outVars := make([]common.OutputVarConfig, len(outVarConfigs))
	for _, v := range outVarConfigs {
		outVars = append(outVars, v)
	}
	return outVars
}
