package config

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	jsonFolder := ""
	jsonFile := "config.json"
	dataJsonFile := "dataconfig.json"
	cfg, dcfg, err := LoadConfig(jsonFolder, jsonFile, dataJsonFile)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(cfg, dcfg)

}
