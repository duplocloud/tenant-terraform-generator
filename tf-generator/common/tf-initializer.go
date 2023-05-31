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

type TfInitializer struct {
	WorkingDir string
	Config     *Config
}

func (tfi *TfInitializer) InitWithWorkspace() *tfexec.Terraform {
	log.Println("[TRACE] <================================== TF init in progress. ==================================>")
	tfVersion := GetEnv("tf_version", TF_DEFAULT_VERSION)
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(tfVersion)),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
	tf, err := tfexec.NewTerraform(tfi.WorkingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}
	//backend := "-backend-config=bucket=duplo-tfstate-" + config.AccountID + " -backend-config=dynamodb_table=duplo-tfstate-" + config.AccountID + "-lock"
	//err = tf.Init(context.Background(), tfexec.Upgrade(true), tfexec.BackendConfig("bucket=duplo-tfstate-"+config.AccountID), tfexec.BackendConfig("dynamodb_table=duplo-tfstate-"+config.AccountID+"-lock"))
	if tfi.Config.S3Backend {
		err = tf.Init(context.Background(), tfexec.Upgrade(true), tfexec.BackendConfig("bucket=duplo-tfstate-"+tfi.Config.AccountID))
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

	if duplosdk.Contains(workspaceList, tfi.Config.TenantName) {
		err = tf.WorkspaceSelect(context.Background(), tfi.Config.TenantName)
		if err != nil {
			log.Fatalf("error running tf workspace select: %s", err)
		}
		log.Printf("[TRACE] (%s) workspace is selected.", tfi.Config.TenantName)
	} else {
		err := tf.WorkspaceNew(context.Background(), tfi.Config.TenantName)
		if err != nil {
			log.Fatalf("error running tf workspace new: %s", err)
		}
		log.Printf("[TRACE] (%s) workspace is created.", tfi.Config.TenantName)
	}
	log.Printf("[TRACE] Terraform initialized with new workspace - %s", tfi.Config.TenantName)
	log.Println("[TRACE] <====================================================================>")
	return tf
}

func (tfi *TfInitializer) Init(config *Config, workingDir string) *tfexec.Terraform {
	tfVersion := GetEnv("tf_version", TF_DEFAULT_VERSION)
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(tfVersion)),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}
	if config.S3Backend {
		err = tf.Init(context.Background(), tfexec.Upgrade(true), tfexec.BackendConfig("bucket=duplo-tfstate-"+config.AccountID), tfexec.BackendConfig("dynamodb_table=duplo-tfstate-"+config.AccountID+"-lock"))
	} else {
		err = tf.Init(context.Background(), tfexec.Upgrade(true))
	}

	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}
	return tf
}

func (tfi *TfInitializer) NewWorkspace(config *Config, tf *tfexec.Terraform) {
	workspaceList, activeWorkspace, err := tf.WorkspaceList(context.Background())
	if err != nil {
		log.Fatalf("error running tf workspace list: %s", err)
	}
	if len(workspaceList) > 0 {
		log.Printf("[TRACE] Workspace List (%s).", workspaceList)
		log.Printf("[TRACE] Active Workspace (%s).", activeWorkspace)
	}
	if !duplosdk.Contains(workspaceList, tfi.Config.TenantName) {
		err := tf.WorkspaceNew(context.Background(), config.TenantName)
		if err != nil {
			log.Fatalf("error running tf workspace new: %s", err)
		}
		log.Printf("[TRACE] (%s) workspace is created.", config.TenantName)
	}

}

func (tfi *TfInitializer) DeleteWorkspace(config *Config, tf *tfexec.Terraform) {
	workspaceList, activeWorkspace, err := tf.WorkspaceList(context.Background())
	if err != nil {
		log.Fatalf("error running tf workspace list: %s", err)
	}
	if len(workspaceList) > 0 {
		log.Printf("[TRACE] Workspace List (%s).", workspaceList)
		log.Printf("[TRACE] Active Workspace (%s).", activeWorkspace)
	}
	if duplosdk.Contains(workspaceList, tfi.Config.TenantName) {
		err := tf.WorkspaceSelect(context.Background(), "default")
		if err != nil {
			log.Fatalf("error running tf workspace select(default): %s", err)
		}
		err = tf.WorkspaceDelete(context.Background(), config.TenantName)
		if err != nil {
			log.Fatalf("error running tf workspace delete: %s", err)
		}
		log.Printf("[TRACE] Workspace deleted (%s).", config.TenantName)
	}

}
