package client

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

type ConfigWrapper struct {
	Configuration Config `json:"configuration"`
}

type Config struct {
	Version string  `yaml:"version" json:"version"`
	ID      string  `yaml:"id" json:"id"`
	Pages   []*Page `yaml:"pages" json:"pages"`
}

type Registry struct {
	Buttons map[string]*Button
}

// implement a fs watcher for the rules file to update the config on changes.
func (c *Client) loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file")
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	c.Config = &cfg

	return nil
}

func (c *Client) deployConfig(cfg *Config) error {
	wrapped := ConfigWrapper{
		Configuration: *cfg,
	}
	c.Logger.Debug("Deploying config", zap.Any("config", wrapped))
	bytes, err := json.Marshal(wrapped)
	if err != nil {
		return fmt.Errorf("failed to marshal beoremote halo config: %w", err)
	}
	if err := c.Conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
		return fmt.Errorf("failed to send beoremote halo config: %w", err)
	}
	return nil
}
