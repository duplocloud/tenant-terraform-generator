package duplosdk

import "fmt"

func (c *Client) PlanConfigGetList(planID string) (*[]DuploCustomDataEx, ClientError) {
	list := []DuploCustomDataEx{}
	err := c.getAPI("PlanConfigGetList()", fmt.Sprintf("v3/admin/plans/%s/configs", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

type PlanWAF struct {
	WebAclName   string
	WebAclId     string
	DashboardUrl string
}

func (c *Client) PlanWAFGetList(planID string) (*[]PlanWAF, ClientError) {
	list := []PlanWAF{}
	err := c.getAPI("PlanConfigGetList()", fmt.Sprintf("v3/admin/plans/%s/waf", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

type DuploPlanCertificate struct {
	CertificateName string `json:"CertificateName"`
	CertificateArn  string `json:"CertificateArn"`
}

func (c *Client) PlanCertificateGetList(planID string) (*[]DuploPlanCertificate, ClientError) {
	list := []DuploPlanCertificate{}
	err := c.getAPI("PlanCertificateGetList()", fmt.Sprintf("v3/admin/plans/%s/certificates", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
