package main

import (
	"context"
	"github.com/Osselnet/gophermart.git/internal/client"
	"github.com/Osselnet/gophermart.git/internal/db"
	"github.com/Osselnet/gophermart.git/internal/gophermart"
	"github.com/Osselnet/gophermart.git/internal/server"
	"github.com/Osselnet/gophermart.git/internal/server/config"
	"github.com/Osselnet/gophermart.git/internal/server/handlers"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultGraceTimeout = 20 * time.Second
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

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sig

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), defaultGraceTimeout)
		defer shutdownCtxCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("[ERROR] Graceful shutdown timed out! Forcing exit.")
			}
		}()

		err := s.Server.Shutdown(context.Background())
		if err != nil {
			log.Fatal("[ERROR] Server shutdown error - ", err)
		}
	}()

	queue := client.NewQueue(st, cfg.AccrualSystemAddress)
	queue.Start()
}
