package main

import (
	"context"
	"log"
	"net/http"
	"time"
	"userbalance/internal/config"
)

type Server struct {
	httpServer *http.Server
	conf       *config.Config
}

func (s *Server) Run(host string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:         host,
		Handler:      handler,
		ReadTimeout:  time.Duration(s.conf.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.conf.WriteTimeout) * time.Second,
	}

	log.Println("сервер запущен")

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
