package api

import (
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"home_automation_server/engine/types"
	"net/http"
	"sync"
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

func (wsm *WSManager) Start(eventCh chan types.ProcessedEvent) {
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
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
