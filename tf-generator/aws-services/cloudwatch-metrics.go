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

const CWM_VAR_PREFIX = "cwm_"

type CloudwatchMetrics struct {
}

func (cwm *CloudwatchMetrics) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	list, clientErr := client.DuploCloudWatchMetricAlarmList(config.TenantId)

	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, clientErr
	}
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}
	if list != nil {
		hostList, _ := client.NativeHostGetList(config.TenantId)
		rdsList, _ := client.RdsInstanceList(config.TenantId)
		dynamoDBList, _ := client.TenantDynamoDBList(config.TenantId)
		log.Println("[TRACE] <====== Cloudwatch metrics TF generation started. =====>")
		for i, cwm := range *list {
			friendlyNames := []string{}
			namespace := strings.Split(cwm.Namespace, "/")[0]
			if len(strings.Split(cwm.Namespace, "/")) > 1 {
				namespace = strings.Split(cwm.Namespace, "/")[1]
			}
			shortName := cwm.MetricName + "-" + namespace + "-" + strconv.Itoa(i+1)
			resourceName := common.GetResourceName(shortName)
			log.Printf("[TRACE] Generating terraform config for duplo Cloudwatch metrics : %s", shortName)

			// create new empty hcl file object
			hclFile := hclwrite.NewEmptyFile()

			// create new file on system
			path := filepath.Join(workingDir, "cwm-"+shortName+".tf")
			tfFile, err := os.Create(path)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			// initialize the body of the new file object
			rootBody := hclFile.Body()

			cwmBlock := rootBody.AppendNewBlock("resource",
				[]string{"duplocloud_aws_cloudwatch_metric_alarm",
					resourceName})
			cwmBody := cwmBlock.Body()
			cwmBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "tenant_id",
				},
			})

			cwmBody.SetAttributeValue("metric_name",
				cty.StringVal(cwm.MetricName))

			cwmBody.SetAttributeValue("comparison_operator",
				cty.StringVal(cwm.ComparisonOperator))

			cwmBody.SetAttributeValue("evaluation_periods",
				cty.NumberIntVal(int64(cwm.EvaluationPeriods)))

			if len(cwm.Namespace) > 0 {
				cwmBody.SetAttributeValue("namespace",
					cty.StringVal(cwm.Namespace))
			}
			if len(cwm.Statistic) > 0 {
				cwmBody.SetAttributeValue("statistic",
					cty.StringVal(cwm.Statistic))
			}
			if cwm.Period > 0 {
				cwmBody.SetAttributeValue("period",
					cty.NumberIntVal(int64(cwm.Period)))
			}
			if cwm.Threshold > 0 {
				cwmBody.SetAttributeValue("threshold",
					cty.NumberIntVal(int64(cwm.Threshold)))
			}
			if cwm.Dimensions != nil && len(*cwm.Dimensions) > 0 {
				for _, dim := range *cwm.Dimensions {
					valAssigned := false
					friendlyNames = append(friendlyNames, dim.Value)
					dimBlock := cwmBody.AppendNewBlock("dimension",
						nil)
					dimBody := dimBlock.Body()
					dimBody.SetAttributeValue("key", cty.StringVal(dim.Name))
					if hostList != nil && dim.Name == "InstanceId" {
						for _, host := range *hostList {
							if dim.Value == host.InstanceID {
								valAssigned = true
								hostShortName := host.FriendlyName[len("duploservices-"+config.TenantName+"-"):len(host.FriendlyName)]
								dimBody.SetAttributeTraversal("value", hcl.Traversal{
									hcl.TraverseRoot{
										Name: "duplocloud_aws_host." + hostShortName,
									},
									hcl.TraverseAttr{
										Name: "instance_id",
									},
								})
								break
							}
						}
					}
					if rdsList != nil && dim.Name == "DBInstanceIdentifier" {
						for _, rds := range *rdsList {
							if dim.Value == rds.Identifier {
								valAssigned = true
								rdsShortName := rds.Identifier[len("duplo"):len(rds.Identifier)]
								dimBody.SetAttributeTraversal("value", hcl.Traversal{
									hcl.TraverseRoot{
										Name: "duplocloud_rds_instance." + rdsShortName,
									},
									hcl.TraverseAttr{
										Name: "fullname	",
									},
								})
								break
							}
						}
					}
					if dynamoDBList != nil && dim.Name == "TableName" {
						for _, dynamodb := range *dynamoDBList {
							if dim.Value == dynamodb.Name {
								valAssigned = true
								dynamoDBShortName, _ := extractDynamoDBName(client, config.TenantId, dynamodb.Name)
								dimBody.SetAttributeTraversal("value", hcl.Traversal{
									hcl.TraverseRoot{
										Name: "duplocloud_aws_dynamodb_table_v2." + dynamoDBShortName,
									},
									hcl.TraverseAttr{
										Name: "name	",
									},
								})
								break
							}
						}
					}
					if !valAssigned {
						dimBody.SetAttributeValue("value", cty.StringVal(dim.Value))
					}
				}
			}
			friendlyNames = append(friendlyNames, cwm.MetricName)
			_, err = tfFile.Write(hclFile.Bytes())
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			log.Printf("[TRACE] Terraform config is generated for duplo Cloudwatch metrics : %s", shortName)

			// Import all created resources.
			if config.GenerateTfState {
				importConfigs = append(importConfigs, common.ImportConfig{
					ResourceAddress: "duplocloud_aws_cloudwatch_metric_alarm." + resourceName,
					ResourceId:      config.TenantId + "/" + strings.Join(friendlyNames, "-"),
					WorkingDir:      workingDir,
				})
				tfContext.ImportConfigs = importConfigs
			}
		}
		log.Println("[TRACE] <====== Cloudwatch Metrics TF generation done. =====>")
	}

	return &tfContext, nil
}
