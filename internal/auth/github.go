package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GitHubOAuth struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	HTTPClient   *http.Client
}

type GitHubUser struct {
	GitHubID  string `json:"github_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	APIToken  string `json:"api_token"`
}

func NewGitHubOAuth(clientID, clientSecret, redirectURL string, client *http.Client) *GitHubOAuth {
	return &GitHubOAuth{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL, HTTPClient: client}
}

func (g *GitHubOAuth) Enabled() bool {
	return strings.TrimSpace(g.ClientID) != "" && strings.TrimSpace(g.ClientSecret) != ""
}

func (g *GitHubOAuth) AuthURL(state string) (string, error) {
	if !g.Enabled() {
		return "", fmt.Errorf("github oauth is not configured")
	}
	values := url.Values{}
	values.Set("client_id", g.ClientID)
	values.Set("redirect_uri", g.RedirectURL)
	values.Set("scope", "read:user user:email")
	values.Set("state", state)
	return "https://github.com/login/oauth/authorize?" + values.Encode(), nil
}

func (g *GitHubOAuth) ExchangeAndFetchUser(ctx context.Context, code string) (*GitHubUser, error) {
	if !g.Enabled() {
		return nil, fmt.Errorf("github oauth is not configured")
	}
	token, err := g.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}
	user, err := g.fetchUser(ctx, token)
	if err != nil {
		return nil, err
	}
	user.APIToken = token
	return user, nil
}

func (g *GitHubOAuth) exchangeCode(ctx context.Context, code string) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"client_id":     g.ClientID,
		"client_secret": g.ClientSecret,
		"code":          code,
		"redirect_uri":  g.RedirectURL,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var decoded struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	if decoded.Error != "" {
		return "", fmt.Errorf("github token exchange failed: %s", decoded.Description)
	}
	if decoded.AccessToken == "" {
		return "", fmt.Errorf("github token exchange returned empty token")
	}
	return decoded.AccessToken, nil
}

func (g *GitHubOAuth) fetchUser(ctx context.Context, token string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github user request failed: status %d", resp.StatusCode)
	}
	var decoded struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	return &GitHubUser{
		GitHubID:  fmt.Sprintf("%d", decoded.ID),
		Username:  decoded.Login,
		Email:     decoded.Email,
		AvatarURL: decoded.AvatarURL,
	}, nil
}
