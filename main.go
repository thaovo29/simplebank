package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/thaovo29/simplebank/api"
	db "github.com/thaovo29/simplebank/db/sqlc"
	"github.com/thaovo29/simplebank/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot read config: ", err)
	}
	
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)
	err = server.Start(config.ServerAddress)
	if err != nil { 
		log.Fatal("cannot start server: ", err)
	}
}