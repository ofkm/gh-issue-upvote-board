package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"
)

// Issue represents a GitHub issue with reactions
type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Reactions Reactions `json:"reactions"`
	Labels    []Label   `json:"labels"`
}

// User represents a GitHub user
type User struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

// Label represents a GitHub label
type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Reactions represents GitHub issue reactions
type Reactions struct {
	PlusOne int `json:"+1"`
	Total   int `json:"total_count"`
}

// Client represents a GitHub API client
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a new GitHub API client
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PaginatedIssues represents a paginated list of issues
type PaginatedIssues struct {
	Issues      []Issue
	CurrentPage int
	TotalPages  int
	PerPage     int
	TotalCount  int
	HasNext     bool
	HasPrev     bool
}

// FetchIssues fetches issues from a GitHub repository with pagination
func (c *Client) FetchIssues(owner, repo, label, state string, page, perPage int) (*PaginatedIssues, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 30
	}

	// First, fetch all issues to sort by upvotes
	var allIssues []Issue
	fetchPage := 1
	fetchPerPage := 100

	for {
		issues, hasMore, err := c.fetchIssuesPage(owner, repo, label, state, fetchPage, fetchPerPage)
		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, issues...)

		if !hasMore {
			break
		}
		fetchPage++
	}

	// Sort by upvotes (reactions +1)
	sort.Slice(allIssues, func(i, j int) bool {
		return allIssues[i].Reactions.PlusOne > allIssues[j].Reactions.PlusOne
	})

	// Calculate pagination
	totalCount := len(allIssues)
	totalPages := (totalCount + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	// Get the slice for the current page
	start := (page - 1) * perPage
	end := start + perPage
	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	pageIssues := allIssues[start:end]

	return &PaginatedIssues{
		Issues:      pageIssues,
		CurrentPage: page,
		TotalPages:  totalPages,
		PerPage:     perPage,
		TotalCount:  totalCount,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
	}, nil
}

func (c *Client) fetchIssuesPage(owner, repo, label, state string, page, perPage int) ([]Issue, bool, error) {
	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	
	params := url.Values{}
	if label != "" {
		params.Add("labels", label)
	}
	if state != "" {
		params.Add("state", state)
	} else {
		params.Add("state", "open")
	}
	params.Add("per_page", fmt.Sprintf("%d", perPage))
	params.Add("page", fmt.Sprintf("%d", page))

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, false, err
	}

	// Add headers
	req.Header.Add("Accept", "application/vnd.github.squirrel-girl-preview+json")
	if c.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, false, err
	}

	// If we got a full page, there might be more
	hasMore := len(issues) == perPage

	return issues, hasMore, nil
}

// GetIssue fetches a single issue by number
func (c *Client) GetIssue(owner, repo string, number int) (*Issue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, number)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/vnd.github.squirrel-girl-preview+json")
	if c.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(body))
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil
}
