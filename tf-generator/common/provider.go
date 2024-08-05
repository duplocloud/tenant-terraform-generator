package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"

	"tenant-terraform-generator/duplosdk"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/zclconf/go-cty/cty"
)

type Provider struct {
}

func (p *Provider) Generate(config *Config, client *duplosdk.Client) {
	log.Println("[TRACE] <====== Provider TF generation started. =====>")
	log.Printf("Config - %s", fmt.Sprintf("%#v", config))
	// create new empty hcl file object
	hclFile := hclwrite.NewEmptyFile()

	// create new file on system
	tenantProject := filepath.Join(config.TFCodePath, config.TenantProject, "providers.tf")
	awsServicesProject := filepath.Join(config.TFCodePath, config.AwsServicesProject, "providers.tf")
	appProject := filepath.Join(config.TFCodePath, config.AppProject, "providers.tf")
	infraProject := filepath.Join(config.TFCodePath, config.InfraProject, "providers.tf")

	tenantProjectFile, err := os.Create(tenantProject)
	if err != nil {
		fmt.Println(err)
		return
	}
	awsServicesProjectFile, err := os.Create(awsServicesProject)
	if err != nil {
		fmt.Println(err)
		return
	}

	appProjectFile, err := os.Create(appProject)
	if err != nil {
		fmt.Println(err)
		return
	}

	infraProjectFile, err := os.Create(infraProject)
	if err != nil {
		fmt.Println(err)
		return
	}
	// initialize the body of the new file object
	rootBody := hclFile.Body()

	// Add duplo terraform block
	tfBlock := rootBody.AppendNewBlock("terraform",
		nil)
	tfBlockBody := tfBlock.Body()
	tfVersion := GetEnv("tf_version", TF_DEFAULT_VERSION)
	tfBlockBody.SetAttributeValue("required_version",
		cty.StringVal(">= "+tfVersion))

	reqProvsBlock := tfBlockBody.AppendNewBlock("required_providers",
		nil)
	reqProvsBlockBody := reqProvsBlock.Body()

	reqProvsBlockBody.SetAttributeValue("duplocloud",
		cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal("duplocloud/duplocloud"),
			"version": cty.StringVal("~> " + config.DuploProviderVersion),
		}))

	// Add duplo provider block
	duploProvider := rootBody.AppendNewBlock("provider",
		[]string{"duplocloud"})

	duploProviderBody := duploProvider.Body()
	// providerBody.SetAttributeValue("duplo_host",
	// 	cty.StringVal(client.HostURL))
	// providerBody.SetAttributeValue("duplo_token",
	// 	cty.StringVal(client.Token))
	duploProviderBody.AppendNewline()

	awsProvider := rootBody.AppendNewBlock("provider",
		[]string{"aws"})
	awsProviderBody := awsProvider.Body()
	awsProviderBody.SetAttributeTraversal("region", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "region",
		},
	})
	awsProviderBody.AppendNewline()

	fmt.Printf("%s", hclFile.Bytes())
	_, err = tenantProjectFile.Write(hclFile.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = appProjectFile.Write(hclFile.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = infraProjectFile.Write(hclFile.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
	reqProvsBlockBody.SetAttributeValue("random",
		cty.ObjectVal(map[string]cty.Value{
			"source":  cty.StringVal("hashicorp/random"),
			"version": cty.StringVal("~> 3.3.2"),
		}))
	randomProvider := rootBody.AppendNewBlock("provider",
		[]string{"random"})

	randomProviderBody := randomProvider.Body()
	randomProviderBody.AppendNewline()

	_, err = awsServicesProjectFile.Write(hclFile.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println("[TRACE] <====== Provider TF generation done. =====>")
}
