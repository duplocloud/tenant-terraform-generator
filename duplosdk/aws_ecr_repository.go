package duplosdk

import (
	"fmt"
)

type DuploAwsEcrRepositoryRequest struct {
	KmsEncryption         string `json:"KmsEncryption,omitempty"`
	EnableTagImmutability bool   `json:"EnableTagImmutability,omitempty"`
	EnableScanImageOnPush bool   `json:"EnableScanImageOnPush,omitempty"`
	Name                  string `json:"Name"`
}

type DuploAwsEcrRepository struct {
	KmsEncryption         string `json:"KmsEncryption,omitempty"`
	KmsEncryptionAlias    string `json:"KmsEncryptionAlias,omitempty"`
	EnableTagImmutability bool   `json:"EnableTagImmutability,omitempty"`
	EnableScanImageOnPush bool   `json:"EnableScanImageOnPush,omitempty"`
	Arn                   string `json:"Arn"`
	ResourceType          int    `json:"ResourceType,omitempty"`
	Name                  string `json:"Name"`
	RegistryId            string `json:"RegistryId,omitempty"`
	RepositoryUri         string `json:"RepositoryUri,omitempty"`
}

func (c *Client) AwsEcrRepositoryList(tenantID string) (*[]DuploAwsEcrRepository, ClientError) {
	rp := []DuploAwsEcrRepository{}
	err := c.getAPI(
		fmt.Sprintf("AwsEcrRepositoryList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/ecrRepository", tenantID),
		&rp,
	)
	return &rp, err
}
