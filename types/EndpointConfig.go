package types

import (
	"os"

	"github.com/gopcua/opcua"
	"gopkg.in/yaml.v2"
)

type Tag struct {
	ID          string `yaml:"id"`
	Enabled     int    `yaml:"enabled"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Address     string `yaml:"address"`
}

type EndpointConfig struct {
	Client   *opcua.Client
	Endpoint string `yaml:"endpoint"`
	Interval int    `yaml:"interval"`
	Tags     []Tag  `yaml:"tags"`
}

func NewEndpointConfig() *[]EndpointConfig {
	config := EndpointConfig{}
	configs := []EndpointConfig{config}
	return &configs
}

func GenerateYaml(data *[]EndpointConfig) error {

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	// Write the YAML data to a file
	err = os.WriteFile("./configs/service_cfg.yaml", yamlData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func GetEndpoints() (*[]EndpointConfig, error) {
	configPath := "./configs/service_cfg.yaml"

	config := []EndpointConfig{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
