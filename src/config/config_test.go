package config

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	jsonFolder := ""
	jsonFile := "config.json"
	cfg := LoadConfig(jsonFolder, jsonFile)
	fmt.Println(cfg)

}
