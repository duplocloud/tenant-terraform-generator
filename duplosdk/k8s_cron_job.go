package duplosdk

import (
	"fmt"

	"k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DuploK8sCronJob represents a kubernetes job in a Duplo tenant
type DuploK8sCronJob struct {
	// NOTE: The TenantId field does not come from the backend - we synthesize it
	TenantId         string                `json:"-"` //nolint:govet
	Metadata         metav1.ObjectMeta     `json:"metadata"`
	Spec             v1beta1.CronJobSpec   `json:"spec"`
	Status           v1beta1.CronJobStatus `json:"status"`
	IsAnyHostAllowed bool                  `json:"IsAnyHostAllowed"`
}

// K8sCronJobGetList retrieves a list of k8s jobs via the Duplo API.
func (c *Client) K8sCronJobGetList(tenantId string) (*[]DuploK8sCronJob, ClientError) {
	var rp []DuploK8sCronJob
	err := c.getAPI(
		fmt.Sprintf("k8sCronJobGetList(%s)", tenantId),
		fmt.Sprintf("/v3/subscriptions/%s/k8s/cronjob", tenantId),
		&rp)

	// Add the tenant ID, then return the result.
	if err == nil {
		for i := range rp {
			rp[i].TenantId = tenantId
		}
	}

	return &rp, err
}
