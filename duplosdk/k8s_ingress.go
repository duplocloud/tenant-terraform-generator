package duplosdk

import "fmt"

type DuploK8sIngress struct {
	Name             string                 `json:"name"`
	IngressClassName string                 `json:"ingressClassName"`
	Annotations      map[string]string      `json:"annotations,omitempty"`
	Labels           map[string]string      `json:"labels,omitempty"`
	LbConfig         *DuploK8sLbConfig      `json:"lbConfig,omitempty"`
	Rules            *[]DuploK8sIngressRule `json:"rules,omitempty"`
}

type DuploK8sLbConfig struct {
	IsPublic          bool                      `json:"isPublic,omitempty"`
	DnsPrefix         string                    `json:"dnsPrefix,omitempty"`
	WafArn            string                    `json:"wafArn,omitempty"`
	EnableAccessLogs  bool                      `json:"enableAccessLogs,omitempty"`
	DropInvalidHeader bool                      `json:"dropInvalidHeader,omitempty"`
	CertArn           string                    `json:"certArn,omitempty"`
	Listeners         *DuploK8sIngressListeners `json:"listeners,omitempty"`
}

type DuploK8sIngressListeners struct {
	Http  []int `json:"http,omitempty"`
	Https []int `json:"https,omitempty"`
	Tcp   []int `json:"tcp,omitempty"`
}

type DuploK8sIngressRule struct {
	Path        string `json:"path,omitempty"`
	PathType    string `json:"pathType,omitempty"`
	ServiceName string `json:"serviceName,omitempty"`
	Host        string `json:"host,omitempty"`
	Port        int    `json:"port,omitempty"`
}

func (c *Client) DuploK8sIngressGetList(tenantID string) (*[]DuploK8sIngress, ClientError) {
	rp := []DuploK8sIngress{}
	err := c.getAPI(
		fmt.Sprintf("DuploK8sIngressGet(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/k8s/ingress", tenantID),
		&rp,
	)
	return &rp, err
}
