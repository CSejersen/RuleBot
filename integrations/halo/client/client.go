package client

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	Config *Config
	Conn   *websocket.Conn
	SendCh <-chan UpdateCommand[any]
	Logger *zap.Logger
}

func (c *Client) runSender() {
	for msg := range c.SendCh {
		if err := c.Conn.WriteJSON(msg); err != nil {
			c.Logger.Error("failed to send message", zap.Error(err))
		}
	}
}

func New(addr, configPath string, logger *zap.Logger) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to beoremote halo websocket: %w", err)
	}

	logger.Debug("established connection to halo ws")

	h := &Client{
		Conn:   conn,
		Logger: logger,
	}

	if err := h.loadConfig(configPath); err != nil {
		return nil, err
	}
	if err := h.deployConfig(h.Config); err != nil {
		return nil, err
	}
	logger.Debug("config successfully deployed")

	return h, nil
}

func (c *Client) ResolveBtnId(name string) (string, error) {
	for _, page := range c.Config.Pages {
		for _, button := range page.Buttons {
			if button.Title == name {
				return button.ID, nil
			}
		}
	}
	return "", fmt.Errorf("could not find button: %s", name)
}
