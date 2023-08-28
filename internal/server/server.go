package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	server       *http.Server
	graceTimeout time.Duration
}

func New(h http.Handler, addr string) *Server {
	const (
		defaultAddress      = ":8080"
		defaultGraceTimeout = 20 * time.Second
	)
	s := &Server{
		server: &http.Server{
			Addr:           defaultAddress,
			ReadTimeout:    time.Second * 10,
			WriteTimeout:   time.Second * 10,
			IdleTimeout:    time.Second * 10,
			MaxHeaderBytes: 1 << 20,
			Handler:        h,
		},
		graceTimeout: defaultGraceTimeout,
	}

	if addr != "" {
		s.server.Addr = addr
	}

	return s
}

func (s *Server) Serve() {
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sig

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), s.graceTimeout)
		defer shutdownCtxCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("[ERROR] Graceful shutdown timed out! Forcing exit.")
			}
		}()

		err := s.server.Shutdown(context.Background())
		if err != nil {
			log.Fatal("[ERROR] Server shutdown error - ", err)
		}
	}()

	log.Printf("[INFO] starting HTTP-server at %s\n", s.server.Addr)
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("[ERROR] Failed to run HTTP-server", err)
	}
}
