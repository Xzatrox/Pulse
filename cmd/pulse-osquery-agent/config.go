package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type AgentConfig struct {
	Server struct {
		URL      string `yaml:"url"`
		APIToken string `yaml:"api_token"`
	} `yaml:"server"`
	
	Agent struct {
		ID       string        `yaml:"id"`
		Interval time.Duration `yaml:"interval"`
	} `yaml:"agent"`
	
	Filter struct {
		Mode            string   `yaml:"mode"`
		ExcludePatterns []string `yaml:"exclude_patterns"`
	} `yaml:"filter"`
	
	Logging struct {
		Debug bool `yaml:"debug"`
	} `yaml:"logging"`
}

func loadConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var cfg AgentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	
	return &cfg, nil
}
