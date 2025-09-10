package client

import (
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"home_automation_server/integrations/halo/client/handlers"
)

type Client struct {
	Conn   *websocket.Conn
	SendCh <-chan handlers.UpdateCommand[any]
	Logger *zap.Logger
}

type NewParams struct {
	Addr   string
	Logger *zap.Logger
}

func (c *Client) runSender() {
	for msg := range c.SendCh {
		if err := c.Conn.WriteJSON(msg); err != nil {
			c.Logger.Error("failed to send message", zap.Error(err))
		}
	}
}

func New(p NewParams) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(p.Addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to beoremote halov1 websocket: %w", err)
	}

	fmt.Println("Connected to beoremote halov1 WebSocket:", p.Addr)
	fmt.Println("Sending initial config")

	h := &Client{
		Conn:   conn,
		Logger: p.Logger,
	}

	if err := h.loadConfig("./config.yaml"); err != nil {
		return nil, err
	}

	return h, nil
}
