package duplosdk

import "fmt"

type DuploCloudWatchEventRule struct {
	Name               string                 `json:"Name"`
	Description        string                 `json:"Description,omitempty"`
	ScheduleExpression string                 `json:"ScheduleExpression"`
	State              string                 `json:"State"`
	Tags               *[]DuploKeyStringValue `json:"Tags,omitempty"`
	EventBusName       string                 `json:"EventBusName,omitempty"`
	RoleArn            string                 `json:"RoleArn,omitempty"`
}

type DuploCloudWatchEventRuleGetReq struct {
	Name               string            `json:"Name"`
	Description        string            `json:"Description,omitempty"`
	ScheduleExpression string            `json:"ScheduleExpression"`
	EventBusName       string            `json:"EventBusName,omitempty"`
	RoleArn            string            `json:"RoleArn,omitempty"`
	Arn                string            `json:"Arn,omitempty"`
	State              *DuploStringValue `json:"State,omitempty"`
}

type DuploCloudWatchEventTargets struct {
	Rule         string                        `json:"Rule"`
	Targets      *[]DuploCloudWatchEventTarget `json:"Targets,omitempty"`
	EventBusName string                        `json:"EventBusName,omitempty"`
}

type DuploCloudWatchEventTargetsDeleteReq struct {
	Rule string   `json:"Rule,omitempty"`
	Ids  []string `json:"Ids,omitempty"`
}

type DuploCloudWatchEventTarget struct {
	Arn     string `json:"Arn"`
	Id      string `json:"Id,omitempty"`
	RoleArn string `json:"RoleArn,omitempty"`
	Input   string `json:"Input,omitempty"`
}

type DuploCloudWatchRunCommandTarget struct {
	Key    string   `json:"Key,omitempty"`
	Values []string `json:"Values,omitempty"`
}

type DuploCloudWatchMetricAlarm struct {
	Statistic          string                  `json:"Statistic,omitempty"`
	MetricName         string                  `json:"MetricName,omitempty"`
	ComparisonOperator string                  `json:"ComparisonOperator,omitempty"`
	Threshold          float64                 `json:"Threshold,omitempty"`
	Period             int                     `json:"Period,omitempty"`
	EvaluationPeriods  int                     `json:"EvaluationPeriods,omitempty"`
	TenantId           string                  `json:"TenantId,omitempty"`
	Namespace          string                  `json:"Namespace,omitempty"`
	State              string                  `json:"State,omitempty"`
	Dimensions         *[]DuploNameStringValue `json:"Dimensions,omitempty"`
	AccountName        string                  `json:"AccountName,omitempty"`
	Name               string                  `json:"Name,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

func (c *Client) DuploCloudWatchEventRuleList(tenantID string) (*[]DuploCloudWatchEventRuleGetReq, ClientError) {
	rp := []DuploCloudWatchEventRuleGetReq{}
	err := c.getAPI(
		fmt.Sprintf("DuploCloudWatchEventRuleList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetAwsEventRules", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploCloudWatchEventTargetsList(tenantID string, ruleName string) (*[]DuploCloudWatchEventTarget, ClientError) {
	rp := []DuploCloudWatchEventTarget{}
	err := c.getAPI(
		fmt.Sprintf("DuploCloudWatchEventTargetsList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/eventTargets/%s", tenantID, ruleName),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploCloudWatchMetricAlarmGet(tenantID, resourceId string) (*DuploCloudWatchMetricAlarm, ClientError) {
	rp := []DuploCloudWatchMetricAlarm{}
	err := c.getAPI(
		fmt.Sprintf("DuploCloudWatchMetricAlarmGet(%s, %s)", tenantID, resourceId),
		fmt.Sprintf("subscriptions/%s/%s/GetAlarms", tenantID, EncodePathParam(resourceId)),
		&rp,
	)
	if len(rp) == 0 {
		return nil, err
	}
	return &rp[0], err
}

func (c *Client) DuploCloudWatchMetricAlarmList(tenantID string) (*[]DuploCloudWatchMetricAlarm, ClientError) {
	rp := []DuploCloudWatchMetricAlarm{}
	err := c.getAPI(
		fmt.Sprintf("DuploCloudWatchMetricAlarmList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/*/GetAlarms", tenantID),
		&rp,
	)
	if len(rp) == 0 {
		return nil, err
	}
	return &rp, err
}
