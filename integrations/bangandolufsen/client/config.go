package client

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Devices map[string]Device `yaml:"devices"` // key: friendly_name (name displayed by the bang and olufsen app)
}

type Device struct {
	IP       string `yaml:"ip"`
	JID      string `yaml:"jid"`
	IsMozart bool   `yaml:"is_mozart"`
}

// TODO: implement a fs watcher for the rules file to update the config on changes.
func (c *Client) loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file")
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	c.Config = cfg
	return nil
}
