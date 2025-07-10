package keycloak

import (
	"context"
	"encoding/json"
	"factory/internal/config"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Nerzal/gocloak/v13"
)

type Client struct {
	gocloak      *gocloak.GoCloak
	config       *config.Config
	clientID     string
	clientSecret string
	realm        string
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		gocloak:      gocloak.NewClient(cfg.KeycloakInternalURL),
		config:       cfg,
		clientID:     cfg.KeycloakClientID,
		clientSecret: cfg.KeycloakClientSecret,
		realm:        cfg.KeycloakRealm,
	}
}

func (c *Client) VerifyToken(ctx context.Context, token string) (*gocloak.UserInfo, error) {
	userInfo, err := c.gocloak.GetUserInfo(ctx, token, c.realm)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (c *Client) GetLoginURL(redirectURI string) string {
	// Строим URL авторизации для браузера (используем внешний URL)
	baseURL := c.config.KeycloakURL
	authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth", baseURL, c.realm)

	params := url.Values{}
	params.Add("client_id", c.clientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "openid profile email")

	return fmt.Sprintf("%s?%s", authURL, params.Encode())
}

func (c *Client) ExchangeCodeForToken(ctx context.Context, code, redirectURI string) (*gocloak.JWT, error) {
	// Выполняем HTTP запрос для обмена кода на токен (используем внутренний URL)
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.config.KeycloakInternalURL, c.realm)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var tokenResp gocloak.JWT
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}
