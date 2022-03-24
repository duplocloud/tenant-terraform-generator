package common

import (
	"context"
	"log"
	"tenant-terraform-generator/duplosdk"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type Importer struct {
}

type ImportConfig struct {
	ResourceAddress string
	ResourceId      string
	WorkingDir      string
}

func (i *Importer) Import(config *Config, importConfig *ImportConfig) {
	log.Println("[TRACE] <================================== TF Import in progress. ==================================>")
	log.Printf("[TRACE] Importing terraform resource  : (%s, %s).", importConfig.ResourceAddress, importConfig.ResourceId)
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("0.14.11")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
	tf, err := tfexec.NewTerraform(importConfig.WorkingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}
	//backend := "-backend-config=bucket=duplo-tfstate-" + config.AccountID + " -backend-config=dynamodb_table=duplo-tfstate-" + config.AccountID + "-lock"
	//err = tf.Init(context.Background(), tfexec.Upgrade(true), tfexec.BackendConfig("bucket=duplo-tfstate-"+config.AccountID), tfexec.BackendConfig("dynamodb_table=duplo-tfstate-"+config.AccountID+"-lock"))
	if config.S3Backend {
		err = tf.Init(context.Background(), tfexec.Upgrade(true), tfexec.BackendConfig("bucket=duplo-tfstate-"+config.AccountID))
	} else {
		err = tf.Init(context.Background(), tfexec.Upgrade(true))
	}

	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	workspaceList, activeWorkspace, err := tf.WorkspaceList(context.Background())
	if err != nil {
		log.Fatalf("error running tf workspace list: %s", err)
	}
	if len(workspaceList) > 0 {
		log.Printf("[TRACE] Workspace List (%s).", workspaceList)
		log.Printf("[TRACE] Active Workspace (%s).", activeWorkspace)
	}

	if duplosdk.Contains(workspaceList, config.TenantName) {
		err = tf.WorkspaceSelect(context.Background(), config.TenantName)
		if err != nil {
			log.Fatalf("error running tf workspace select: %s", err)
		}
		log.Printf("[TRACE] (%s) workspace is selected.", config.TenantName)
	} else {
		err := tf.WorkspaceNew(context.Background(), config.TenantName)
		if err != nil {
			log.Fatalf("error running tf workspace new: %s", err)
		}
		log.Printf("[TRACE] (%s) workspace is created.", config.TenantName)
	}

	err = tf.Import(context.Background(), importConfig.ResourceAddress, importConfig.ResourceId)
	if err != nil {
		log.Fatalf("error running Import: %s", err)
	}
	_, err = tf.Show(context.Background())
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	//_, err = json.Marshal(state.Values)
	//fmt.Println(string(stateJson))

	log.Printf("[TRACE] Terraform resource (%s, %s) is imported.", importConfig.ResourceAddress, importConfig.ResourceId)
	log.Println("[TRACE] <====================================================================>")
}

func (i *Importer) ImportWithoutInit(config *Config, importConfig *ImportConfig, tf *tfexec.Terraform) {
	log.Println("[TRACE] <================================== TF Import in progress. ==================================>")
	log.Printf("[TRACE] Importing terraform resource  : (%s, %s).", importConfig.ResourceAddress, importConfig.ResourceId)

	err := tf.Import(context.Background(), importConfig.ResourceAddress, importConfig.ResourceId)
	if err != nil {
		log.Fatalf("error running Import: %s", err)
	}
	_, err = tf.Show(context.Background())
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	//_, err = json.Marshal(state.Values)
	//fmt.Println(string(stateJson))

	log.Printf("[TRACE] Terraform resource (%s, %s) is imported.", importConfig.ResourceAddress, importConfig.ResourceId)
	log.Println("[TRACE] <====================================================================>")
}
