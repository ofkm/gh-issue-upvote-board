# GitHub Issue Upvote Board

Simple Web UI built with Go, HTMX, and HTML templates that displays GitHub issues sorted by upvotes (👍 reactions).

Initially built this for a better way to view upvoted issues for Arcane. Figured others may like it as well :)

## Installation

1. Clone the repository:

```bash
git clone https://github.com/ofkm/gh-issue-upvote-board.git
cd gh-issue-upvote-board
```

2. Build the application:

```bash
go build -o gh-issue-upvote-board
```

## Configuration

The application can be configured using environment variables:

| Variable       | Description                   | Default                     |
| -------------- | ----------------------------- | --------------------------- |
| `GITHUB_OWNER` | Repository owner              | `getarcaneapp`              |
| `GITHUB_REPO`  | Repository name               | `arcane`                    |
| `GITHUB_LABEL` | Filter by label               | `needs more upvotes`        |
| `GITHUB_STATE` | Issue state (open/closed/all) | `open`                      |
| `GITHUB_TOKEN` | GitHub Personal Access Token  | (auto-detected from gh CLI) |
| `PORT`         | Server port                   | `8080`                      |

### Authentication

The application automatically tries to get your GitHub token from:

1. `GITHUB_TOKEN` environment variable
2. GitHub CLI (`gh auth token`)

Without authentication, you'll be limited by GitHub's unauthenticated API rate limits (60 requests/hour).

To authenticate with GitHub CLI:

```bash
gh auth login
```

Or set a token manually:

```bash
export GITHUB_TOKEN=your_token_here
```

## Usage

### Run with defaults

```bash
./issue-upvote-board
```

Then open http://localhost:8080 in your browser.

### Run with custom configuration

```bash
GITHUB_OWNER=facebook \
GITHUB_REPO=react \
GITHUB_LABEL="Type: Bug" \
GITHUB_STATE=open \
PORT=3000 \
./issue-upvote-board
```
