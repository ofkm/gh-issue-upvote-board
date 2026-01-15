package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"go.ofkm.dev/gh-issue-upvote-board/pkg/config"
	"go.ofkm.dev/gh-issue-upvote-board/pkg/github"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	config    *config.Config
	client    *github.Client
	templates *template.Template
}

// New creates a new Handler
func New(cfg *config.Config, client *github.Client, templatesFS embed.FS) (*Handler, error) {
	// Parse templates with custom functions from embedded FS
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"markdown": markdownToHTML,
		"truncate": truncate,
		"add":      add,
		"sub":      sub,
	}).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Handler{
		config:    cfg,
		client:    client,
		templates: tmpl,
	}, nil
}

// Home renders the main page
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Config": h.config,
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// IssueList returns the list of issues (for HTMX)
func (h *Handler) IssueList(w http.ResponseWriter, r *http.Request) {
	// Get query parameters for dynamic filtering
	owner := r.URL.Query().Get("owner")
	repo := r.URL.Query().Get("repo")
	label := r.URL.Query().Get("label")
	state := r.URL.Query().Get("state")
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")

	// Use config defaults if not provided
	if owner == "" {
		owner = h.config.Owner
	}
	if repo == "" {
		repo = h.config.Repo
	}
	if label == "" {
		label = h.config.Label
	}
	if state == "" {
		state = h.config.State
	}

	// Parse pagination parameters
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 30
	if perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	paginatedIssues, err := h.client.FetchIssues(owner, repo, label, state, page, perPage)
	if err != nil {
		log.Printf("Error fetching issues: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching issues: %v", err), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Issues":      paginatedIssues.Issues,
		"Owner":       owner,
		"Repo":        repo,
		"Label":       label,
		"State":       state,
		"CurrentPage": paginatedIssues.CurrentPage,
		"TotalPages":  paginatedIssues.TotalPages,
		"TotalCount":  paginatedIssues.TotalCount,
		"PerPage":     paginatedIssues.PerPage,
		"HasNext":     paginatedIssues.HasNext,
		"HasPrev":     paginatedIssues.HasPrev,
	}

	if err := h.templates.ExecuteTemplate(w, "issue-list.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// IssueDetail returns detailed view of a single issue (for HTMX)
func (h *Handler) IssueDetail(w http.ResponseWriter, r *http.Request) {
	// Extract owner, repo, and issue number from URL
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	owner := parts[0]
	repo := parts[1]
	numberStr := parts[2]

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		http.Error(w, "Invalid issue number", http.StatusBadRequest)
		return
	}

	issue, err := h.client.GetIssue(owner, repo, number)
	if err != nil {
		log.Printf("Error fetching issue: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching issue: %v", err), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Issue": issue,
		"Owner": owner,
		"Repo":  repo,
	}

	if err := h.templates.ExecuteTemplate(w, "issue-detail.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// markdownToHTML converts GitHub-flavored markdown to HTML
func markdownToHTML(text string) template.HTML {
	if text == "" {
		return ""
	}

	// Create markdown parser with GitHub Flavored Markdown extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(text))

	// Create HTML renderer with unsafe mode to allow raw HTML
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	html := markdown.Render(doc, renderer)

	// Sanitize GitHub issue references (#123) and convert to links
	// This is a basic implementation - you could enhance it further
	htmlStr := string(html)

	return template.HTML(htmlStr)
}

// truncate truncates text to specified length
func truncate(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length] + "..."
}

// add adds two integers (for template pagination)
func add(a, b int) int {
	return a + b
}

// sub subtracts two integers (for template pagination)
func sub(a, b int) int {
	return a - b
}
