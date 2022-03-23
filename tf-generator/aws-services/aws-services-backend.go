package awsservices

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type AwsServicesBackend struct {
}

func (asb *AwsServicesBackend) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	log.Println("[TRACE] <====== AWS Services backend TF generation started. =====>")
	// create new empty hcl file object
	hclFile := hclwrite.NewEmptyFile()

	// create new file on system
	path := filepath.Join(config.TFCodePath, config.AwsServicesProject, "backend.tf")
	tfFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// initialize the body of the new file object
	rootBody := hclFile.Body()

	// Add duplo terraform block
	tfBlock := rootBody.AppendNewBlock("terraform",
		nil)
	tfBlockBody := tfBlock.Body()
	s3Backend := tfBlockBody.AppendNewBlock("backend",
		[]string{"s3"})

	s3BackendBody := s3Backend.Body()

	s3BackendBody.SetAttributeValue("region",
		cty.StringVal("us-west-2")) // TODO - Take region from ENV VAR
	s3BackendBody.SetAttributeValue("key",
		cty.StringVal(config.AwsServicesProject))

	s3BackendBody.SetAttributeValue("workspace_key_prefix",
		cty.StringVal("tenant:"))
	s3BackendBody.SetAttributeValue("encrypt",
		cty.True)

	fmt.Printf("%s", hclFile.Bytes())
	tfFile.Write(hclFile.Bytes())
	log.Println("[TRACE] <====== AWS Services backend TF generation done. =====>")
	return nil, nil
}
