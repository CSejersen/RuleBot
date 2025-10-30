package api

import (
	"encoding/json"
	"home_automation_server/types"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WSManager struct {
	clients map[*websocket.Conn]struct{}
	mu      sync.Mutex
}

func NewWSManager() *WSManager {
	return &WSManager{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (wsm *WSManager) AddClient(conn *websocket.Conn) {
	wsm.mu.Lock()
	wsm.clients[conn] = struct{}{}
	wsm.mu.Unlock()
}

func (wsm *WSManager) RemoveClient(conn *websocket.Conn) {
	wsm.mu.Lock()
	delete(wsm.clients, conn)
	wsm.mu.Unlock()
	conn.Close()
}

func (wsm *WSManager) Start(eventCh chan types.Event) {
	go func() {
		for event := range eventCh {
			wsm.mu.Lock()
			for client := range wsm.clients {
				if err := client.WriteJSON(event); err != nil {
					client.Close()
					delete(wsm.clients, client)
				}
			}
			wsm.mu.Unlock()
		}
	}()
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // TODO: fix cors
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Logger.Error("Failed WS upgrade", zap.Error(err))
		return
	}

	s.WSManager.AddClient(conn)
	defer s.WSManager.RemoveClient(conn)

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			s.Logger.Info("WS client disconnected", zap.Error(err))
			break
		}

		var msg struct {
			Type string                 `json:"type"`
			Data map[string]interface{} `json:"data,omitempty"`
		}
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			s.Logger.Warn("Invalid WS message", zap.Error(err))
			continue
		}

		switch msg.Type {
		case "reload_automations":
			if err := s.Engine.LoadAutomations(s.ctx); err != nil {
				s.Logger.Error("Failed to reload rules", zap.Error(err))
			}
		case "load_integration":
			integrationName := msg.Data["integration_name"].(string)
			if err := s.Engine.LoadIntegration(s.ctx, integrationName); err != nil {
				s.Logger.Error("Failed to load integration", zap.Error(err), zap.String("integration_name", integrationName))
			}
		default:
			s.Logger.Warn("Unknown WS command", zap.String("type", msg.Type))
		}
	}
}
