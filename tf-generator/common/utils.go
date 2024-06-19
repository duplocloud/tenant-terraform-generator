package common

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type ObjectAttrTokens struct {
	Name  hclwrite.Tokens
	Value hclwrite.Tokens
}

func TokensForObject(attrs []ObjectAttrTokens) hclwrite.Tokens {
	var toks hclwrite.Tokens
	toks = append(toks, &hclwrite.Token{
		Type:  hclsyntax.TokenOBrace,
		Bytes: []byte{'{'},
	})
	if len(attrs) > 0 {
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		})
	}
	for _, attr := range attrs {
		toks = append(toks, attr.Name...)
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenEqual,
			Bytes: []byte{'='},
		})
		toks = append(toks, attr.Value...)
		toks = append(toks, &hclwrite.Token{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte{'\n'},
		})
	}
	toks = append(toks, &hclwrite.Token{
		Type:  hclsyntax.TokenCBrace,
		Bytes: []byte{'}'},
	})

	//format(toks) // fiddle with the SpacesBefore field to get canonical spacing
	return toks
}

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
	if name[0] >= '0' && name[0] <= '9' {
		name = "duplo_" + name
	}
	return strings.ToLower(replacer.Replace(name))
}

func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func RepalceStringInFile(file string, stringsToRepalce map[string]string) {
	input, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(err)
	}
	newStr := string(input)
	for key, element := range stringsToRepalce {
		newStr = strings.Replace(newStr, key, element, -1)
	}

	err = ioutil.WriteFile(file, []byte(newStr), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func ValidateAndFormatTfCode(tfDir, tfVersion string) {
	log.Printf("[TRACE] Validation and formatting of terraform code generated at %s is started.", tfDir)
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(tfVersion)),
		//Version: version.NewConstraint(">= 1.0, < 1.4"),
	}
	// constraint, _ := version.NewConstraint(">= 1.2.8")
	// installer := &releases.Versions{
	// 	Product:     product.Terraform,
	// 	Constraints: constraint,
	// }

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
	tf, err := tfexec.NewTerraform(tfDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}
	log.Printf("[TRACE] Validation of terraform code generated at %s is started.", tfDir)
	_, err = tf.Validate(context.Background())
	if err != nil {
		log.Fatalf("error running terraform validate: %s", err)
	}
	log.Printf("[TRACE] Validation of terraform code generated at %s is done.", tfDir)
	log.Printf("[TRACE] Formatting of terraform code generated at %s is started.", tfDir)
	err = tf.FormatWrite(context.Background())
	if err != nil {
		log.Fatalf("error running terraform format: %s", err)
	}
	log.Printf("[TRACE] Formatting of terraform code generated at %s is done.", tfDir)
	log.Printf("[TRACE] Validation and formatting of terraform code generated at %s is done.", tfDir)
}

func ExtractResourceNameFromArn(arn string) string {
	var validID = regexp.MustCompile(`[^:/]*$`)
	return validID.FindString(arn)
}

func MakeMapUpperCamelCase(m map[string]interface{}) {
	for k := range m {

		// Only convert lowercase entries.
		if unicode.IsLower([]rune(k)[0]) {
			//nolint:staticcheck // SA1019 ignore this!
			upper := strings.Title(k)

			// Add the upper camel-case entry, if it doesn't exist.
			if _, ok := m[upper]; !ok {
				m[upper] = m[k]
			}

			// Remove the lower camel-case entry.
			delete(m, k)
		}
	}
}

// Internal function to reduce empty or nil map entries.
func ReduceNilOrEmptyMapEntries(m map[string]interface{}) {
	for k, v := range m {
		if IsInterfaceNil(v) || IsInterfaceNilOrEmptyMap(v) || IsInterfaceNilOrEmptySlice(v) {
			delete(m, k)
		}
	}
}

func IsInterfaceNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

func IsInterfaceNilOrEmptySlice(v interface{}) bool {
	if IsInterfaceNil(v) {
		return true
	}

	slice := reflect.ValueOf(v)
	return slice.Kind() == reflect.Slice && slice.IsValid() && (slice.IsNil() || slice.Len() == 0)
}

func IsInterfaceEmptySlice(v interface{}) bool {
	slice := reflect.ValueOf(v)
	return slice.Kind() == reflect.Slice && slice.IsValid() && !slice.IsNil() && slice.Len() == 0
}

//nolint:deadcode,unused // utility function
func IsInterfaceEmptyMap(v interface{}) bool {
	emap := reflect.ValueOf(v)
	return emap.Kind() == reflect.Map && emap.IsValid() && !emap.IsNil() && emap.Len() == 0
}

func IsInterfaceNilOrEmptyMap(v interface{}) bool {
	if IsInterfaceNil(v) {
		return true
	}

	emap := reflect.ValueOf(v)
	return emap.Kind() == reflect.Map && emap.IsValid() && (emap.IsNil() || emap.Len() == 0)
}

func GetDuploManagedAwsTags() []string {
	return []string{"duplo-project", "TENANT_NAME"}
}

func StringSliceToListVal(ele []string) []cty.Value {
	var ctyValues []cty.Value
	for _, v := range ele {
		ctyValues = append(ctyValues, cty.StringVal(v))
	}
	return ctyValues
}

func Int64SliceToListVal(ele []int64) []cty.Value {
	var ctyValues []cty.Value
	for _, v := range ele {
		ctyValues = append(ctyValues, cty.NumberIntVal(v))
	}
	return ctyValues
}

func ConstructConfigVars(vars []VarConfig) map[string]interface{} {
	m := make(map[string]interface{})
	for _, v := range vars {
		if v.Name != "" {
			m[v.Name] = v.DefaultVal
		}
	}
	return m
}
