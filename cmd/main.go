package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"userbalance"
	c "userbalance/internal/config"
	"userbalance/internal/handler"
	"userbalance/internal/repository"
	"userbalance/internal/service"
)

// @title UserBalance API
// @version 1.0
// @description Microservice for working with user balance

// @host localhost:8081
// @BasePath /

func main() {
	var services *service.Service
	var db *sql.DB
	var err error
	var conf *c.Config
	var path string = "./configs/config.toml"

	conf, err = c.GetConfig(path)
	if err != nil {
		log.Println(err)
		return
	}

	if err = migration(conf); err != nil {
		log.Println(err)
	}

	if db, err = repository.Connect(conf); err != nil {
		log.Println(err)
		return
	}

	repos := repository.NewRepository(db)
	services = service.NewService(repos)
	handlers := handler.NewHandler(services)

	server := new(userbalance.Server)

	go func() {
		if err := server.Run(conf.Port, handlers.Init()); err != nil {
			log.Fatalf("ошибка при запуске http сервера: %s", err.Error())
		}
	}()

	log.Println("сервер запущен")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("сервер останавливается")

	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("произошла ошибка при выключении сервера: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		log.Fatalf("произошла ошибка при закрытии соединения с БД: %s", err.Error())
	}

}
