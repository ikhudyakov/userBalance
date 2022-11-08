package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	var path string
	var defaultPath string = "./configs/config.yaml"

	path = defaultPath

	flag.StringVar(&path, "config", "./configs/config.yaml", "example -config ./configs/config.yaml")
	migrationup := flag.Bool("migrationup", false, "use migrationup to perform migrationup")
	migrationdown := flag.Bool("migrationdown", false, "use migrationdown to perform migrationdown")

	flag.Parse()

	conf, err = c.GetConfig(path)
	if err != nil {
		log.Printf("%s, use default config '%s'", err, defaultPath)
		conf, err = c.GetConfig(defaultPath)
		if err != nil {
			log.Println(err)
			return
		}
	}

	if err = repository.Migration(conf, *migrationup, *migrationdown); err != nil {
		log.Println(err)
	}

	if db, err = repository.Connect(conf); err != nil {
		log.Println(err)
		return
	}

	repos := repository.NewRepository(db)
	services = service.NewService(repos, conf, db)
	handlers := handler.NewHandler(services)

	server := new(Server)
	server.conf = conf

	go func() {
		if err := server.Run(conf.Host, handlers.Init()); err != nil {
			log.Fatalf("ошибка при запуске http сервера: %s", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("сервер останавливается")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContexTimeout)*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("произошла ошибка при выключении сервера: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		log.Fatalf("произошла ошибка при закрытии соединения с БД: %s", err.Error())
	}

}
