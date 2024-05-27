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

type DuploPlanDnsConfig struct {
	DomainId          string `json:"DomainId,omitempty"`
	InternalDnsSuffix string `json:"InternalDnsSuffix,omitempty"`
	ExternalDnsSuffix string `json:"ExternalDnsSuffix,omitempty"`
	IsGlobalDNS       bool   `json:"IsGlobalDNS,omitempty"`
	IgnoreGlobalDNS   bool   `json:"IgnoreGlobalDNS,omitempty"`
}

type DuploPlanSettings struct {
	NwProvider            int    `json:"NwProvider,omitempty"` // FIXME: put a proper enum here.
	BlockBYOHosts         bool   `json:"BlockBYOHosts"`
	UnrestrictedExtLB     bool   `json:"UnrestrictedExtLB,omitempty"`
	InfraOwner            string `json:"InfraOwner"`
	DefaultApplicationUrl string `json:"DefaultApplicationUrl"`
}

func (c *Client) PlanCertificateGetList(planID string) (*[]DuploPlanCertificate, ClientError) {
	list := []DuploPlanCertificate{}
	err := c.getAPI("PlanCertificateGetList()", fmt.Sprintf("v3/admin/plans/%s/certificates", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) PlanGetSettings(planID string) (*DuploPlanSettings, ClientError) {
	settings := DuploPlanSettings{}
	err := c.getAPI(fmt.Sprintf("PlanGetSettings(%s)", planID), fmt.Sprintf("v3/admin/plans/%s/settings", planID), &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (c *Client) PlanGetDnsConfig(planID string) (*DuploPlanDnsConfig, ClientError) {
	dns := DuploPlanDnsConfig{}
	err := c.getAPI(fmt.Sprintf("PlanGetDnsConfig(%s)", planID), fmt.Sprintf("v3/admin/plans/%s/dnsConfig", planID), &dns)
	if err != nil {
		return nil, err
	}
	return &dns, nil
}

func (c *Client) PlanMetadataGetList(planID string) (*[]DuploKeyStringValue, ClientError) {
	list := []DuploKeyStringValue{}
	err := c.getAPI("PlanMetadataGetList()", fmt.Sprintf("v3/admin/plans/%s/metadata", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

type DuploPlanImage struct {
	Name     string                 `json:"Name"`
	ImageId  string                 `json:"ImageId,omitempty"`
	OS       string                 `json:"OS,omitempty"`
	Tags     *[]DuploKeyStringValue `json:"Tags,omitempty"`
	Username string                 `json:"Username,omitempty"`
}

func (c *Client) PlanImageGetList(planID string) (*[]DuploPlanImage, ClientError) {
	list := []DuploPlanImage{}
	err := c.getAPI(fmt.Sprintf("PlanImageGetList(%s)", planID), fmt.Sprintf("v3/admin/plans/%s/images", planID), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
