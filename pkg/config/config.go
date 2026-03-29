package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Owner          string
	Repo           string
	Label          string
	State          string
	GitHubToken    string
	Port           string
	SiteTitle      string
	PrimaryColor   string
	SecondaryColor string
}

// Load loads configuration from environment variables or defaults
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		Owner:          getEnv("GITHUB_OWNER", "getarcaneapp"),
		Repo:           getEnv("GITHUB_REPO", "arcane"),
		Label:          getEnv("GITHUB_LABEL", "feature"),
		State:          getEnv("GITHUB_STATE", "open"),
		Port:           getEnv("PORT", "8080"),
		SiteTitle:      getEnv("SITE_TITLE", "Issue Board"),
		PrimaryColor:   getEnv("PRIMARY_COLOR", "#2563eb"),
		SecondaryColor: getEnv("SECONDARY_COLOR", "#64748b"),
	}

	// Try to get token from environment first
	token := os.Getenv("GITHUB_TOKEN")
	
	// If not in env, try gh CLI
	if token == "" {
		token = getGHToken()
	}

	config.GitHubToken = token

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getGHToken attempts to get the GitHub token from gh CLI
func getGHToken() string {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Owner == "" {
		return fmt.Errorf("GITHUB_OWNER is required")
	}
	if c.Repo == "" {
		return fmt.Errorf("GITHUB_REPO is required")
	}
	return nil
}

// RepoURL returns the full GitHub repository URL
func (c *Config) RepoURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", c.Owner, c.Repo)
}
