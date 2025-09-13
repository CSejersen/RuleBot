package client

import (
	"encoding/json"
	"fmt"
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

func (c *Client) UpdateButtonValue(id string, val int) error {

	update := UpdateCommand[UpdateButton]{
		Update: UpdateButton{
			Type:  "button",
			Id:    id,
			Value: val,
			Content: UpdateButtonContent{
				Icon: "lights",
			},
		},
	}

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

func (c *Client) Buttons() []*Button {
	buttons := []*Button{}
	for _, page := range c.Config.Pages {
		for _, button := range page.Buttons {
			buttons = append(buttons, button)
		}
	}
	return buttons
}
