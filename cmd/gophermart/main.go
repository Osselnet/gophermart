package main

import (
	"github.com/Osselnet/gophermart.git/internal/db"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"github.com/Osselnet/gophermart.git/internal/server"
	"github.com/Osselnet/gophermart.git/internal/server/config"
	"github.com/Osselnet/gophermart.git/internal/server/handlers"
	"log"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		panic(err)
	}
	log.Printf("[DEBUG] Receive config: %#v\n", cfg)

	st, err := db.New(cfg.DatabaseURI)
	if err != nil {
		log.Fatalln("[FATAL] Postgres initialization failed - ", err)
	}

	gm := gophermart.New(st)

	h := handlers.New(gm)

	s := server.New(h.GetRouter(), cfg.Addr)
	go s.Serve()

	queue := gophermart.NewQueue(st, cfg.AccrualSystemAddress)
	queue.Start()
}
