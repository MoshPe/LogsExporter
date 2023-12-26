package utils

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

func FormatOutput(in interface{}) (string, error) {
	//Convert the data to YAML
	yamlBytes, err := yaml.Marshal(in)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	yamlString := string(yamlBytes)
	return yamlString, nil
}
