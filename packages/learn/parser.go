package learn

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type URLType int

const (
	URLTypePR URLType = iota
	URLTypeCommit
	URLTypeUnknown
)

type ParsedURL struct {
	Type      URLType
	Owner     string
	Repo      string
	Number    int
	CommitSHA string
}

type FileDiff struct {
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Patch     string `json:"patch"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

type DiffResult struct {
	Title       string
	Description string
	Files       []FileDiff
}

var (
	prRegex     = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)`)
	commitRegex = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/commit/([a-f0-9]+)`)
)

func ParseURL(url string) (*ParsedURL, error) {
	if matches := prRegex.FindStringSubmatch(url); matches != nil {
		number, _ := strconv.Atoi(matches[3])
		return &ParsedURL{
			Type:   URLTypePR,
			Owner:  matches[1],
			Repo:   matches[2],
			Number: number,
		}, nil
	}

	if matches := commitRegex.FindStringSubmatch(url); matches != nil {
		return &ParsedURL{
			Type:      URLTypeCommit,
			Owner:     matches[1],
			Repo:      matches[2],
			CommitSHA: matches[3],
		}, nil
	}

	return nil, fmt.Errorf("unsupported URL format: %s\nSupported formats:\n  - https://github.com/owner/repo/pull/123\n  - https://github.com/owner/repo/commit/abc1234", url)
}

func FetchDiff(parsed *ParsedURL) (*DiffResult, error) {
	switch parsed.Type {
	case URLTypePR:
		return fetchPRDiff(parsed)
	case URLTypeCommit:
		return fetchCommitDiff(parsed)
	default:
		return nil, fmt.Errorf("unknown URL type")
	}
}

func fetchPRDiff(parsed *ParsedURL) (*DiffResult, error) {
	// Get PR metadata
	metaCmd := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/pulls/%d", parsed.Owner, parsed.Repo, parsed.Number),
		"--jq", "{title: .title, body: .body}")
	metaOutput, err := metaCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR metadata: %w", err)
	}

	var meta struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.Unmarshal(metaOutput, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse PR metadata: %w", err)
	}

	// Get PR files
	filesCmd := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/pulls/%d/files", parsed.Owner, parsed.Repo, parsed.Number))
	filesOutput, err := filesCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR files: %w", err)
	}

	var files []FileDiff
	if err := json.Unmarshal(filesOutput, &files); err != nil {
		return nil, fmt.Errorf("failed to parse PR files: %w", err)
	}

	return &DiffResult{
		Title:       meta.Title,
		Description: meta.Body,
		Files:       files,
	}, nil
}

func fetchCommitDiff(parsed *ParsedURL) (*DiffResult, error) {
	cmd := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/commits/%s", parsed.Owner, parsed.Repo, parsed.CommitSHA))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commit: %w", err)
	}

	var commit struct {
		Commit struct {
			Message string `json:"message"`
		} `json:"commit"`
		Files []FileDiff `json:"files"`
	}
	if err := json.Unmarshal(output, &commit); err != nil {
		return nil, fmt.Errorf("failed to parse commit: %w", err)
	}

	// Split commit message into title and body
	parts := strings.SplitN(commit.Commit.Message, "\n", 2)
	title := parts[0]
	body := ""
	if len(parts) > 1 {
		body = strings.TrimSpace(parts[1])
	}

	return &DiffResult{
		Title:       title,
		Description: body,
		Files:       commit.Files,
	}, nil
}

func (d *DiffResult) FormatForPrompt() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Title: %s\n", d.Title))
	if d.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", d.Description))
	}
	sb.WriteString("\nFiles changed:\n")

	for _, f := range d.Files {
		sb.WriteString(fmt.Sprintf("\n--- %s (%s, +%d -%d) ---\n", f.Filename, f.Status, f.Additions, f.Deletions))
		if f.Patch != "" {
			// Truncate very long patches
			patch := f.Patch
			if len(patch) > 2000 {
				patch = patch[:2000] + "\n... (truncated)"
			}
			sb.WriteString(patch)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
