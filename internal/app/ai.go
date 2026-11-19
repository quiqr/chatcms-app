package app

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/version"
	"github.com/quiqr/chatcms-app/internal/workspace"
)

func (s *Server) startAI() error {
	books := aigateway.NewPostgres(s.db.Queries)
	if !s.cfg.AIConfigured() {
		s.logger.Info("the AI gateway is off: no provider is configured")
		return nil
	}

	gateway, err := aigateway.New(books, aigateway.Settings{
		ProviderURL:   s.cfg.AIProviderURL,
		Key:           s.cfg.AIKey,
		Model:         s.cfg.AIModel,
		SessionHeader: s.cfg.AISessionHeader,
		UserAgent:     "quiqr-chatcms/" + version.String(),
		Logger:        s.logger,
	})
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", s.cfg.AIListen)
	if err != nil {
		return fmt.Errorf("the AI gateway could not listen on %s: %w", s.cfg.AIListen, err)
	}
	s.aiListener = listener
	s.aiGateway = gateway
	s.aiMux = http.NewServeMux()
	s.aiMux.Handle("/v1/", gateway)
	s.aiMux.Handle("/", gateway)
	s.aiHTTP = &http.Server{Handler: s.aiMux, ReadHeaderTimeout: 10 * time.Second}
	s.aiTokens = aigateway.NewTokens(books)
	return nil
}

func (s *Server) workspaceTokens() workspace.Tokens {
	if s.aiTokens == nil {
		return nil
	}
	return s.aiTokens
}

func (s *Server) gatewayPort() int {
	if s.aiListener == nil {
		return 0
	}
	_, port, err := net.SplitHostPort(s.aiListener.Addr().String())
	if err != nil {
		return 0
	}
	number, _ := strconv.Atoi(port)
	return number
}

func (s *Server) AIAddr() string {
	if s.aiListener == nil {
		return ""
	}
	return s.aiListener.Addr().String()
}
