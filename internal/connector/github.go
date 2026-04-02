package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/crypto"
	"golang.org/x/oauth2"
)

type GitHubConnector struct {
	integrationRepo storage.IntegrationRepository
	encryptionKey   string
}

func NewGitHubConnector(
	integrationRepo storage.IntegrationRepository,
	encryptionKey string,
) *GitHubConnector {
	return &GitHubConnector{
		integrationRepo: integrationRepo,
		encryptionKey:   encryptionKey,
	}
}

func (gc *GitHubConnector) Name() string {
	return "github"
}

// readGitHubError is a helper to parse error messages from the GitHub API.
func readGitHubError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("github api error: status %s (could not read error body)", resp.Status)
	}
	return fmt.Errorf("github api error: status %s, body: %s", resp.Status, string(body))
}

func (gc *GitHubConnector) Execute(
	ctx context.Context,
	config map[string]interface{},
	prevResults map[string]interface{},
) (interface{}, error) {
	action, _ := config["action"].(string)

	switch action {
	case "create_issue":
		return gc.createIssue(ctx, config)
	case "create_pull_request":
		return gc.createPullRequest(ctx, config)
	case "add_comment":
		return gc.addComment(ctx, config)
	default:
		return nil, fmt.Errorf("hành động không tồn tại: %s", action)
	}
}

// getAccessToken is a helper to retrieve and decrypt the token for a given integration.
func (gc *GitHubConnector) getAccessToken(ctx context.Context, integrationID string) (string, error) {
	integration, err := gc.integrationRepo.GetByID(ctx, integrationID)
	if err != nil || integration == nil {
		return "", fmt.Errorf("github: failed to get integration with id %s: %w", integrationID, err)
	}

	decryptedTokenJSON, err := crypto.Decrypt(integration.AccessToken, gc.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("github: failed to decrypt access token: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(decryptedTokenJSON), &token); err != nil {
		return "", fmt.Errorf("github: failed to unmarshal access token: %w", err)
	}
	
	return token.AccessToken, nil
}


func (gc *GitHubConnector) createIssue(
	ctx context.Context,
	config map[string]interface{},
) (interface{}, error) {
	integrationID, _ := config["integration_id"].(string)
	accessToken, err := gc.getAccessToken(ctx, integrationID)
	if err != nil {
		return nil, err
	}

	repo, _ := config["repo"].(string)
	title, _ := config["title"].(string)
	body, _ := config["body"].(string)

	var labelStrs []string
	if labels, ok := config["labels"].([]interface{}); ok {
		for _, l := range labels {
			if label, ok := l.(string); ok {
				labelStrs = append(labelStrs, label)
			}
		}
	}
	
	payload := map[string]interface{}{"title": title, "body": body, "labels": labelStrs}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.github.com/repos/%s/issues", repo), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, readGitHubError(resp)
	}

	var result struct {
		Number int    `json:"number"`
		URL    string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode github response: %w", err)
	}

	return result, nil
}

func (gc *GitHubConnector) createPullRequest(
	ctx context.Context,
	config map[string]interface{},
) (interface{}, error) {
	integrationID, _ := config["integration_id"].(string)
	accessToken, err := gc.getAccessToken(ctx, integrationID)
	if err != nil {
		return nil, err
	}

	repo, _ := config["repo"].(string)
	title, _ := config["title"].(string)
	head, _ := config["head"].(string)
	base, _ := config["base"].(string)
	body, _ := config["body"].(string)

	payload := map[string]interface{}{"title": title, "head": head, "base": base, "body": body}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.github.com/repos/%s/pulls", repo), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, readGitHubError(resp)
	}

	var result struct {
		Number int    `json:"number"`
		URL    string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode github response: %w", err)
	}

	return map[string]interface{}{
		"pr_number": result.Number,
		"url":       result.URL,
	}, nil
}

func (gc *GitHubConnector) addComment(
	ctx context.Context,
	config map[string]interface{},
) (interface{}, error) {
	integrationID, _ := config["integration_id"].(string)
	accessToken, err := gc.getAccessToken(ctx, integrationID)
	if err != nil {
		return nil, err
	}

	repo, _ := config["repo"].(string)
	text, _ := config["text"].(string)
	
	issueNumberVal, exists := config["issue_number"]
	if !exists {
		return nil, fmt.Errorf("github: 'issue_number' is required for add_comment")
	}

	var issueNumber float64
	switch v := issueNumberVal.(type) {
	case float64:
		issueNumber = v
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("github: failed to parse 'issue_number' string '%s' as number: %w", v, err)
		}
		issueNumber = parsed
	default:
		return nil, fmt.Errorf("github: 'issue_number' must be a number or a string, but got %T", v)
	}

	payload := map[string]interface{}{"body": text}
	data, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%.0f/comments", repo, issueNumber)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, readGitHubError(resp)
	}

	return map[string]interface{}{"success": true}, nil
}

func (gc *GitHubConnector) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["action"].(string); !ok {
		return fmt.Errorf("thiếu trường action")
	}
	return nil
}
