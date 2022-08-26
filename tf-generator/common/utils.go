package common

import (
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

const (
	Host = iota
	S3
)

func Interpolate(body *hclwrite.Body, config Config, resourceName string, attrName string, resourceType int, dependsOnAttr string) {
	var duploResourceAddress string
	if resourceType == S3 {
		s3ShortName := resourceName
		prefix := "duploservices-" + config.TenantName + "-"
		if strings.HasPrefix(resourceName, prefix) {
			s3ShortName = resourceName[len(prefix):len(s3ShortName)]
			parts := strings.Split(s3ShortName, "-")
			if len(parts) > 0 {
				parts = parts[:len(parts)-1]
			}
			s3ShortName = strings.Join(parts, "-")
		}
		duploResourceAddress = "duplocloud_s3_bucket." + GetResourceName(s3ShortName)
	}
	body.SetAttributeTraversal(attrName, hcl.Traversal{
		hcl.TraverseRoot{
			Name: duploResourceAddress,
		},
		hcl.TraverseAttr{
			Name: dependsOnAttr,
		},
	})
}

func GetResourceName(name string) string {
	replacer := strings.NewReplacer("/", "_", "-", "_", ".", "_", " ", "_")
	return strings.ToLower(replacer.Replace(name))
}

func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
