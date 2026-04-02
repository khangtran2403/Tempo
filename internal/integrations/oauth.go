package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"net/http"
	"time"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/crypto"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type OAuthManager struct {
	storage       storage.IntegrationRepository
	encryptionKey string
	config        *config.Config
}

func NewOAuthManager(storage storage.IntegrationRepository, encryptionKey string, config *config.Config) *OAuthManager {
	return &OAuthManager{
		storage:       storage,
		encryptionKey: encryptionKey,
		config:        config,
	}
}
func (om *OAuthManager) GetURL(state string, provider string) (string, error) {
	config := om.getOAuth2Config(provider)
	if config == nil {
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, nil
}

// Hàm này xử lý callback từ OAuth provider sau khi người dùng đăng nhập
func (om *OAuthManager) HandleCallback(ctx context.Context, provider string, code string, userID string) (*domain.Integration, error) {
	config := om.getOAuth2Config(provider)
	// doi code lay token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	encryptedAccessToken, err := crypto.Encrypt(token.AccessToken, om.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt token: %w", err)
	}

	encryptedRefreshToken := ""
	if token.RefreshToken != "" {
		encryptedRefreshToken, _ = crypto.Encrypt(token.RefreshToken, om.encryptionKey)
	}

	// Lay metadata tu provider
	metadata, err := om.getProviderMetadata(ctx, provider, token.AccessToken)
	if err != nil {
		metadata = make(map[string]interface{})
	}

	integration := &domain.Integration{
		ID:           uuid.New().String(),
		UserID:       userID,
		Provider:     provider,
		Name:         metadata["name"].(string),
		AccessToken:  encryptedAccessToken,
		RefreshToken: encryptedRefreshToken,
		IsActive:     true,
		Metadata:     metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if token.Expiry != (time.Time{}) {
		integration.TokenExpiresAt = &token.Expiry
	}

	err = om.storage.Create(ctx, integration)
	if err != nil {
		return nil, err
	}

	return integration, nil
}
func (om *OAuthManager) RefreshToken(
	ctx context.Context,
	integration *domain.Integration,
) error {
	if integration.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	config := om.getOAuth2Config(integration.Provider)
	if config == nil {
		return fmt.Errorf("provider not supported")
	}

	refreshToken, err := crypto.Decrypt(integration.RefreshToken, om.encryptionKey)
	if err != nil {
		return err
	}

	// Refresh token
	tokenSource := config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update with new token
	encryptedAccessToken, _ := crypto.Encrypt(newToken.AccessToken, om.encryptionKey)
	integration.AccessToken = encryptedAccessToken

	if newToken.Expiry != (time.Time{}) {
		integration.TokenExpiresAt = &newToken.Expiry
	}

	return om.storage.Update(ctx, integration)
}

func (om *OAuthManager) getOAuth2Config(provider string) *oauth2.Config {
	switch provider {
	case "slack":
		return &oauth2.Config{
			ClientID:     om.config.Slack.ClientID,
			ClientSecret: om.config.Slack.ClientSecret,
			RedirectURL:  om.config.OAuth.RedirectURL + "/slack",
			Scopes: []string{
				"chat:write",
				"channels:read",
				"users:read",
			},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://slack.com/oauth_authorize",
				TokenURL: "https://slack.com/api/oauth.v2.access",
			},
		}

	case "github":
		return &oauth2.Config{
			ClientID:     om.config.GitHub.ClientID,
			ClientSecret: om.config.GitHub.ClientSecret,
			RedirectURL:  om.config.OAuth.RedirectURL + "/github",
			Scopes: []string{
				"repo",
				"read:user",
			},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
		}

	case "discord":
		return &oauth2.Config{
			ClientID:     om.config.Discord.ClientID,
			ClientSecret: om.config.Discord.ClientSecret,
			RedirectURL:  om.config.OAuth.RedirectURL + "/discord",
			Scopes: []string{
				"identify",
				"guilds",
				"messages.read",
			},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://discord.com/api/oauth2/authorize",
				TokenURL: "https://discord.com/api/oauth2/token",
			},
		}

	default:
		return nil
	}
}
func (om *OAuthManager) getProviderMetadata(
	ctx context.Context,
	provider string,
	accessToken string,
) (map[string]interface{}, error) {
	switch provider {
	case "slack":
		return om.getSlackMetadata(ctx, accessToken)
	case "github":
		return om.getGitHubMetadata(ctx, accessToken)
	case "discord":
		return om.getDiscordMetadata(ctx, accessToken)
	default:
		return map[string]interface{}{}, nil
	}
}

func (om *OAuthManager) getSlackMetadata(
	ctx context.Context,
	accessToken string,
) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://slack.com/api/auth.test", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK       bool   `json:"ok"`
		URL      string `json:"url"`
		TeamID   string `json:"team_id"`
		TeamName string `json:"team_name"`
		UserID   string `json:"user_id"`
		UserName string `json:"user_name"`
	}

	json.NewDecoder(resp.Body).Decode(&result)
	if result.TeamName == "" || result.TeamID == "" || result.UserID == "" {
		return nil, fmt.Errorf("missing required fields in Slack response")
	}

	return map[string]interface{}{
		"name":      result.TeamName,
		"team_id":   result.TeamID,
		"user_id":   result.UserID,
		"workspace": result.URL,
	}, nil
}

func (om *OAuthManager) getGitHubMetadata(
	ctx context.Context,
	accessToken string,
) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user struct {
		Login string `json:"login"`
		ID    int    `json:"id"`
		Name  string `json:"name"`
	}

	json.NewDecoder(resp.Body).Decode(&user)

	return map[string]interface{}{
		"name":    user.Name,
		"login":   user.Login,
		"user_id": user.ID,
	}, nil
}

func (om *OAuthManager) getDiscordMetadata(
	ctx context.Context,
	accessToken string,
) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user struct {
		ID         int         `json:"id"`
		UserName   string      `json:"username"`
		GlobalName string      `json:"global_name"`
		Avatar     image.Image `json:"avatar"`
		Email      string      `json:"email"`
	}

	json.NewDecoder(resp.Body).Decode(&user)

	return map[string]interface{}{
		"name":        user.UserName,
		"global_name": user.GlobalName,
		"user_id":     user.ID,
		"avatar":      user.Avatar,
		"email":       user.Email,
	}, nil

}
