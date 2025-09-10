package client

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"os"
)

type ConfigWrapper struct {
	Configuration Config `json:"configuration"`
}

type Config struct {
	Version string  `json:"version"`
	ID      string  `json:"id"`
	Pages   []*Page `json:"pages"`
}

type Page struct {
	Title   string    `json:"title"`
	ID      string    `json:"id"`
	Buttons []*Button `json:"buttons"`
}

type Button struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Subtitle string  `json:"subtitle"`
	Value    int     `json:"value"`
	State    string  `json:"state"`
	Content  Content `json:"content"`
	Default  bool    `json:"default"`
}

type Content struct {
	Text string `json:"text,omitempty"`
	Icon string `json:"icon,omitempty"`
}

type Registry struct {
	Buttons map[string]*Button
}

// TODO: Maybe move responsibility to the engine
func (c *Client) loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file")
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	// TODO: build registry

	if err := c.deployConfig(&cfg); err != nil {
		return err
	}

	return nil
}

func (c *Client) deployConfig(cfg *Config) error {
	wrapped := ConfigWrapper{
		Configuration: *cfg,
	}
	bytes, err := json.Marshal(wrapped)
	if err != nil {
		return fmt.Errorf("failed to marshal beoremote halo config: %w", err)
	}
	if err := c.Conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
		return fmt.Errorf("failed to send beoremote halo config: %w", err)
	}
	return nil
}
