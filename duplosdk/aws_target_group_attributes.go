package duplosdk

import (
	"fmt"
)

type DuploTargetGroupAttributes struct {
	Attributes     *[]DuploKeyStringValue `json:"Attributes,omitempty"`
	IsEcsLB        bool                   `json:"IsEcsLB,omitempty"`
	IsPassThruLB   bool                   `json:"IsPassThruLB,omitempty"`
	Port           int                    `json:"Port,omitempty"`
	RoleName       string                 `json:"RoleName,omitempty"`
	TargetGroupArn string                 `json:"TargetGroupArn,omitempty"`
}

type DuploTargetGroupAttributesGetReq struct {
	IsEcsLB        bool   `json:"IsEcsLB,omitempty"`
	IsPassThruLB   bool   `json:"IsPassThruLB,omitempty"`
	Port           int    `json:"Port,omitempty"`
	RoleName       string `json:"RoleName,omitempty"`
	TargetGroupArn string `json:"TargetGroupArn,omitempty"`
}

func (c *Client) DuploAwsTargetGroupAttributesGet(tenantID string, rq DuploTargetGroupAttributesGetReq) (*[]DuploKeyStringValue, ClientError) {
	rp := []DuploKeyStringValue{}
	err := c.postAPI(
		fmt.Sprintf("TargetGroupAttributesGet(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/targetGroupAttributes", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}
