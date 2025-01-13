package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"Marriner/cmd/web"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", s.HelloWorldHandler)

	r.Get("/health", s.healthHandler)

	fileServer := http.FileServer(http.FS(web.Files))
	r.Handle("/assets/*", fileServer)
	r.Get("/web", templ.Handler(web.CatalogList(s.CatalogItemsMapping())).ServeHTTP)
	r.Post("/hello", web.HelloWebHandler)
	r.Get("/tasks", s.TaskHandler)

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) CatalogItemsMapping() []web.CatalogItem {
	servers, err := s.Catalog.GetGameServers()
	if err != nil {
		log.Println(fmt.Errorf("getting games servers: %w", err))
	}
	catalog := []web.CatalogItem{}
	for _, server := range servers {
		item := web.CatalogItem{
			Title:       server.Name,
			Description: server.Image,
			Type:        "Service",
			Owner:       "Cube-go",
			Tags:        []string{"golang", "Server", "api", server.HealthCheck},
		}
		catalog = append(catalog, item)
	}
	return catalog
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}

func (s *Server) TaskHandler(w http.ResponseWriter, r *http.Request) {
	servers, err := s.Catalog.GetGameServers()
	if err != nil {
		log.Println(fmt.Errorf("getting games servers: %w", err))
	}
	jsonResp, err := json.Marshal(servers)
	if err != nil {
		log.Println(fmt.Errorf("marshalling json: %w", err))
	}

	_, _ = w.Write(jsonResp)
}
