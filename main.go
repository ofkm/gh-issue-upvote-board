package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"go.ofkm.dev/gh-issue-upvote-board/pkg/config"
	"go.ofkm.dev/gh-issue-upvote-board/pkg/github"
	"go.ofkm.dev/gh-issue-upvote-board/pkg/handlers"
)

//go:embed static/*
var staticFS embed.FS

//go:embed templates/*
var templatesFS embed.FS

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Log configuration
	log.Printf("Starting GitHub Issue Upvote Board")
	log.Printf("Repository: %s/%s", cfg.Owner, cfg.Repo)
	if cfg.Label != "" {
		log.Printf("Label filter: %s", cfg.Label)
	}
	log.Printf("State filter: %s", cfg.State)
	if cfg.GitHubToken != "" {
		log.Printf("Using authenticated GitHub API requests")
	} else {
		log.Printf("Warning: No GitHub token found. API rate limits will be lower.")
		log.Printf("Set GITHUB_TOKEN env var or authenticate with 'gh auth login'")
	}

	// Create GitHub client
	client := github.NewClient(cfg.GitHubToken)

	// Create handlers
	handler, err := handlers.New(cfg, client, templatesFS)
	if err != nil {
		log.Fatalf("Failed to create handlers: %v", err)
	}

	// Set up routes
	mux := http.NewServeMux()

	// Serve static files from embedded FS
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub filesystem: %v", err)
	}
	fileServer := http.FileServer(http.FS(staticSub))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Application routes
	mux.HandleFunc("/issues", handler.IssueList)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handler.Home(w, r)
		} else {
			handler.IssueDetail(w, r)
		}
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
