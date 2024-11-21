package duplosdk

import (
	"fmt"
)

const (
	KeyPairType_None    = 0
	KeyPairType_RSA     = 1
	KeyPairType_ED25519 = 2
)

// DuploNativeHost is a Duplo SDK object that represents an nativehost
type DuploNativeHost struct {
	InstanceID         string                             `json:"InstanceId"`
	UserAccount        string                             `json:"UserAccount,omitempty"`
	TenantID           string                             `json:"TenantId,omitempty"`
	FriendlyName       string                             `json:"FriendlyName,omitempty"`
	Capacity           string                             `json:"Capacity,omitempty"`
	Zone               int                                `json:"Zone"`
	IsMinion           bool                               `json:"IsMinion"`
	ImageID            string                             `json:"ImageId,omitempty"`
	Base64UserData     string                             `json:"Base64UserData,omitempty"`
	AgentPlatform      int                                `json:"AgentPlatform"`
	IsEbsOptimized     bool                               `json:"IsEbsOptimized"`
	AllocatedPublicIP  bool                               `json:"AllocatedPublicIp,omitempty"`
	Cloud              int                                `json:"Cloud"`
	KeyPairType        int                                `json:"KeyPairType"`
	EncryptDisk        bool                               `json:"EncryptDisk,omitempty"`
	Status             string                             `json:"Status,omitempty"`
	IdentityRole       string                             `json:"IdentityRole,omitempty"`
	PrivateIPAddress   string                             `json:"PrivateIpAddress,omitempty"`
	NetworkInterfaceId string                             `json:"NetworkInterfaceId,omitempty"`
	NetworkInterfaces  *[]DuploNativeHostNetworkInterface `json:"NetworkInterfaces,omitempty"`
	Volumes            *[]DuploNativeHostVolume           `json:"Volumes,omitempty"`
	MetaData           *[]DuploKeyStringValue             `json:"MetaData,omitempty"`
	Tags               *[]DuploKeyStringValue             `json:"Tags,omitempty"`
	TagsEx             *[]DuploKeyStringValue             `json:"TagsEx,omitempty"`
	MinionTags         *[]DuploKeyStringValue             `json:"MinionTags,omitempty"`
	Taints             *[]DuploTaints                     `json:"Taints,omitempty"`
}

type DuploTaints struct {
	Key    string `json:"Key"`
	Value  string `json:"Value"`
	Effect string `json:"Effect"`
}

// DuploNativeHostNetworkInterface is a Duplo SDK object that represents a network interface of a native host
type DuploNativeHostNetworkInterface struct {
	NetworkInterfaceID string                 `json:"NetworkInterfaceId,omitempty"`
	SubnetID           string                 `json:"SubnetId,omitempty"`
	AssociatePublicIP  bool                   `json:"AssociatePublicIpAddress,omitempty"`
	Groups             *[]string              `json:"Groups,omitempty"`
	DeviceIndex        int                    `json:"DeviceIndex,omitempty"`
	MetaData           *[]DuploKeyStringValue `json:"MetaData,omitempty"`
}

// DuploNativeHostVolume is a Duplo SDK object that represents a volume of a native host
type DuploNativeHostVolume struct {
	Iops       int    `json:"Iops,omitempty"`
	Name       string `json:"Name,omitempty"`
	Size       int    `Size:"Size,omitempty"`
	VolumeID   string `json:"VolumeId,omitempty"`
	VolumeType string `json:"VolumeType,omitempty"`
}

// NativeHostGetList retrieves a list of native hosts via the Duplo API.
func (c *Client) NativeHostGetList(tenantID string) (*[]DuploNativeHost, ClientError) {
	rp := []DuploNativeHost{}
	err := c.getAPI(fmt.Sprintf("NativeHostGetList(%s)", tenantID),
		fmt.Sprintf("v2/subscriptions/%s/NativeHostV2", tenantID),
		&rp)
	return &rp, err
}

func (c *Client) GetMinionForHost(tenantID, name string) (*DuploMinion, ClientError) {
	list, err := c.TenantListMinions(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, minion := range *list {
			if minion.Name == name {
				return &minion, nil
			}
		}
	}
	return nil, nil
}
