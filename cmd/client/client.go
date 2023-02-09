package main

import (
	"log"
	c "userbalance/internal/config"
	"userbalance/pkg/api"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client api.UserBalanceClient
	conn   *grpc.ClientConn
}

func (c *Client) Run(conf *c.Config) error {
	conn, err := grpc.Dial(conf.ServerHost+conf.ServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	c.conn = conn
	if err != nil {
		return err
	}
	c.client = api.NewUserBalanceClient(conn)
	log.Println("подключение установлено")

	return nil
}

func (c *Client) Shutdown() {
	c.conn.Close()
}
