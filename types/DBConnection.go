package types

import (
	"os"

	"gopkg.in/yaml.v2"
)

type DBConnection struct {
	Userid   string `yaml:"userid"`
	Password string `yaml:"password"`
	Server   string `yaml:"server"`
	Database string `yaml:"database"`
}

func NewDBConnection() (*DBConnection, error) {

	configPath := "./configs/mssql_cfg.yaml"

	config := &DBConnection{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
