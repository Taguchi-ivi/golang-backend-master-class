package main

import (
	"context"
	"input-backend-master-class/api"
	db "input-backend-master-class/db/generated"
	"input-backend-master-class/mail"
	"input-backend-master-class/util"
	"input-backend-master-class/worker"
	"log"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	// conn, err := sql.Open(config.DBDriver, config.DBSource)
	conn, err := pgx.Connect(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	go runTaskProcessor(config, redisOpt, store)

	server, err := api.NewServer(config, &store, taskDistributor)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}

func runTaskProcessor(config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal("cannot start task processor:", err)
	}
}
