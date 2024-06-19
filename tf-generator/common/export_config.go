package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

type ConfigVars struct {
	TargetLocation string
	Config         map[string]interface{}
	Project        string
}

/*
func (ec *ConfigVars) Generate() {

		//if len(ec.Config) > 0 {
		log.Println("[TRACE] <====== Genertating config values  started. =====>")

		// create new empty hcl file object
		hclFile := hclwrite.NewEmptyFile()
		// create new file on system
		path := filepath.Join(ec.TargetLocation, "config.vars.tf")
		tfFile, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
			return
		}

		// initialize the body of the new file object
		rootBody := hclFile.Body()
		rootBody.
		fmt.Printf("%s", hclFile.Bytes())
		_, err = tfFile.Write(hclFile.Bytes())
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Println("[TRACE] <====== Variables TF generation done. =====>")
		// }
	}
*/

func (ec *ConfigVars) Generate() error {

	if len(ec.Config) == 0 {
		log.Println("[TRACE] Config is empty, skipping generation.")
		return nil
	}

	log.Println("[TRACE] <====== Generating JSON config values started. =====>")

	// Marshal Config map to JSON
	jsonData, err := json.MarshalIndent(ec.Config, "", "  ") // Use MarshalIndent for indentation
	if err != nil {
		fmt.Println("Error marshalling config to JSON:", err)
		return err
	}

	// Create the file path

	path := filepath.Join(ec.TargetLocation, ec.Project+".tfvars.json")
	fmt.Println("\n*********** \n Path to write config file ", path)
	// Write JSON data to file
	err = ioutil.WriteFile(path, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON to file:", err)
		return err
	}

	log.Println("[TRACE] <====== JSON config generation done. =====>")
	return nil
}
