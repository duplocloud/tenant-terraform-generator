package duplosdk

import (
	"fmt"
)

// DuploK8sConfigMap represents a kubernetes config map in a Duplo tenant
type DuploK8sConfigMap struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"` //nolint:govet

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"-"` //nolint:govet

	Data     map[string]interface{} `json:"data,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// K8ConfigMapGetList retrieves a list of k8s config maps via the Duplo API.
func (c *Client) K8ConfigMapGetList(tenantID string) (*[]DuploK8sConfigMap, ClientError) {
	rp := []DuploK8sConfigMap{}
	err := c.getAPI(
		fmt.Sprintf("K8ConfigMapGetList(%s)", tenantID),
		fmt.Sprintf("v2/subscriptions/%s/K8ConfigMapApiV2", tenantID),
		&rp)

	// Add the tenant Id and name, then return the result.
	if err == nil {
		for i := range rp {
			rp[i].TenantID = tenantID
			if name, ok := rp[i].Metadata["name"]; ok {
				rp[i].Name = name.(string)
			}
		}
	}
	return &rp, err
}
