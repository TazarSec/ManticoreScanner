package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/etsubu/manticore-scanner/pkg/api"
	"github.com/etsubu/manticore-scanner/pkg/vcs"
)

var _ vcs.Provider = (*Provider)(nil)

const commentMarker = "<!-- manticore-scanner-results -->"

// Provider implements the vcs.Provider interface for GitHub.
type Provider struct {
	httpClient *http.Client
}

// NewProvider creates a new GitHub VCS provider.
func NewProvider(httpClient *http.Client) *Provider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Provider{httpClient: httpClient}
}

func (p *Provider) Name() string { return "github" }

func (p *Provider) Detect() (*vcs.Context, error) {
	return Detect()
}

func (p *Provider) PostResults(ctx context.Context, vcsCtx *vcs.Context, results []api.BatchResultItem) error {
	body := buildCommentBody(results)
	return p.postOrUpdateComment(ctx, vcsCtx, body)
}

type ghComment struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type ghCommentRequest struct {
	Body string `json:"body"`
}

func (p *Provider) postOrUpdateComment(ctx context.Context, vcsCtx *vcs.Context, body string) error {
	apiBase := "https://api.github.com"

	// Find existing comment.
	listURL := fmt.Sprintf("%s/repos/%s/issues/%d/comments", apiBase, vcsCtx.Repository, vcsCtx.PRNumber)
	existingID, err := p.findExistingComment(ctx, listURL, vcsCtx.Token)
	if err != nil {
		return fmt.Errorf("finding existing comment: %w", err)
	}

	reqBody, err := json.Marshal(ghCommentRequest{Body: body})
	if err != nil {
		return fmt.Errorf("marshalling comment: %w", err)
	}

	var method, url string
	if existingID > 0 {
		method = http.MethodPatch
		url = fmt.Sprintf("%s/repos/%s/issues/comments/%d", apiBase, vcsCtx.Repository, existingID)
	} else {
		method = http.MethodPost
		url = listURL
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+vcsCtx.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("posting comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (p *Provider) findExistingComment(ctx context.Context, listURL, token string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, nil // Can't list comments, will create a new one.
	}

	var comments []ghComment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return 0, nil
	}

	for _, c := range comments {
		if strings.Contains(c.Body, commentMarker) {
			return c.ID, nil
		}
	}

	return 0, nil
}
