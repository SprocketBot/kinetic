package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"strings"
)

const (
	discordTokenURL   = "https://discord.com/api/oauth2/token"
	discordUserURL    = "https://discord.com/api/users/@me"
	discordAuthorizeURL = "https://discord.com/api/oauth2/authorize"
)

// discordTokenResponse is the JSON body returned by Discord's token endpoint.
type discordTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// discordUserResponse is the JSON body returned by Discord's /users/@me endpoint.
type discordUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// discordBuildAuthorizeURL constructs the Discord OAuth2 authorization redirect URL.
func discordBuildAuthorizeURL(clientID, redirectURL, state string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "identify")
	params.Set("state", state)
	return discordAuthorizeURL + "?" + params.Encode()
}

// discordExchangeCode exchanges an OAuth2 authorization code for a Discord access token.
func discordExchangeCode(ctx context.Context, clientID, clientSecret, redirectURL, code string) (string, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURL)

	req, err := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodPost, discordTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("discord token exchange: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("discord token exchange: http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("discord token exchange: read body: %w", err)
	}

	var tokenResp discordTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("discord token exchange: decode response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("discord token exchange: %s: %s", tokenResp.Error, tokenResp.ErrorDesc)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("discord token exchange: empty access token (status %d)", resp.StatusCode)
	}

	return tokenResp.AccessToken, nil
}

// discordGetCurrentUser retrieves the Discord user ID and username using the given access token.
func discordGetCurrentUser(ctx context.Context, accessToken string) (id, username string, err error) {
	req, err := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodGet, discordUserURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("discord get user: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("discord get user: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != stdhttp.StatusOK {
		return "", "", fmt.Errorf("discord get user: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("discord get user: read body: %w", err)
	}

	var userResp discordUserResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return "", "", fmt.Errorf("discord get user: decode response: %w", err)
	}

	if userResp.ID == "" {
		return "", "", fmt.Errorf("discord get user: empty user ID in response")
	}

	return userResp.ID, userResp.Username, nil
}
