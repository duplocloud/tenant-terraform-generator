package duplosdk

import (
	"fmt"
	"log"
)

type DuploAsgProfile struct {
	MinSize           int                                `json:"MinSize"`
	MaxSize           int                                `json:"MaxSize"`
	DesiredCapacity   int                                `json:"DesiredCapacity"`
	AccountName       string                             `json:"AccountName,omitempty"`
	TenantId          string                             `json:"TenantId,omitempty"`
	FriendlyName      string                             `json:"FriendlyName,omitempty"`
	Capacity          string                             `json:"Capacity,omitempty"`
	Zone              int                                `json:"Zone"`
	IsMinion          bool                               `json:"IsMinion"`
	ImageID           string                             `json:"ImageId,omitempty"`
	Base64UserData    string                             `json:"Base64UserData,omitempty"`
	AgentPlatform     int                                `json:"AgentPlatform"`
	IsEbsOptimized    bool                               `json:"IsEbsOptimized"`
	AllocatedPublicIP bool                               `json:"AllocatedPublicIp,omitempty"`
	Cloud             int                                `json:"Cloud"`
	EncryptDisk       bool                               `json:"EncryptDisk,omitempty"`
	Status            string                             `json:"Status,omitempty"`
	NetworkInterfaces *[]DuploNativeHostNetworkInterface `json:"NetworkInterfaces,omitempty"`
	Volumes           *[]DuploNativeHostVolume           `json:"Volumes,omitempty"`
	MetaData          *[]DuploKeyStringValue             `json:"MetaData,omitempty"`
	Tags              *[]DuploKeyStringValue             `json:"Tags,omitempty"`
	MinionTags        *[]DuploKeyStringValue             `json:"MinionTags,omitempty"`
	CustomDataTags    *[]DuploKeyStringValue             `json:"CustomDataTags,omitempty"`
	KeyPairType       int                                `json:"KeyPairType,omitempty"`
	MaxSpotPrice      string                             `json:"SpotPrice,omitempty"`
	UseSpotInstances  bool                               `json:"UseSpotInstances"`
	Taints            *[]DuploTaints                     `json:"Taints,omitempty"`
}

type DuploAsgProfileDeleteReq struct {
	FriendlyName string `json:"FriendlyName,omitempty"`
	State        string `json:"State,omitempty"`
}

// AsgProfileGetList retrieves a list of ASG profiles via the Duplo API.
func (c *Client) AsgProfileGetList(tenantID string) (*[]DuploAsgProfile, ClientError) {
	log.Printf("[DEBUG] Duplo API - Get ASG Profile List(TenantId-%s)", tenantID)
	rp := []DuploAsgProfile{}
	err := c.getAPI(fmt.Sprintf("AsgProfileGetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetTenantAsgProfiles", tenantID),
		&rp)
	return &rp, err
}
