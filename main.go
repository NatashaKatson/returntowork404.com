package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"whatdidimiss/cache"
	"whatdidimiss/claude"
	"whatdidimiss/handlers"
)

func main() {
	// Load configuration from environment
	port := getEnv("PORT", "8080")
	claudeAPIKey := os.Getenv("CLAUDE_API_KEY")

	if claudeAPIKey == "" {
		log.Fatal("CLAUDE_API_KEY environment variable is required")
	}

	// Initialize in-memory cache
	memCache := cache.NewMemoryCache()
	defer memCache.Close()

	// Initialize Claude client
	claudeClient := claude.NewClient(claudeAPIKey)

	// Initialize API handler
	apiHandler := handlers.NewAPIHandler(memCache, claudeClient)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Post("/catchup", apiHandler.CatchUp)
		r.Get("/health", apiHandler.Health)
	})

	// Static files (served by reproxy in production, but useful for local dev)
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/*", fileServer)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
