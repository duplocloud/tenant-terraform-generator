package admininfra

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type PlanCertificate struct {
}

var CertPrefix = "certificates"

func (p PlanCertificate) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	// create new empty hcl file object
	infraName := config.DuploPlanId
	workingDir := filepath.Join(config.InfraDir, config.Infra)
	planCert, clientErr := client.PlanCertificateGetList(infraName)
	if clientErr != nil {
		return nil, errors.New(clientErr.Error())
	}
	if len(*planCert) == 0 {
		return nil, nil
	}
	tfContext := common.TFContext{}
	inputVars := generateCertVars(*planCert, CertPrefix)
	tfContext.InputVars = append(tfContext.InputVars, inputVars...)
	log.Println("Input vars plan_certificate ", inputVars)
	hclFile := hclwrite.NewEmptyFile()

	// create new file on system
	path := filepath.Join(workingDir, "plan.tf")
	tfFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	resourceName := common.GetResourceName(infraName)
	rootBody := hclFile.Body()
	// initialize the body of the new file object
	planBlock := rootBody.AppendNewBlock("resource",
		[]string{"duplocloud_plan_certificates", "plan_cert"})

	planBody := planBlock.Body()
	planBody.SetAttributeTraversal("plan_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "duplocloud_infrastructure",
		},
		hcl.TraverseAttr{
			Name: "infra",
		},
		hcl.TraverseAttr{
			Name: "infra_name",
		},
	})
	planBody.SetAttributeTraversal("depends_on", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "[duplocloud_infrastructure",
		},
		hcl.TraverseAttr{
			Name: "infra]",
		},
	})

	if planCert != nil && len(*planCert) > 0 {
		content := planBody.AppendNewBlock("dynamic", []string{"certificate"}).Body()
		certs := content.AppendNewBlock("content", nil).Body()
		certs.SetAttributeTraversal("name", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "certificate",
			},
			hcl.TraverseAttr{
				Name: "value.name",
			},
		})
		certs.SetAttributeTraversal("id", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "certificate",
			},
			hcl.TraverseAttr{
				Name: "value.id",
			},
		})
		content.SetAttributeTraversal("for_each", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "var",
			},
			hcl.TraverseAttr{
				Name: CertPrefix,
			},
		})
		//for _, v := range *planCert {
		//	cert := planBody.AppendNewBlock("certificate", nil).Body()
		//	cert.SetAttributeValue("name", cty.StringVal(v.CertificateName))
		//	cert.SetAttributeValue("id", cty.StringVal(v.CertificateArn))
		//}
	}
	rootBody.AppendNewline()
	_, err = tfFile.Write(hclFile.Bytes())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if config.GenerateTfState {
		importConfigs := []common.ImportConfig{}
		importConfigs = append(importConfigs, common.ImportConfig{
			ResourceAddress: "duplocloud_plan_certificate." + resourceName,
			ResourceId:      infraName,
			WorkingDir:      workingDir,
		})
		tfContext.ImportConfigs = importConfigs
	}

	return &tfContext, nil
}

type varCert struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func setVarCerts(duplo []duplosdk.DuploPlanCertificate) []varCert {
	ar := []varCert{}
	for _, data := range duplo {
		ar = append(ar, varCert{Id: data.CertificateArn, Name: data.CertificateName})
	}
	return ar
}
func generateCertVars(duplo []duplosdk.DuploPlanCertificate, prefix string) []common.VarConfig {
	varConfigs := make(map[string]common.VarConfig, 0)
	value := setVarCerts(duplo)
	certs, err := json.Marshal(&value)
	if err != nil {
		log.Fatal(err)
	}
	var1 := common.VarConfig{
		Name:       prefix,
		DefaultVal: string(certs),
		TypeVal: `list(object({
			id = string
			name = string
		  }))`,
	}
	varConfigs[prefix] = var1

	vars := make([]common.VarConfig, len(varConfigs))
	for _, v := range varConfigs {
		vars = append(vars, v)
	}

	return vars
}
