package tfgenerator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type Importer struct {
}

type ImportConfig struct {
	resourceAddress string
	resourceId      string
	workingDir      string
}

func (i *Importer) Import(config *Config, importConfig *ImportConfig) {
	log.Printf("[TRACE] Importing terraform resource  : (%s, %s).", importConfig.resourceAddress, importConfig.resourceId)
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("0.14.11")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
	tf, err := tfexec.NewTerraform(importConfig.workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}
	err = tf.Import(context.Background(), importConfig.resourceAddress, importConfig.resourceId)
	if err != nil {
		log.Fatalf("error running Import: %s", err)
	}
	state, err := tf.Show(context.Background())
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	stateJson, err := json.Marshal(state.Values)
	fmt.Println(string(stateJson))

	log.Printf("[TRACE] Terraform resource (%s, %s) is imported.", importConfig.resourceAddress, importConfig.resourceId)
}
