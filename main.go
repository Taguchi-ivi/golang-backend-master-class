package main

import (
	"context"
	"input-backend-master-class/api"
	db "input-backend-master-class/db/generated"
	"input-backend-master-class/util"
	"log"

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
	server, err := api.NewServer(config, &store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}
