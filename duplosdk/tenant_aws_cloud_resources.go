package duplosdk

import (
	"fmt"
	"log"
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
)

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
	IsInternal bool   `json:"IsInternal,omitempty"`
	WebACLID   string `json:"WebACLID,omitempty"`
}

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
