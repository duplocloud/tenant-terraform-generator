package duplosdk

import (
	"fmt"
)

// DuploEcacheInstance is a Duplo SDK object that represents an ECache instance
type DuploEcacheInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Identifier          string `json:"Identifier"`
	Arn                 string `json:"Arn"`
	Endpoint            string `json:"Endpoint,omitempty"`
	CacheType           int    `json:"CacheType,omitempty"`
	Size                string `json:"Size,omitempty"`
	Replicas            int    `json:"Replicas,omitempty"`
	EncryptionAtRest    bool   `json:"EnableEncryptionAtRest,omitempty"`
	EncryptionInTransit bool   `json:"EnableEncryptionAtTransit,omitempty"`
	KMSKeyID            string `json:"KmsKeyId,omitempty"`
	AuthToken           string `json:"AuthToken,omitempty"`
	InstanceStatus      string `json:"InstanceStatus,omitempty"`
}

func (c *Client) EcacheInstanceList(tenantID string) (*[]DuploEcacheInstance, ClientError) {
	rp := []DuploEcacheInstance{}
	err := c.getAPI(fmt.Sprintf("RdsInstanceList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetEcacheInstances", tenantID),
		&rp)
	return &rp, err
}
