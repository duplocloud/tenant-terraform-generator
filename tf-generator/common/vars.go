package common

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
	TargetLocation string
	Vars           []VarConfig
}

func (v *Vars) Generate() {

	if len(v.Vars) > 0 {
		log.Println("[TRACE] <====== Variables TF generation started. =====>")

		// create new empty hcl file object
		hclFile := hclwrite.NewEmptyFile()
		// create new file on system
		path := filepath.Join(v.TargetLocation, "vars.tf")
		tfFile, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
			return
		}

		// initialize the body of the new file object
		rootBody := hclFile.Body()
		for _, varConfig := range v.Vars {
			if len(varConfig.Name) > 0 {
				varblock := rootBody.AppendNewBlock("variable",
					[]string{varConfig.Name})
				varBody := varblock.Body()
				if len(varConfig.DefaultVal) > 0 {
					if varConfig.DefaultVal == "null" {
						varBody.SetAttributeValue("default",
							cty.NullVal(cty.String))
					} else {
						if varConfig.TypeVal == "string" {
							varBody.SetAttributeValue("default",
								cty.StringVal(varConfig.DefaultVal))
						} else if varConfig.TypeVal == "number" || varConfig.TypeVal == "bool" {
							varBody.SetAttributeTraversal("default", hcl.Traversal{
								hcl.TraverseRoot{
									Name: varConfig.DefaultVal,
								},
							})
						}
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
			}

		}

		fmt.Printf("%s", hclFile.Bytes())
		_, err = tfFile.Write(hclFile.Bytes())
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Println("[TRACE] <====== Variables TF generation done. =====>")
	}
}
