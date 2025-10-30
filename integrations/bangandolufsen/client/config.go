package client

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Devices map[string]Device `yaml:"devices"` // key: JID
}

type Device struct {
	IP       string `yaml:"ip"`
	JID      string `yaml:"jid"`
	IsMozart bool   `yaml:"is_mozart"`
}

func (c *Client) IpForDevice(jid string) (string, bool) {
	device, ok := c.Config.Devices[jid]
	return device.IP, ok
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

func (c *Client) WatchConfig(ctx context.Context, path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		c.Logger.Error("failed to create fs watcher", zap.Error(err))
		return
	}
	defer watcher.Close()

	if err := watcher.Add(path); err != nil {
		c.Logger.Error("failed to add rules file to watcher", zap.String("path", path), zap.Error(err))
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-watcher.Events:
			if evt.Op&(fsnotify.Write|fsnotify.Create) > 0 {
				c.Logger.Info("config file changed, reloading", zap.String("file", evt.Name))
				if err := c.loadConfig(path); err != nil {
					c.Logger.Error("failed to reload config", zap.Error(err))
				}
			}
		case err := <-watcher.Errors:
			c.Logger.Warn("config watcher error", zap.Error(err))
		}
	}
}
