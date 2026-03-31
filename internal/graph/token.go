package graph

import (
	"context"
	"fmt"
	"net/http"

	"github.com/frankhildebrandt/teams2issue/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type TokenProvider struct {
	cfg        clientcredentials.Config
	httpClient *http.Client
}

func NewTokenProvider(appCfg config.Config, httpClient *http.Client) (*TokenProvider, error) {
	if appCfg.OAuth2.ClientID == "" || appCfg.OAuth2.ClientSecret == "" || appCfg.OAuth2.TokenURL == "" {
		return nil, fmt.Errorf("oauth2 config incomplete (client_id/client_secret/token_url required)")
	}

	cc := clientcredentials.Config{
		ClientID:     appCfg.OAuth2.ClientID,
		ClientSecret: appCfg.OAuth2.ClientSecret,
		TokenURL:     appCfg.OAuth2.TokenURL,
		Scopes:       appCfg.OAuth2.Scopes,
	}

	// Ensure we use the app's HTTP client (timeouts, proxy, etc.).
	cc.AuthStyle = oauth2.AuthStyleInHeader

	return &TokenProvider{cfg: cc, httpClient: httpClient}, nil
}

func (p *TokenProvider) Token(ctx context.Context) (*oauth2.Token, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.httpClient) //nolint:revive,staticcheck // oauth2 expects context value
	return p.cfg.Token(ctx)
}

