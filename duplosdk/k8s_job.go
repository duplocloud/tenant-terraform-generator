package duplosdk

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DuploK8sCronJob represents a kubernetes job in a Duplo tenant
type DuploK8sJob struct {
	// NOTE: The TenantId field does not come from the backend - we synthesize it
	TenantId         string            `json:"-"` //nolint:govet
	Metadata         metav1.ObjectMeta `json:"metadata"`
	Spec             batchv1.JobSpec   `json:"spec"`
	Status           batchv1.JobStatus `json:"status"`
	IsAnyHostAllowed bool              `json:"IsAnyHostAllowed"`
}

// K8sJobGetList retrieves a list of k8s jobs via the Duplo API.
func (c *Client) K8sJobGetList(tenantId string) (*[]DuploK8sJob, ClientError) {
	var rp []DuploK8sJob
	err := c.getAPI(
		fmt.Sprintf("k8sJobGetList(%s)", tenantId),
		fmt.Sprintf("v3/subscriptions/%s/k8s/job", tenantId),
		&rp)

	// Add the tenant ID, then return the result.
	if err == nil {
		for i := range rp {
			rp[i].TenantId = tenantId
		}
	}

	return &rp, err
}
