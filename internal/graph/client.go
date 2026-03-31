package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
	tokens  *TokenProvider
}

func NewClient(httpClient *http.Client, tokens *TokenProvider) *Client {
	return &Client{
		baseURL: "https://graph.microsoft.com/v1.0",
		http:    httpClient,
		tokens:  tokens,
	}
}

func (c *Client) DoJSON(ctx context.Context, method string, path string, in any, out any) (*http.Response, error) {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(b)
	}

	url := strings.TrimRight(c.baseURL, "/") + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	tok, err := c.tokens.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graph request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return resp, fmt.Errorf("graph %s %s failed: status=%d body=%s", method, path, resp.StatusCode, string(b))
	}

	if out == nil {
		return resp, nil
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(out); err != nil {
		return resp, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
	}
}

