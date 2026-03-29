package github

import "testing"

func TestExtractIssueNumbers(t *testing.T) {
	owner := "ofkm"
	repo := "gh-issue-upvote-board"
	text := `
Implements milestone badges.
Fixes #12
Related: ofkm/gh-issue-upvote-board#34
See also https://github.com/ofkm/gh-issue-upvote-board/issues/56
`

	issueNumbers := extractIssueNumbers(text, owner, repo)

	for _, issueNumber := range []int{12, 34, 56} {
		if _, ok := issueNumbers[issueNumber]; !ok {
			t.Fatalf("expected issue %d to be detected", issueNumber)
		}
	}
}

func TestExtractIssueNumbersIgnoresEmptyText(t *testing.T) {
	issueNumbers := extractIssueNumbers("", "ofkm", "gh-issue-upvote-board")
	if len(issueNumbers) != 0 {
		t.Fatalf("expected no issue numbers for empty text, got %d", len(issueNumbers))
	}
}

func TestApplyIssueFilters(t *testing.T) {
	issues := []Issue{
		{Number: 1},
		{Number: 2, Milestone: &Milestone{Title: "v1.0"}},
		{Number: 3, OpenPRs: []PullRequest{{Number: 101}}},
		{Number: 4, Milestone: &Milestone{Title: "v1.0"}, OpenPRs: []PullRequest{{Number: 102}}},
	}

	t.Run("milestone only", func(t *testing.T) {
		filtered := applyIssueFilters(issues, true, false)
		if len(filtered) != 2 || filtered[0].Number != 2 || filtered[1].Number != 4 {
			t.Fatalf("unexpected milestone filter result: %#v", filtered)
		}
	})

	t.Run("open pr only", func(t *testing.T) {
		filtered := applyIssueFilters(issues, false, true)
		if len(filtered) != 2 || filtered[0].Number != 3 || filtered[1].Number != 4 {
			t.Fatalf("unexpected PR filter result: %#v", filtered)
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		filtered := applyIssueFilters(issues, true, true)
		if len(filtered) != 1 || filtered[0].Number != 4 {
			t.Fatalf("unexpected combined filter result: %#v", filtered)
		}
	})
}
