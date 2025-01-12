package web

import (
	"log"
	"net/http"
)

type CatalogItem struct {
	Title       string
	Description string
	Type        string
	Icon        string
	Owner       string
	Tags        []string
}

func CatalogHandler(w http.ResponseWriter, r *http.Request) {
	items := CatalogItems()
	component := CatalogList(items)
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error rendering in HelloWebHandler: %e", err)
	}
}

func CatalogItems() []CatalogItem {
	return []CatalogItem{
		{
			Title:       "User Service",
			Description: "Handles user authentication and management",
			Type:        "Service",
			Owner:       "Team Auth",
			Tags:        []string{"golang", "authentication", "api"},
		},
		{
			Title:       "Order Service",
			Description: "Handles orders of products",
			Type:        "Service",
			Owner:       "Team Consumer Sales",
			Tags:        []string{"golang", "Consumer", "api"},
		},
	}
}
