package app

import (
	"fmt"
	"os"
	"path/filepath"
	"tenant-terraform-generator/duplosdk"
	"tenant-terraform-generator/tf-generator/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type K8sJob struct {
}

func (k8sJob *K8sJob) Generate(config *common.Config, client *duplosdk.Client) (*common.TFContext, error) {
	workingDir := filepath.Join(config.TFCodePath, config.AppProject)
	list, clientErr := client.K8sJobGetList(config.TenantId)
	if clientErr != nil {
		fmt.Println(clientErr)
		return nil, nil
	}
	fmt.Println("List \n\n ", list)
	tfContext := common.TFContext{}
	importConfigs := []common.ImportConfig{}

	for _, d := range *list {
		hclFile := hclwrite.NewEmptyFile()

		path := filepath.Join(workingDir, "k8s-job-"+d.Metadata.Name+".tf")
		tfFile, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		resourceName := common.GetResourceName(d.Metadata.Name)
		rootBody := hclFile.Body()
		cronJobBlock := rootBody.AppendNewBlock("resource",
			[]string{"duplocloud_k8s_cron_job",
				resourceName})
		cronjobBody := cronJobBlock.Body()
		cronjobBody.SetAttributeTraversal("tenant_id", hcl.Traversal{
			hcl.TraverseRoot{
				Name: "local",
			},
			hcl.TraverseAttr{
				Name: "tenant_id",
			},
		})
		metadataBlock := cronjobBody.AppendNewBlock("metadata", nil)
		metadataBody := metadataBlock.Body()
		flattenMetadata(d.Metadata, metadataBody)
		specBlock := cronjobBody.AppendNewBlock("spec", nil)
		specBody := specBlock.Body()
		flattenJobV1Spec(d.Spec, specBody)

		cronjobBody.SetAttributeValue("is_any_host_allowed", cty.BoolVal(d.IsAnyHostAllowed))

		_, err = tfFile.Write(hclFile.Bytes())
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		// Import all created resources.
		if config.GenerateTfState {
			importConfigs = append(importConfigs, common.ImportConfig{
				ResourceAddress: "duplocloud_k8_secret." + resourceName,
				ResourceId:      "v3/subscriptions/" + config.TenantId + "/k8s/cronjob/" + resourceName,
				WorkingDir:      workingDir,
			})

			tfContext.ImportConfigs = importConfigs
		}
		// initialize the body of the new file object
	}
	return &tfContext, nil

}
