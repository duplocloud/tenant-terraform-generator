package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	tfgenerator "tenant-terraform-generator/tf-generator"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type AppMain struct {
}

func (am *AppMain) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AppProject)

	log.Println("[TRACE] <====== App services main TF generation started. =====>")

	//1. ==========================================================================================
	// Generate locals
	hclFile := hclwrite.NewEmptyFile()

	// create new file on system
	path := filepath.Join(workingDir, "main.tf")
	tfFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// initialize the body of the new file object
	rootBody := hclFile.Body()

	awsCallerIdBlock := rootBody.AppendNewBlock("data",
		[]string{"aws_caller_identity",
			"current"})
	awsCallerIdBody := awsCallerIdBlock.Body()
	awsCallerIdBody.Clear()
	rootBody.AppendNewline()

	regionBlock := rootBody.AppendNewBlock("data",
		[]string{"aws_region",
			"current"})
	regionBody := regionBlock.Body()
	regionBody.Clear()
	rootBody.AppendNewline()

	localsBlock := rootBody.AppendNewBlock("locals",
		nil)
	localsBlockBody := localsBlock.Body()
	tfstateBucketTokens := hclwrite.Tokens{
		{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
		{Type: hclsyntax.TokenIdent, Bytes: []byte(`duplo-tfstate-${data.aws_caller_identity.current.account_id}`)},
		{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
	}
	localsBlockBody.SetAttributeRaw("tfstate_bucket", tfstateBucketTokens)

	localsBlockBody.SetAttributeTraversal("region", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "var",
		},
		hcl.TraverseAttr{
			Name: "region",
		},
	})
	localsBlockBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "data.terraform_remote_state",
		},
		hcl.TraverseAttr{
			Name: "tenant.outputs[\"tenant_id\"]",
		},
	})
	localsBlockBody.SetAttributeTraversal("cert_arn", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "data.terraform_remote_state",
		},
		hcl.TraverseAttr{
			Name: "tenant.outputs[\"cert_arn\"]",
		},
	})
	localsBlockBody.SetAttributeTraversal("tenant_name", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "data.terraform_remote_state",
		},
		hcl.TraverseAttr{
			Name: "tenant.outputs[\"tenant_name\"]",
		},
	})
	rootBody.AppendNewline()

	remoteStateBlock := rootBody.AppendNewBlock("data",
		[]string{"terraform_remote_state",
			"tenant"})
	remoteStateBody := remoteStateBlock.Body()
	remoteStateBody.SetAttributeValue("backend",
		cty.StringVal("s3"))
	remoteStateBody.SetAttributeTraversal("workspace", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "terraform",
		},
		hcl.TraverseAttr{
			Name: "workspace",
		},
	})

	//configMap := map[string]cty.Value{}
	// tokens := hclwrite.Tokens{
	// 	//{Type: hclsyntax.TokenOQuote, Bytes: []byte(`"`)},
	// 	//{Type: hclsyntax.TokenTemplateInterp, Bytes: []byte(`${`)},
	// 	{Type: hclsyntax.TokenIdent, Bytes: []byte(`local.tfstate_bucket`)},
	// 	//{Type: hclsyntax.TokenTemplateSeqEnd, Bytes: []byte(`}`)},
	// 	//{Type: hclsyntax.TokenCQuote, Bytes: []byte(`"`)},
	// }
	//remoteStateBody.SetAttributeRaw("bucket", tokens)

	// configTokens := map[string]hclwrite.Tokens{
	// 	"bucket": hclwrite.TokensForTraversal(hcl.Traversal{
	// 		hcl.TraverseRoot{Name: "local"},
	// 		hcl.TraverseAttr{Name: "tfstate_bucket"},
	// 	}),
	// 	"workspace_key_prefix": hclwrite.TokensForValue(cty.StringVal("admin:")),
	// 	"key":                  hclwrite.TokensForValue(cty.StringVal("tenant")),
	// 	"region": hclwrite.TokensForTraversal(hcl.Traversal{
	// 		hcl.TraverseRoot{Name: "local"},
	// 		hcl.TraverseAttr{Name: "region"},
	// 	}),
	// }
	configTokens := []tfgenerator.ObjectAttrTokens{
		{
			Name: hclwrite.TokensForTraversal(hcl.Traversal{
				hcl.TraverseRoot{Name: "bucket"},
			}),
			Value: hclwrite.TokensForTraversal(hcl.Traversal{
				hcl.TraverseRoot{Name: "local"},
				hcl.TraverseAttr{Name: "tfstate_bucket"},
			}),
		},
		{
			Name: hclwrite.TokensForTraversal(hcl.Traversal{
				hcl.TraverseRoot{Name: "workspace_key_prefix"},
			}),
			Value: hclwrite.TokensForValue(cty.StringVal("admin:")),
		},
		{
			Name: hclwrite.TokensForTraversal(hcl.Traversal{
				hcl.TraverseRoot{Name: "key"},
			}),
			Value: hclwrite.TokensForValue(cty.StringVal("tenant")),
		},
		{
			Name: hclwrite.TokensForTraversal(hcl.Traversal{
				hcl.TraverseRoot{Name: "region"},
			}),
			Value: hclwrite.TokensForTraversal(hcl.Traversal{
				hcl.TraverseRoot{Name: "local"},
				hcl.TraverseAttr{Name: "region"},
			}),
		},
	}
	tokens := tfgenerator.TokensForObject(configTokens)
	remoteStateBody.SetAttributeRaw("config", tokens)
	// 	cty.ObjectVal(configMap))

	// configMap["bucket"] = cty.StringVal("${local.tfstate_bucket}")
	// configMap["workspace_key_prefix"] = cty.StringVal("admin:")
	// configMap["key"] = cty.StringVal("tenant")
	// configMap["region"] = cty.StringVal("${local.region}")

	// //configMap["region"] = cty.CapsuleVal("${local.region}")
	// remoteStateBody.SetAttributeValue("config",
	// 	cty.ObjectVal(configMap))

	//fmt.Printf("%s", hclFile.Bytes())
	_, err = tfFile.Write(hclFile.Bytes())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	log.Println("[TRACE] <====== Aws services main TF generation done. =====>")
	return &common.TFContext{
		InputVars: generateVars(),
	}, nil
}

func generateVars() []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig)

	regionVar := common.VarConfig{
		Name:       "region",
		DefaultVal: "us-west-2",
		TypeVal:    "string",
	}
	varConfigs["region"] = regionVar

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	return vars
}
