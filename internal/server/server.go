package server

import (
	"log"
	"net/http"
	"time"
)

type Server struct {
	Server *http.Server
}

func New(h http.Handler, addr string) *Server {
	const (
		defaultAddress = ":8080"
	)
	s := &Server{
		Server: &http.Server{
			Addr:           defaultAddress,
			ReadTimeout:    time.Second * 10,
			WriteTimeout:   time.Second * 10,
			IdleTimeout:    time.Second * 10,
			MaxHeaderBytes: 1 << 20,
			Handler:        h,
		},
	}

	if addr != "" {
		s.Server.Addr = addr
	}

	return s
}

func (s *Server) Serve() {
	log.Printf("[INFO] starting HTTP-server at %s\n", s.Server.Addr)
	err := s.Server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("[ERROR] Failed to run HTTP-server", err)
	}
}
