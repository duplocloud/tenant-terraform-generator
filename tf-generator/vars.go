package tfgenerator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type VarConfig struct {
	Name       string
	TypeVal    string
	DefaultVal string
	DescVal    string
}

type Vars struct {
	targetLocation string
	vars           []VarConfig
}

func (v *Vars) Generate() {
	log.Println("[TRACE] <====== Variables TF generation started. =====>")

	// create new empty hcl file object
	hclFile := hclwrite.NewEmptyFile()
	// create new file on system
	path := filepath.Join(v.targetLocation, "vars.tf")
	tfFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	// initialize the body of the new file object
	rootBody := hclFile.Body()
	for _, varConfig := range v.vars {
		if len(varConfig.Name) > 0 {
			varblock := rootBody.AppendNewBlock("variable",
				[]string{varConfig.Name})
			varBody := varblock.Body()
			if len(varConfig.DefaultVal) > 0 {
				if varConfig.DefaultVal == "null" {
					varBody.SetAttributeValue("default",
						cty.NullVal(cty.String))
				} else {
					varBody.SetAttributeValue("default",
						cty.StringVal(varConfig.DefaultVal))
				}

			}

			if len(varConfig.DescVal) > 0 {
				varBody.SetAttributeValue("description",
					cty.StringVal(varConfig.DescVal))
			}

			varBody.SetAttributeTraversal("type", hcl.Traversal{
				hcl.TraverseRoot{
					Name: varConfig.TypeVal,
				},
			})
			rootBody.AppendNewline()
		}

	}

	fmt.Printf("%s", hclFile.Bytes())
	tfFile.Write(hclFile.Bytes())
	log.Println("[TRACE] <====== Variables TF generation done. =====>")
}
