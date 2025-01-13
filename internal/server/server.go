package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"Marriner/internal/database"
)

type Server struct {
	port    int
	Catalog Catalog
	db      database.Service
	Server  *http.Server
}

func NewServer() *Server {
	port := 8000
	catalog, err := NewCatalog("catalog")
	if err != nil {
		log.Println(fmt.Errorf("server failed to create catalog: %w", err))
	}
	NewServer := &Server{
		port:    port,
		Catalog: *catalog,
		db:      database.New(),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	NewServer.Server = server
	return NewServer
}
