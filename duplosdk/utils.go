package duplosdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
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

// handle path parameter encoding when it might contain slashes
func EncodePathParam(param string) string {
	return url.PathEscape(url.PathEscape(param))
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

func JSONMarshal(t interface{}) (string, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(t)
	return buffer.String(), err
}

func SelectKeyValues(metadata *[]DuploKeyStringValue, keys []string) *[]DuploKeyStringValue {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return SelectKeyValuesFromMap(metadata, specified)
}

func SelectKeyValuesFromMap(metadata *[]DuploKeyStringValue, keys map[string]interface{}) *[]DuploKeyStringValue {
	settings := make([]DuploKeyStringValue, 0, len(keys))
	for _, kv := range *metadata {
		if _, ok := keys[kv.Key]; ok {
			settings = append(settings, kv)
		}
	}

	return &settings
}

func CopyDirectory(scrDir, dest string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
