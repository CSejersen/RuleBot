package client

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"time"
)

type Client struct {
	Config    *Config
	Conn      *websocket.Conn
	SendCh    chan UpdateCommand[any]
	ReceiveCh chan []byte
	Logger    *zap.Logger
}

func New(configPath string, logger *zap.Logger) (*Client, error) {
	c := &Client{
		Logger:    logger,
		ReceiveCh: make(chan []byte, 100),
		SendCh:    make(chan UpdateCommand[any], 100),
	}

	err := c.loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	logger.Debug("config successfully deployed")
	return c, nil
}

func (c *Client) Run(ctx context.Context, addr string) {
	for {
		conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
		if err != nil {
			c.Logger.Warn("failed to connect to halo ws", zap.Error(err))
			select {
			case <-time.After(5 * time.Second):
				continue
			case <-ctx.Done():
				return
			}
		}

		c.Conn = conn
		c.Logger.Info("connected to halo websocket")

		if err := c.deployConfig(c.Config); err != nil {
			c.Logger.Warn("failed to deploy config", zap.Error(err))
		}

		done := make(chan struct{})
		go c.readLoop(done)
		go c.writeLoop(done)

		select {
		case <-done:
			_ = conn.Close()
			c.Logger.Warn("connection dropped, retrying...")
			time.Sleep(3 * time.Second)
		case <-ctx.Done():
			_ = conn.Close()
			return
		}
	}
}

func (c *Client) readLoop(done chan<- struct{}) {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			close(done)
			return
		}
		c.ReceiveCh <- msg
	}
}

func (c *Client) writeLoop(done chan<- struct{}) {
	for msg := range c.SendCh {
		if err := c.Conn.WriteJSON(msg); err != nil {
			close(done)
			return
		}
	}
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
