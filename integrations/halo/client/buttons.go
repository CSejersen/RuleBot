package client

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type UpdateButton struct {
	Type    string              `json:"type"`
	Id      string              `json:"id"`
	Value   int                 `json:"value,omitempty"`
	Content UpdateButtonContent `json:"content,omitempty"`
}

type UpdateButtonContent struct {
	Text string `json:"text,omitempty"`
	Icon string `json:"icon,omitempty"`
}

func (c *Client) UpdateButtonValue(name string, val int) error {
	btnID, err := c.ResolveBtnId(name)
	if err != nil {
		c.Logger.Error("Failed to resolve button id", zap.String("name", name), zap.Error(err))
		return err
	}

	update := UpdateCommand[UpdateButton]{
		Update: UpdateButton{
			Type:  "button",
			Id:    btnID,
			Value: val,
			Content: UpdateButtonContent{
				Icon: "lights",
			},
		},
	}

	c.Logger.Info("Updating button", zap.String("name", name), zap.Int("value", val), zap.Any("update", update))

	bytes, err := json.Marshal(update)
	if err != nil {
		c.Logger.Error("failed to marshal update request", zap.Error(err))
		return err
	}

	err = c.Conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		c.Logger.Error("failed to send update request", zap.Error(err))
	}

	return nil
}

func (c *Client) Buttons() []*Button {
	buttons := []*Button{}
	for _, page := range c.Config.Pages {
		for _, button := range page.Buttons {
			buttons = append(buttons, button)
		}
	}
	return buttons
}
