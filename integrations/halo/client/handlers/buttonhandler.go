package handlers

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"home_automation_server/integrations/halo/client"
)

type UpdateButton struct {
	Content UpdateButtonContent `json:"content,omitempty"`
	Type    string              `json:"type"`
	Id      string              `json:"id"`
	Value   int                 `json:"value,omitempty"`
}

type UpdateButtonContent struct {
	Text string `json:"text,omitempty"`
	Icon string `json:"icon,omitempty"`
}

func UpdateButtonValue(c *client.Client, id string, val int) error {
	// TODO: clamp value between 0..100
	c.Logger.Info("Updating button value", zap.String("id", id), zap.Int("value", val))

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
