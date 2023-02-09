package main

import (
	"log"
	"net"

	"userbalance/internal/config"
	"userbalance/internal/handler"
	"userbalance/pkg/api"

	"google.golang.org/grpc"
)

type Server struct {
	server *grpc.Server
	conf   *config.Config
}

func (s *Server) Run(port string, handler *handler.Handler) error {
	s.server = grpc.NewServer()
	api.RegisterUserBalanceServer(s.server, handler)

	l, err := net.Listen("tcp", s.conf.Port)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("сервер запущен")

	if err := s.server.Serve(l); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *Server) Shutdown() {
	s.server.GracefulStop()
}
