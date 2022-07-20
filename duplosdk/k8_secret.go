package duplosdk

import (
	"fmt"
)

// DuploK8sSecret represents a kubernetes secret in a Duplo tenant
type DuploK8sSecret struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"` //nolint:govet

	SecretName        string                 `json:"SecretName"`
	SecretType        string                 `json:"SecretType"`
	SecretVersion     string                 `json:"SecretVersion,omitempty"`
	SecretData        map[string]interface{} `json:"SecretData"`
	SecretAnnotations map[string]string      `json:"SecretAnnotations,omitempty"`
}

// K8SecretGetList retrieves a list of k8s secrets via the Duplo API.
func (c *Client) K8SecretGetList(tenantID string) (*[]DuploK8sSecret, ClientError) {
	rp := []DuploK8sSecret{}
	err := c.getAPI(
		fmt.Sprintf("K8SecretGetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetAllK8Secrets", tenantID),
		&rp)

	// Add the tenant ID, then return the result.
	if err == nil {
		for i := range rp {
			rp[i].TenantID = tenantID
		}
	}

	return &rp, err
}
