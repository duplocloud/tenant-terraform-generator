package duplosdk

import (
	"fmt"
	"log"
	"time"
)

const (
	// ResourceTypeS3Bucket represents an S3 bucket
	ResourceTypeS3Bucket int = 1

	// ResourceTypeDynamoDBTable represents an DynamoDB table
	ResourceTypeDynamoDBTable int = 2

	// ResourceTypeKafkaCluster represents a Kafka cluster
	ResourceTypeKafkaCluster int = 14

	// ResourceTypeApplicationLB represents an AWS application LB
	ResourceTypeApplicationLB int = 16

	// ResourceTypeApiGatewayRestAPI represents an AWS Api gateway REST API
	ResourceTypeApiGatewayRestAPI int = 8

	ResourceTypeSQSQueue int = 3

	ResourceTypeSNSTopic int = 4
)

type CustomComponentType int

const (
	NETWORK CustomComponentType = iota
	REPLICATIONCONTROLLER
	MINION
	ASG
	AGENTPOOL
)

type DuploHostOOBData struct {
	IPAddress   string                 `json:"IpAddress"`
	InstanceId  string                 `json:"InstanceId"`
	Cloud       int                    `json:"Cloud"`
	Credentials *[]DuploHostCredential `json:"Credentials"`
}

type DuploHostCredential struct {
	Username   string `json:"Username"`
	Password   string `json:"Password,omitempty"`
	Privatekey string `json:"Privatekey,omitempty"`
}

type CustomDataUpdate struct {
	ComponentId   string              `json:"ComponentId,omitempty"`
	ComponentType CustomComponentType `json:"ComponentType,omitempty"`
	State         string              `json:"State,omitempty"`
	Key           string              `json:"Key,omitempty"`
	Value         string              `json:"Value,omitempty"`
}

type DuploMinion struct {
	Name             string                 `json:"Name"`
	ConnectionURL    string                 `json:"ConnectionUrl"`
	NetworkAgentURL  string                 `json:"NetworkAgentUrl,omitempty"`
	ConnectionStatus string                 `json:"ConnectionStatus,omitempty"`
	Subnet           string                 `json:"Subnet,omitempty"`
	DirectAddress    string                 `json:"DirectAddress"`
	Tags             *[]DuploKeyStringValue `json:"Tags,omitempty"`
	ExternalAddress  string                 `json:"ExternalAddress,omitempty"`
	Tier             string                 `json:"Tier"`
	UserAccount      string                 `json:"UserAccount,omitempty"`
	Tunnel           int                    `json:"Tunnel"`
	AgentPlatform    int                    `json:"AgentPlatform"`
	Cloud            int                    `json:"Cloud"`
}

// DuploAwsCloudResource represents a generic AWS cloud resource for a Duplo tenant
type DuploAwsCloudResource struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Type     int    `json:"ResourceType,omitempty"`
	Name     string `json:"Name,omitempty"`
	Arn      string `json:"Arn,omitempty"`
	MetaData string `json:"MetaData,omitempty"`

	// S3 bucket and load balancer
	EnableAccessLogs bool                   `json:"EnableAccessLogs,omitempty"`
	Tags             *[]DuploKeyStringValue `json:"Tags,omitempty"`

	// Only S3 bucket
	EnableVersioning  bool     `json:"EnableVersioning,omitempty"`
	AllowPublicAccess bool     `json:"AllowPublicAccess,omitempty"`
	DefaultEncryption string   `json:"DefaultEncryption,omitempty"`
	Policies          []string `json:"Policies,omitempty"`

	// Only Load balancer
	IsInternal bool              `json:"IsInternal,omitempty"`
	WebACLID   string            `json:"WebACLID,omitempty"`
	LbType     *DuploStringValue `json:"LbType,omitempty"`
}

// DuploS3Bucket represents an S3 bucket resource for a Duplo tenant
type DuploS3Bucket struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name              string                 `json:"Name,omitempty"`
	Arn               string                 `json:"Arn,omitempty"`
	MetaData          string                 `json:"MetaData,omitempty"`
	EnableVersioning  bool                   `json:"EnableVersioning,omitempty"`
	EnableAccessLogs  bool                   `json:"EnableAccessLogs,omitempty"`
	AllowPublicAccess bool                   `json:"AllowPublicAccess,omitempty"`
	DefaultEncryption string                 `json:"DefaultEncryption,omitempty"`
	Policies          []string               `json:"Policies,omitempty"`
	Tags              *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

// DuploApplicationLB represents an AWS application load balancer resource for a Duplo tenant
type DuploApplicationLB struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name             string                 `json:"Name,omitempty"`
	Arn              string                 `json:"Arn,omitempty"`
	DNSName          string                 `json:"MetaData,omitempty"`
	EnableAccessLogs bool                   `json:"EnableAccessLogs,omitempty"`
	IsInternal       bool                   `json:"IsInternal,omitempty"`
	WebACLID         string                 `json:"WebACLID,omitempty"`
	LbType           *DuploStringValue      `json:"LbType,omitempty"`
	Tags             *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

// DuploAwsLBConfiguration represents a request to create an AWS application load balancer resource
type DuploAwsLBConfiguration struct {
	Name             string `json:"Name"`
	State            string `json:"State,omitempty"`
	IsInternal       bool   `json:"IsInternal,omitempty"`
	EnableAccessLogs bool   `json:"EnableAccessLogs,omitempty"`
}

type DuploAwsLbState struct {
	Code *DuploStringValue `json:"Code,omitempty"`
}

type DuploAwsLbAvailabilityZone struct {
	SubnetID string `json:"SubnetId,omitempty"`
	ZoneName string `json:"ZoneName,omitempty"`
}

// DuploAwsLbSettings represents an AWS application load balancer's details, via a Duplo Service
type DuploAwsLbDetailsInService struct {
	LoadBalancerName      string                       `json:"LoadBalancerName"`
	LoadBalancerArn       string                       `json:"LoadBalancerArn"`
	AvailabilityZones     []DuploAwsLbAvailabilityZone `json:"AvailabilityZones"`
	CanonicalHostedZoneId string                       `json:"CanonicalHostedZoneId"`
	CreatedTime           time.Time                    `json:"CreatedTime"`
	DNSName               string                       `json:"DNSName"`
	IPAddressType         *DuploStringValue            `json:"IPAddressType,omitempty"`
	Scheme                *DuploStringValue            `json:"Scheme,omitempty"`
	Type                  *DuploStringValue            `json:"Type,omitempty"`
	SecurityGroups        []string                     `json:"SecurityGroups"`
	State                 *DuploAwsLbState             `json:"State,omitempty"`
	VpcID                 string                       `json:"VpcId,omitempty"`
}

// DuploAwsLbListenerCertificate represents a AWS load balancer listener's SSL Certificate
type DuploAwsLbListenerCertificate struct {
	CertificateArn string `json:"CertificateArn"`
	IsDefault      bool   `json:"IsDefault,omitempty"`
}

// DuploAwsLbListenerAction represents a AWS load balancer listener action
type DuploAwsLbListenerAction struct {
	Order          int               `json:"Order"`
	TargetGroupArn string            `json:"TargetGroupArn"`
	Type           *DuploStringValue `json:"Type,omitempty"`
}

type DuploAwsLbListenerActionCreate struct {
	TargetGroupArn string `json:"TargetGroupArn"`
	Type           string `json:"Type,omitempty"`
}

type DuploAwsLbListenerDeleteRequest struct {
	ListenerArn string `json:"ListenerArn"`
}

// DuploAwsLbSettings represents an AWS application load balancer's settings
type DuploAwsLbSettings struct {
	LoadBalancerArn     string `json:"LoadBalancerArn"`
	EnableAccessLogs    bool   `json:"EnableAccessLogs,omitempty"`
	DropInvalidHeaders  bool   `json:"DropInvalidHeaders,omitempty"`
	WebACLID            string `json:"WebACLId,omitempty"`
	HttpToHttpsRedirect bool   `json:"HttpToHttpsRedirect,omitempty"`
}

// DuploAwsLbListener represents an AWS application load balancer listener
type DuploAwsLbListener struct {
	LoadBalancerArn string                          `json:"LoadBalancerArn"`
	Certificates    []DuploAwsLbListenerCertificate `json:"Certificates"`
	ListenerArn     string                          `json:"ListenerArn"`
	SSLPolicy       string                          `json:"SslPolicy,omitempty"`
	Port            int                             `json:"Port"`
	Protocol        *DuploStringValue               `json:"Protocol,omitempty"`
	DefaultActions  []DuploAwsLbListenerAction      `json:"DefaultActions"`
}

type DuploAwsLbListenerCreate struct {
	Certificates   []DuploAwsLbListenerCertificate  `json:"Certificates"`
	Port           int                              `json:"Port"`
	Protocol       string                           `json:"Protocol,omitempty"`
	DefaultActions []DuploAwsLbListenerActionCreate `json:"DefaultActions"`
}

// DuploAwsTargetGroupMatcher represents an AWS lb target group matcher
type DuploAwsTargetGroupMatcher struct {
	HttpCode string `json:"HttpCode"`
	GrpcCode string `json:"GrpcCode"`
}

// DuploAwsLbTargetGroup represents an AWS lb target group
type DuploAwsLbTargetGroup struct {
	HealthCheckEnabled         bool                        `json:"HealthCheckEnabled"`
	HealthCheckIntervalSeconds int                         `json:"HealthCheckIntervalSeconds"`
	HealthCheckPath            string                      `json:"HealthCheckPath"`
	HealthCheckPort            string                      `json:"HealthCheckPort"`
	HealthCheckProtocol        *DuploStringValue           `json:"HealthCheckProtocol,omitempty"`
	HealthyThreshold           int                         `json:"HealthyThresholdCount"`
	HealthCheckTimeoutSeconds  int                         `json:"HealthCheckTimeoutSeconds"`
	LoadBalancerArns           []string                    `json:"LoadBalancerArns"`
	HealthMatcher              *DuploAwsTargetGroupMatcher `json:"Matcher,omitempty"`
	Protocol                   *DuploStringValue           `json:"Protocol,omitempty"`
	ProtocolVersion            string                      `json:"ProtocolVersion"`
	TargetGroupArn             string                      `json:"TargetGroupArn"`
	TargetGroupName            string                      `json:"TargetGroupName"`
	TargetType                 *DuploStringValue           `json:"TargetType,omitempty"`
	UnhealthyThreshold         int                         `json:"UnhealthThresholdCount"`
	VpcID                      string                      `json:"VpcId"`
}

// DuploAwsLBAccessLogsRequest represents a request to retrieve an AWS application load balancer's settings.
type DuploAwsLbSettingsRequest struct {
	LoadBalancerArn string `json:"LoadBalancerArn"`
}

// DuploAwsLBAccessLogsUpdateRequest represents a request to update an AWS application load balancer's settings.
type DuploAwsLbSettingsUpdateRequest struct {
	LoadBalancerArn    string `json:"LoadBalancerArn"`
	EnableAccessLogs   bool   `json:"EnableAccessLogs,omitempty"`
	DropInvalidHeaders bool   `json:"DropInvalidHeaders,omitempty"`
	WebACLID           string `json:"WebACLId,omitempty"`
}

// DuploS3BucketRequest represents a request to create an S3 bucket resource
type DuploS3BucketRequest struct {
	Type           int    `json:"ResourceType"`
	Name           string `json:"Name"`
	State          string `json:"State,omitempty"`
	InTenantRegion bool   `json:"InTenantRegion"`
}

// DuploS3BucketSettingsRequest represents a request to create an S3 bucket resource
type DuploS3BucketSettingsRequest struct {
	Name              string   `json:"Name"`
	EnableVersioning  bool     `json:"EnableVersioning,omitempty"`
	EnableAccessLogs  bool     `json:"EnableAccessLogs,omitempty"`
	AllowPublicAccess bool     `json:"AllowPublicAccess,omitempty"`
	DefaultEncryption string   `json:"DefaultEncryption,omitempty"`
	Policies          []string `json:"Policies,omitempty"`
}

type DuploApiGatewayRequest struct {
	Name           string `json:"Name"`
	LambdaFunction string `json:"LambdaFunction,omitempty"`
	State          string `json:"State,omitempty"`
}

type DuploApiGatewayResource struct {
	Name         string `json:"Name"`
	MetaData     string `json:"MetaData,omitempty"`
	ResourceType int    `json:"ResourceType,omitempty"`
}

type DuploAwsResource struct {
	Name         string `json:"Name"`
	ResourceType int    `json:"ResourceType,omitempty"`
}

// TenantListAwsCloudResources retrieves a list of the generic AWS cloud resources for a tenant via the Duplo API.
func (c *Client) TenantListAwsCloudResources(tenantID string) (*[]DuploAwsCloudResource, ClientError) {
	apiName := fmt.Sprintf("TenantListAwsCloudResources(%s)", tenantID)
	list := []DuploAwsCloudResource{}

	// Get the list from Duplo
	err := c.getAPI(apiName, fmt.Sprintf("subscriptions/%s/GetCloudResources", tenantID), &list)
	if err != nil {
		return nil, err
	}

	// Add the tenant ID to each element and return the list.
	log.Printf("[TRACE] %s: %d items", apiName, len(list))
	for i := range list {
		list[i].TenantID = tenantID
	}
	return &list, nil
}

// TenantGetAwsCloudResource retrieves a cloud resource by type and name
func (c *Client) TenantGetAwsCloudResource(tenantID string, resourceType int, name string) (*DuploAwsCloudResource, ClientError) {
	allResources, err := c.TenantListAwsCloudResources(tenantID)
	if err != nil {
		return nil, err
	}

	// Find and return the secret with the specific type and name.
	for _, resource := range *allResources {
		if resource.Type == resourceType && resource.Name == name {
			return &resource, nil
		}
	}

	// No resource was found.
	return nil, nil
}

// TenantGetApplicationLbFullName retrieves the full name of a pass-thru AWS application load balancer.
func (c *Client) TenantGetApplicationLbFullName(tenantID string, name string) (string, ClientError) {
	return c.GetResourceName("duplo3", tenantID, name, false)
}

// TenantGetS3Bucket retrieves a managed S3 bucket via the Duplo API
func (c *Client) TenantGetS3Bucket(tenantID string, name string) (*DuploS3Bucket, ClientError) {
	// Figure out the full resource name.
	fullName, err := c.GetDuploServicesNameWithAws(tenantID, name)
	if err != nil {
		return nil, err
	}

	// Get the resource from Duplo.
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeS3Bucket, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploS3Bucket{
		TenantID:          tenantID,
		Name:              resource.Name,
		Arn:               resource.Arn,
		MetaData:          resource.MetaData,
		EnableVersioning:  resource.EnableVersioning,
		AllowPublicAccess: resource.AllowPublicAccess,
		EnableAccessLogs:  resource.EnableAccessLogs,
		DefaultEncryption: resource.DefaultEncryption,
		Policies:          resource.Policies,
		Tags:              resource.Tags,
	}, nil
}

// TenantGetApplicationLB retrieves an application load balancer via the Duplo API
func (c *Client) TenantGetApplicationLB(tenantID string, name string) (*DuploApplicationLB, ClientError) {
	// Figure out the full resource name.
	fullName, err := c.TenantGetApplicationLbFullName(tenantID, name)
	if err != nil {
		return nil, err
	}

	// Get the resource from Duplo.
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeApplicationLB, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploApplicationLB{
		TenantID:         tenantID,
		Name:             resource.Name,
		Arn:              resource.Arn,
		DNSName:          resource.MetaData,
		IsInternal:       resource.IsInternal,
		EnableAccessLogs: resource.EnableAccessLogs,
		Tags:             resource.Tags,
	}, nil
}

// TenantCreateS3Bucket creates an S3 bucket resource via Duplo.
func (c *Client) TenantCreateS3Bucket(tenantID string, duplo DuploS3BucketRequest) ClientError {
	duplo.Type = ResourceTypeS3Bucket

	// Create the bucket via Duplo.
	return c.postAPI(
		fmt.Sprintf("TenantCreateS3Bucket(%s, %s)", tenantID, duplo.Name),
		fmt.Sprintf("subscriptions/%s/S3BucketUpdate", tenantID),
		&duplo,
		nil)
}

// TenantDeleteS3Bucket deletes an S3 bucket resource via Duplo.
func (c *Client) TenantDeleteS3Bucket(tenantID string, name string) ClientError {

	// Get the full name of the S3 bucket
	fullName, err := c.GetDuploServicesNameWithAws(tenantID, name)
	if err != nil {
		return err
	}

	// Delete the bucket via Duplo.
	return c.postAPI(
		fmt.Sprintf("TenantDeleteS3Bucket(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/S3BucketUpdate", tenantID),
		&DuploS3BucketRequest{Type: ResourceTypeS3Bucket, Name: fullName, State: "delete"},
		nil)
}

// TenantGetS3BucketSettings gets a non-cached view of the  S3 buckets's settings via Duplo.
func (c *Client) TenantGetS3BucketSettings(tenantID string, name string) (*DuploS3Bucket, ClientError) {
	rp := DuploS3Bucket{}

	err := c.getAPI(fmt.Sprintf("TenantGetS3BucketSettings(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/GetS3BucketSettings/%s", tenantID, name),
		&rp)
	if err != nil || rp.Name == "" {
		return nil, err
	}
	return &rp, err
}

// TenantApplyS3BucketSettings applies settings to an S3 bucket resource via Duplo.
func (c *Client) TenantApplyS3BucketSettings(tenantID string, duplo DuploS3BucketSettingsRequest) (*DuploS3Bucket, ClientError) {
	apiName := fmt.Sprintf("TenantApplyS3BucketSettings(%s, %s)", tenantID, duplo.Name)

	// Figure out the full resource name.
	fullName, err := c.GetDuploServicesNameWithAws(tenantID, duplo.Name)
	if err != nil {
		return nil, err
	}
	duplo.Name = fullName

	// Apply the settings via Duplo.
	rp := DuploS3Bucket{}
	err = c.postAPI(apiName, fmt.Sprintf("subscriptions/%s/ApplyS3BucketSettings", tenantID), &duplo, &rp)
	if err != nil {
		return nil, err
	}

	// Deal with a missing response.
	if rp.Name == "" {
		message := fmt.Sprintf("%s: unexpected missing response from backend", apiName)
		log.Printf("[TRACE] %s", message)
		return nil, newClientError(message)
	}

	// Return the response.
	rp.TenantID = tenantID
	return &rp, nil
}

// TenantCreateKafkaCluster creates a kafka cluster resource via Duplo.
func (c *Client) TenantCreateKafkaCluster(tenantID string, duplo DuploKafkaClusterRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("TenantCreateKafkaCluster(%s, %s)", tenantID, duplo.Name),
		fmt.Sprintf("subscriptions/%s/KafkaClusterUpdate", tenantID),
		&duplo,
		nil)
}

// TenantDeleteKafkaCluster deletes a kafka cluster resource via Duplo.
func (c *Client) TenantDeleteKafkaCluster(tenantID, arn string) ClientError {
	return c.postAPI(
		fmt.Sprintf("TenantDeleteKafkaCluster(%s, %s)", tenantID, arn),
		fmt.Sprintf("subscriptions/%s/KafkaClusterUpdate", tenantID),
		&DuploKafkaClusterRequest{Arn: arn, State: "delete"},
		nil)
}

// TenantUpdateApplicationLbSettings updates an application LB resource's settings via Duplo.
func (c *Client) TenantUpdateApplicationLbSettings(tenantID string, duplo DuploAwsLbSettingsUpdateRequest) ClientError {
	return c.postAPI("TenantUpdateApplicationLbSettings",
		fmt.Sprintf("subscriptions/%s/UpdateLbSettings", tenantID),
		&duplo,
		nil)
}

// TenantGetApplicationLbSettings updates an application LB resource's WAF association via Duplo.
func (c *Client) TenantGetApplicationLbSettings(tenantID string, loadBalancerArn string) (*DuploAwsLbSettings, ClientError) {
	rp := DuploAwsLbSettings{}

	err := c.postAPI("TenantGetApplicationLbSettings",
		fmt.Sprintf("subscriptions/%s/GetLbSettings", tenantID),
		&DuploAwsLbSettingsRequest{LoadBalancerArn: loadBalancerArn},
		&rp)

	return &rp, err
}

// TenantGetLbDetailsInService retrieves load balancer details via a Duplo service.
func (c *Client) TenantGetLbDetailsInService(tenantID string, name string) (*DuploAwsLbDetailsInService, ClientError) {
	apiName := fmt.Sprintf("TenantGetLbDetailsInService(%s, %s)", tenantID, name)
	details := DuploAwsLbDetailsInService{}

	// Get the list from Duplo
	err := c.getAPI(apiName, fmt.Sprintf("subscriptions/%s/GetLbDetailsInService/%s", tenantID, name), &details)
	if err != nil {
		return nil, err
	}

	return &details, nil
}

// TenantCreateApplicationLB creates an application LB resource via Duplo.
func (c *Client) TenantCreateApplicationLB(tenantID string, duplo DuploAwsLBConfiguration) ClientError {
	return c.postAPI("TenantCreateApplicationLB",
		fmt.Sprintf("subscriptions/%s/ApplicationLbUpdate", tenantID),
		&duplo,
		nil)
}

// TenantDeleteApplicationLB deletes an AWS application LB resource via Duplo.
func (c *Client) TenantDeleteApplicationLB(tenantID string, name string) ClientError {
	// Get the full name of the ALB.
	fullName, err := c.TenantGetApplicationLbFullName(tenantID, name)
	if err != nil {
		return err
	}

	// Call the API.
	return c.postAPI("TenantDeleteApplicationLB",
		fmt.Sprintf("subscriptions/%s/ApplicationLbUpdate", tenantID),
		&DuploAwsLBConfiguration{Name: fullName, State: "delete"},
		nil)
}

// TenantListApplicationLbTargetGroups retrieves a list of AWS LB target groups
func (c *Client) TenantListApplicationLbTargetGroups(tenantID string) (*[]DuploAwsLbTargetGroup, ClientError) {
	rp := []DuploAwsLbTargetGroup{}

	err := c.getAPI("TenantListApplicationLbTargetGroups",
		fmt.Sprintf("subscriptions/%s/ListApplicationLbTargetGroups", tenantID),
		&rp)

	return &rp, err
}

// TenantListApplicationLbListeners retrieves a list of AWS LB listeners
func (c *Client) TenantListApplicationLbListeners(tenantID string, name string) (*[]DuploAwsLbListener, ClientError) {
	// Get the full name of the ALB.
	fullName, err := c.TenantGetApplicationLbFullName(tenantID, name)
	if err != nil {
		return nil, err
	}

	rp := []DuploAwsLbListener{}

	err = c.getAPI("TenantListApplicationLbListeners",
		fmt.Sprintf("subscriptions/%s/ListApplicationLbListerner/%s", tenantID, fullName),
		&rp)

	return &rp, err
}

func (c *Client) TenantUpdateCustomData(tenantID string, customeData CustomDataUpdate) ClientError {
	return c.postAPI("TenantUpdateCustomData",
		fmt.Sprintf("subscriptions/%s/UpdateCustomData", tenantID),
		customeData,
		nil)
}

func (c *Client) TenantApplicationLbListenersByTargetGrpArn(tenantID string, fullName string, targetGrpArn string) (*DuploAwsLbListener, ClientError) {
	rp := []DuploAwsLbListener{}

	err := c.getAPI("TenantListApplicationLbListeners",
		fmt.Sprintf("subscriptions/%s/ListApplicationLbListerner/%s", tenantID, fullName),
		&rp)
	for _, item := range rp {
		for _, action := range item.DefaultActions {
			if action.TargetGroupArn == targetGrpArn {
				return &item, nil
			}
		}
	}
	return nil, err
}

// TenantCreateApplicationLbListener creates a AWS LB listener
func (c *Client) TenantCreateApplicationLbListener(tenantID string, fullName string, duplo DuploAwsLbListenerCreate) ClientError {
	return c.postAPI("TenantCreateApplicationLB",
		fmt.Sprintf("subscriptions/%s/CreateApplicationLbListerner/%s", tenantID, fullName),
		&duplo,
		nil)
}

// TenantDeleteApplicationLbListener deletes an AWS application LB listener via Duplo.
func (c *Client) TenantDeleteApplicationLbListener(tenantID string, fullName string, listenerArn string) ClientError {
	// Call the API.
	return c.postAPI("TenantDeleteApplicationLB",
		fmt.Sprintf("subscriptions/%s/DeleteApplicationLbListerner/%s", tenantID, fullName),
		&DuploAwsLbListenerDeleteRequest{ListenerArn: listenerArn},
		nil)
}

func (c *Client) TenantCreateAPIGateway(tenantID string, duplo DuploApiGatewayRequest) ClientError {
	return c.postAPI("TenantCreateAPIGateway",
		fmt.Sprintf("subscriptions/%s/ApiGatewayRestApiUpdate", tenantID),
		&duplo,
		nil)
}

func (c *Client) TenantDeleteAPIGateway(tenantID, name string) ClientError {
	return c.postAPI("TenantCreateAPIGateway",
		fmt.Sprintf("subscriptions/%s/ApiGatewayRestApiUpdate", tenantID),
		&DuploApiGatewayRequest{Name: name, State: "delete"},
		nil)
}

func (c *Client) TenantListMinions(tenantID string) (*[]DuploMinion, ClientError) {
	apiName := fmt.Sprintf("TenantListMinions(%s)", tenantID)
	list := []DuploMinion{}

	err := c.getAPI(apiName, fmt.Sprintf("subscriptions/%s/GetMinions", tenantID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) TenantGetAPIGateway(tenantID string, fullName string) (*DuploApiGatewayResource, ClientError) {
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeApiGatewayRestAPI, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploApiGatewayResource{
		Name:         resource.Name,
		MetaData:     resource.MetaData,
		ResourceType: resource.Type,
	}, nil
}

func (c *Client) TenantListS3Buckets(tenantID string) (*[]DuploS3Bucket, ClientError) {
	allResources, err := c.TenantListAwsCloudResources(tenantID)
	m := make(map[string]DuploS3Bucket)
	if err != nil {
		return nil, err
	}
	for _, resource := range *allResources {
		if resource.Type == ResourceTypeS3Bucket {
			m[resource.Arn] = DuploS3Bucket{
				TenantID: tenantID,
				Name:     resource.Name,
				Arn:      resource.Arn,
			}
		}
	}
	buckets := make([]DuploS3Bucket, 0, len(m))
	for _, v := range m {
		buckets = append(buckets, v)
	}
	return &buckets, nil
}

func (c *Client) TenantListSQS(tenantID string) (*[]DuploAwsResource, ClientError) {
	allResources, err := c.TenantListAwsCloudResources(tenantID)
	m := make(map[string]DuploAwsResource)
	if err != nil {
		return nil, err
	}
	for _, resource := range *allResources {
		if resource.Type == ResourceTypeSQSQueue {
			m[resource.Name] = DuploAwsResource{
				Name:         resource.Name,
				ResourceType: ResourceTypeSQSQueue,
			}
		}
	}
	sqsList := make([]DuploAwsResource, 0, len(m))
	for _, i := range m {
		sqsList = append(sqsList, i)
	}
	return &sqsList, nil
}

func (c *Client) TenantListSnsTopic(tenantID string) (*[]DuploAwsResource, ClientError) {
	rp := []DuploAwsResource{}
	err := c.getAPI(
		fmt.Sprintf("TenantListSnsTopic(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/snsTopic", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) TenantGetApplicationLBList(tenantID string) (*[]DuploApplicationLB, ClientError) {
	allResources, err := c.TenantListAwsCloudResources(tenantID)
	m := make(map[string]DuploApplicationLB)
	if err != nil {
		return nil, err
	}
	for _, resource := range *allResources {
		if resource.Type == ResourceTypeApplicationLB {
			m[resource.Name] = DuploApplicationLB{
				TenantID:         tenantID,
				Name:             resource.Name,
				Arn:              resource.Arn,
				DNSName:          resource.MetaData,
				IsInternal:       resource.IsInternal,
				EnableAccessLogs: resource.EnableAccessLogs,
				Tags:             resource.Tags,
				LbType:           resource.LbType,
			}
		}
	}
	lbList := make([]DuploApplicationLB, 0, len(m))
	for _, i := range m {
		lbList = append(lbList, i)
	}
	return &lbList, nil
}

func (c *Client) TenantGetApplicationApiGatewayList(tenantID string) (*[]DuploApiGatewayResource, ClientError) {
	allResources, err := c.TenantListAwsCloudResources(tenantID)
	m := make(map[string]DuploApiGatewayResource)
	if err != nil {
		return nil, err
	}
	for _, resource := range *allResources {
		if resource.Type == ResourceTypeApiGatewayRestAPI {
			m[resource.Name] = DuploApiGatewayResource{
				Name:         resource.Name,
				MetaData:     resource.MetaData,
				ResourceType: resource.Type,
			}
		}
	}
	list := make([]DuploApiGatewayResource, 0, len(m))
	for _, i := range m {
		list = append(list, i)
	}
	return &list, nil
}

func (c *Client) TenantDynamoDBList(tenantID string) (*[]DuploAwsResource, ClientError) {
	allResources, err := c.TenantListAwsCloudResources(tenantID)
	m := make(map[string]DuploAwsResource)
	if err != nil {
		return nil, err
	}
	for _, resource := range *allResources {
		if resource.Type == ResourceTypeDynamoDBTable {
			m[resource.Name] = DuploAwsResource{
				Name:         resource.Name,
				ResourceType: resource.Type,
			}
		}
	}
	list := make([]DuploAwsResource, 0, len(m))
	for _, i := range m {
		list = append(list, i)
	}
	return &list, nil
}

func (c *Client) TenantByohList(tenantID string) (*[]DuploMinion, ClientError) {
	m := make(map[string]DuploMinion)
	list, err := c.TenantListMinions(tenantID)
	if err != nil {
		return nil, err
	}
	for _, minion := range *list {
		if minion.Cloud == 4 {
			m[minion.Name] = minion
		}
	}
	minionList := make([]DuploMinion, 0, len(m))
	for _, i := range m {
		minionList = append(minionList, i)
	}
	return &minionList, nil
}

func (c *Client) TenantHostCredentialsGet(tenantID string, duplo DuploHostOOBData) (*DuploHostCredential, ClientError) {
	resp := DuploHostCredential{}
	err := c.postAPI("TenantHostCredentialsGet",
		fmt.Sprintf("subscriptions/%s/FindHostCredentialsFromOOBData", tenantID),
		&duplo,
		&resp)
	if err != nil {
		return nil, err
	}
	return &resp, err
}
