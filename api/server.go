package api

import (
	"context"
	"go.uber.org/zap"
	"home_automation_server/engine"
	"home_automation_server/engine/types"
	"net/http"
)

type Server struct {
	mux       *http.ServeMux
	httpSrv   *http.Server
	Engine    *engine.Engine
	WSManager *WSManager
	Logger    *zap.Logger
}

func NewServer(e *engine.Engine, logger *zap.Logger, eventCh chan types.ProcessedEvent) *Server {
	s := &Server{
		Engine:    e,
		WSManager: NewWSManager(),
		mux:       http.NewServeMux(),
		Logger:    logger,
	}

	s.WSManager.Start(eventCh)
	s.routes()
	return s
}

func (s *Server) Start(addr string) {
	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: s,
	}

	go func() {
		s.Logger.Info("Starting API + WS server", zap.String("addr", addr))
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.Fatal("Server failed", zap.Error(err))
		}
	}()
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/integrations", s.handleIntegrations)
	s.mux.HandleFunc("/api/services", s.handleServices)
	s.mux.HandleFunc("/api/integrations/", s.handleIntegrationSubresources)
	s.mux.HandleFunc("/ws", s.handleWS)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Close all WS clients
	s.WSManager.mu.Lock()
	for client := range s.WSManager.clients {
		client.Close()
		delete(s.WSManager.clients, client)
	}
	s.WSManager.mu.Unlock()

	// Shutdown HTTP server
	if s.httpSrv != nil {
		return s.httpSrv.Shutdown(ctx)
	}
	return nil
}
