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

type OutputVarConfig struct {
	Name          string
	ActualVal     string
	DescVal       string
	RootTraversal bool
}

type OutputVars struct {
	TargetLocation string
	OutputVars     []OutputVarConfig
}

func (ov *OutputVars) Generate() {
	if len(ov.OutputVars) > 0 {
		log.Println("[TRACE] <====== Output Variables TF generation started. =====>")

		// create new empty hcl file object
		hclFile := hclwrite.NewEmptyFile()
		// create new file on system
		path := filepath.Join(ov.TargetLocation, "outputs.tf")
		tfFile, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
			return
		}

		// initialize the body of the new file object
		rootBody := hclFile.Body()
		for _, outVarConfig := range ov.OutputVars {
			if len(outVarConfig.Name) > 0 {
				outputVarblock := rootBody.AppendNewBlock("output",
					[]string{outVarConfig.Name})
				outputVarBody := outputVarblock.Body()
				if len(outVarConfig.ActualVal) > 0 {
					if outVarConfig.RootTraversal {
						outputVarBody.SetAttributeTraversal("value", hcl.Traversal{
							hcl.TraverseRoot{
								Name: outVarConfig.ActualVal,
							},
						})
					} else {
						outputVarBody.SetAttributeValue("value",
							cty.StringVal(outVarConfig.ActualVal))
					}
				}

				if len(outVarConfig.DescVal) > 0 {
					outputVarBody.SetAttributeValue("description",
						cty.StringVal(outVarConfig.DescVal))
				}
			}
		}

		fmt.Printf("%s", hclFile.Bytes())
		_, err = tfFile.Write(hclFile.Bytes())
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Println("[TRACE] <====== Output Variables TF generation done. =====>")
	}
}
