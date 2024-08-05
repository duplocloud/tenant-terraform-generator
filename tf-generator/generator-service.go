package tfgenerator

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tenant-terraform-generator/duplosdk"
	adminInfra "tenant-terraform-generator/tf-generator/admin-infra"
	awsservices "tenant-terraform-generator/tf-generator/aws-services"
	"tenant-terraform-generator/tf-generator/common"
	"tenant-terraform-generator/tf-generator/tenant"
)

type IGeneratorService interface {
	PreProcess(config *common.Config, client *duplosdk.Client) error
	StartTFGeneration(config *common.Config, client *duplosdk.Client) error
	PostProcess(config *common.Config, client *duplosdk.Client) error
}

type TfGeneratorService struct {
}

func (tfg *TfGeneratorService) PreProcess(config *common.Config, client *duplosdk.Client) error {
	log.Println("[TRACE] <====== Initialize target directory with customer name and tenant id. =====>")
	config.TFCodePath = filepath.Join("target", config.CustomerName, "terraform")
	config.ConfigVars = filepath.Join("target", config.CustomerName, "config")
	tenantProject := filepath.Join(config.TFCodePath, config.TenantProject)
	err := os.RemoveAll(filepath.Join("target", config.CustomerName, config.TenantName))
	if err != nil {
		log.Fatal(err)
	}

	err = os.RemoveAll(tenantProject)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(tenantProject, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	config.AdminTenantDir = tenantProject

	err = os.RemoveAll(config.ConfigVars)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Remove Completed \n Creating ", config.ConfigVars)

	err = os.MkdirAll(config.ConfigVars, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created")
	fmt.Println("Creating env folder under config")

	awsServicesProject := filepath.Join(config.TFCodePath, config.AwsServicesProject)
	err = os.RemoveAll(awsServicesProject)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(awsServicesProject, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	config.AwsServicesDir = awsServicesProject

	appProject := filepath.Join(config.TFCodePath, config.AppProject)
	err = os.RemoveAll(appProject)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(appProject, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	config.AppDir = appProject

	infraProject := filepath.Join(config.TFCodePath, config.InfraProject)
	err = os.RemoveAll(infraProject)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(infraProject, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	config.InfraDir = infraProject

	scriptsPath := filepath.Join("target", config.CustomerName, "scripts")
	err = os.RemoveAll(scriptsPath)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(scriptsPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = duplosdk.CopyDirectory("./scripts", scriptsPath)
	if err != nil {
		log.Fatal(err)
	}
	var mapToRepalce = map[string]string{
		"<--admin-tenant-->": config.TenantProject,
		"<--aws-services-->": config.AwsServicesProject,
		"<--app-->":          config.AppProject,
	}
	common.RepalceStringInFile(filepath.Join(scriptsPath, "plan.sh"), mapToRepalce)
	common.RepalceStringInFile(filepath.Join(scriptsPath, "apply.sh"), mapToRepalce)
	common.RepalceStringInFile(filepath.Join(scriptsPath, "destroy.sh"), mapToRepalce)

	err = duplosdk.Copy(".gitignore", filepath.Join("target", config.CustomerName, ".gitignore"))
	if err != nil {
		log.Fatal(err)
	}
	err = duplosdk.Copy(".envrc", filepath.Join("target", config.CustomerName, ".envrc"))
	if err != nil {
		log.Fatal(err)
	}
	envFile, err := os.OpenFile(filepath.Join("target", config.CustomerName, ".envrc"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer envFile.Close()
	if _, err := envFile.WriteString("\nexport tenant_id=\"" + config.TenantId + "\""); err != nil {
		log.Fatal(err)
	}
	//========

	log.Println("[TRACE] <====== Initialized target directory with customer name and tenant id. =====>")
	return nil
}

func (tfg *TfGeneratorService) StartTFGeneration(config *common.Config, client *duplosdk.Client) error {
	// var tf *tfexec.Terraform
	providerGen := &common.Provider{}
	providerGen.Generate(config, client)

	// if config.GenerateTfState {
	// 	tf := tfInit(config, config.AdminTenantDir)
	// 	tfNewWorkspace(config, tf)
	// }
	configVarsBase := config.ConfigVars
	if !config.SkipAdminTenant {
		log.Println("[TRACE] <====== Start TF generation for tenant project. =====>")
		config.ConfigVars = filepath.Join(config.ConfigVars, config.TenantName)
		err := os.MkdirAll(config.ConfigVars, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		// Register New TF generator for Tenant Project
		tenantGeneratorList := TenantGenerators
		if config.S3Backend {
			tenantGeneratorList = append(tenantGeneratorList, &tenant.TenantBackend{})
		}

		starTFGenerationForProject(config, client, tenantGeneratorList, config.AdminTenantDir)
		if config.ValidateTf {
			common.ValidateAndFormatTfCode(config.AdminTenantDir, config.TFVersion)
		}
		config.ConfigVars = configVarsBase
		log.Println("[TRACE] <====== End TF generation for tenant project. =====>")
	}

	if !config.SkipAwsServices {
		log.Println("[TRACE] <====== Start TF generation for aws services project. =====>")
		// Register New TF generator for AWS Services project
		config.ConfigVars = filepath.Join(config.ConfigVars, config.TenantName)
		awsServcesGeneratorList := AWSServicesGenerators
		if config.S3Backend {
			awsServcesGeneratorList = append(awsServcesGeneratorList, &awsservices.AwsServicesBackend{})
		}
		starTFGenerationForProject(config, client, awsServcesGeneratorList, config.AwsServicesDir)
		if config.ValidateTf {
			common.ValidateAndFormatTfCode(config.AwsServicesDir, config.TFVersion)
		}
		config.ConfigVars = configVarsBase
		log.Println("[TRACE] <====== End TF generation for aws services project. =====>")
	}

	if !config.SkipApp {
		log.Println("[TRACE] <====== Start TF generation for app project. =====>")
		// Register New TF generator for App Services project
		config.ConfigVars = filepath.Join(config.ConfigVars, config.TenantName)
		appGeneratorList := AppGenerators
		//if config.S3Backend { //talk with tahir
		//	appGeneratorList = append(appGeneratorList, &app.AppBackend{})
		//}
		starTFGenerationForProject(config, client, appGeneratorList, config.AppDir)
		if config.ValidateTf {
			common.ValidateAndFormatTfCode(config.AppDir, config.TFVersion)
		}
		config.ConfigVars = configVarsBase
		log.Println("[TRACE] <====== End TF generation for app project. =====>")
	}

	if !config.SkipAdminInfra {
		log.Println("[TRACE] <====== Start TF generation for Admin project. =====>")
		config.ConfigVars = filepath.Join(config.ConfigVars, config.DuploPlanId)
		err := os.MkdirAll(config.ConfigVars, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		// Register New TF generator for Admin Services project
		adminInfraGeneratorList := AdminInfraGenerator
		if config.S3Backend {
			adminInfraGeneratorList = append(adminInfraGeneratorList, &adminInfra.InfraBackend{})
		}
		fmt.Println("adminInfraGeneratorList ", adminInfraGeneratorList)
		config.ConfigVars = strings.Split(config.ConfigVars, "/"+config.TenantName)[0]
		starTFGenerationForProject(config, client, adminInfraGeneratorList, config.InfraDir)
		if config.ValidateTf {
			common.ValidateAndFormatTfCode(config.InfraDir, config.TFVersion)
		}
		config.ConfigVars = configVarsBase
		log.Println("[TRACE] <====== End TF generation for Admin project. =====>")
	}
	return nil
}

func starTFGenerationForProject(config *common.Config, client *duplosdk.Client, generatorList []Generator, targetLocation string) {

	tfContext := common.TFContext{
		TargetLocation: targetLocation,
		InputVars:      []common.VarConfig{},
		OutputVars:     []common.OutputVarConfig{},
		ConfgiVars:     common.ConfigVars{},
	}
	fmt.Println("TF context ", tfContext)
	// 1. Generate Duplo TF resources.
	for _, g := range generatorList {
		c, err := g.Generate(config, client)
		if err != nil {
			log.Fatalf("error running admin tenant tf generation: %s", err)
		}
		fmt.Println("Checking tf context ", c)
		if c != nil {
			if len(c.InputVars) > 0 {
				tfContext.InputVars = append(tfContext.InputVars, c.InputVars...)
			}
			if len(c.OutputVars) > 0 {
				tfContext.OutputVars = append(tfContext.OutputVars, c.OutputVars...)
			}
			if len(c.ImportConfigs) > 0 {
				tfContext.ImportConfigs = append(tfContext.ImportConfigs, c.ImportConfigs...)
			}
		}
	}
	fmt.Println("Checking tf context input vars")

	// 2. Generate input vars.
	if len(tfContext.InputVars) > 0 {
		varsGenerator := common.Vars{
			TargetLocation: tfContext.TargetLocation,
			Vars:           tfContext.InputVars,
		}
		fmt.Println("\n*************\nInput Vars Generator ", varsGenerator, "\n************")
		varsGenerator.Generate()
	}
	// 3. Generate output vars.
	if len(tfContext.OutputVars) > 0 {
		outVarsGenerator := common.OutputVars{
			TargetLocation: tfContext.TargetLocation,
			OutputVars:     tfContext.OutputVars,
		}
		outVarsGenerator.Generate()
	}
	// 4. Generate json for config folder
	token := strings.Split(tfContext.TargetLocation, "/")
	projectName := token[len(token)-1]

	configVarsGenerator := common.ConfigVars{
		TargetLocation: config.ConfigVars,
		Config:         common.ConstructConfigVars(tfContext.InputVars),
		Project:        projectName,
	}
	configVarsGenerator.Generate()

	// 5. Import all resources
	if config.GenerateTfState && len(tfContext.ImportConfigs) > 0 {
		tfInitializer := common.TfInitializer{
			WorkingDir: targetLocation,
			Config:     config,
		}
		tf := tfInitializer.InitWithWorkspace()
		importer := &common.Importer{}
		// Get state file if already present.
		state, err := tf.Show(context.Background())
		if err != nil {
			// log.Fatalf("error running Show: %s", err)
			fmt.Println(err)
		}
		importedResourceAddresses := []string{}
		if state != nil && state.Values != nil && state.Values.RootModule != nil && len(state.Values.RootModule.Resources) > 0 {
			for _, r := range state.Values.RootModule.Resources {
				importedResourceAddresses = append(importedResourceAddresses, r.Address)
			}
		}
		for _, ic := range tfContext.ImportConfigs {
			//importer.Import(config, &ic)
			if common.Contains(importedResourceAddresses, ic.ResourceAddress) {
				log.Printf("[TRACE] Resource %s is already imported.", ic.ResourceAddress)
				continue
			}
			importer.ImportWithoutInit(config, &ic, tf)
		}
		//tfInitializer.DeleteWorkspace(config, tf)
	}
}

func (tfg *TfGeneratorService) PostProcess(config *common.Config, client *duplosdk.Client) error {
	return nil
}
