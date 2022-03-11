package duplosdk

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

//nolint:deadcode,unused // utility function
func isInterfaceNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

// GetDuploServicesNameWithAws builds a duplo resource name, given a tenant ID. The name includes the AWS account ID suffix.
func (c *Client) GetDuploServicesNameWithAws(tenantID, name string) (string, ClientError) {
	return c.GetResourceName("duploservices", tenantID, name, true)
}

// GetDuploServicesName builds a duplo resource name, given a tenant ID.
func (c *Client) GetDuploServicesName(tenantID, name string) (string, ClientError) {
	return c.GetResourceName("duploservices", tenantID, name, false)
}

// GetResourceName builds a duplo resource name, given a tenant ID.  It can optionally include the AWS account ID suffix.
func (c *Client) GetResourceName(prefix, tenantID, name string, withAccountSuffix bool) (string, ClientError) {
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return "", err
	}
	if withAccountSuffix {
		accountID, err := c.TenantGetAwsAccountID(tenantID)
		if err != nil {
			return "", err
		}
		return strings.Join([]string{prefix, tenant.AccountName, name, accountID}, "-"), nil
	}
	return strings.Join([]string{prefix, tenant.AccountName, name}, "-"), nil
}

// GetDuploServicesPrefix builds a duplo resource name, given a tenant ID.
func (c *Client) GetDuploServicesPrefix(tenantID string) (string, ClientError) {
	return c.GetResourcePrefix("duploservices", tenantID)
}

// GetResourcePrefix builds a duplo resource prefix, given a tenant ID.
func (c *Client) GetResourcePrefix(prefix, tenantID string) (string, ClientError) {
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{prefix, tenant.AccountName}, "-"), nil
}

// UnprefixName removes a duplo resource prefix from a name.
func UnprefixName(prefix, name string) (string, bool) {
	if strings.HasPrefix(name, prefix) {
		return name[len(prefix)+1:], true
	}

	return name, false
}

// UnwrapName removes a duplo resource prefix and AWS account ID suffix from a name.
func UnwrapName(prefix, accountID, name string) (string, bool) {
	suffix := "-" + accountID

	if !strings.HasSuffix(name, suffix) {
		return name, false
	}

	part := name[0 : len(name)-len(suffix)]
	if !strings.HasPrefix(part, prefix) {
		return name, false
	}

	return part[len(prefix)+1:], true
}

func PrettyStruct(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func UnmarshalEscapedJson(escapedString string) (string, error) {
	var val []byte = []byte(escapedString)
	var data map[string]interface{}

	s, _ := strconv.Unquote(string(val))

	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		return "", err
	}
	return s, nil
}
