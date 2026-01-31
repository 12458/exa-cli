package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	baseURL    = "https://api.exa.ai"
	apiKeyEnv  = "EXA_API_KEY"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func New(apiKey string) (*Client, error) {
	if apiKey == "" {
		apiKey = os.Getenv(apiKeyEnv)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key required: set %s environment variable or use --api-key flag", apiKeyEnv)
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Error != "" {
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Error)
		}
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// Search performs a web search using Exa
func (c *Client) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	var result SearchResponse
	if err := c.doRequest(ctx, http.MethodPost, "/search", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetContents retrieves content from URLs
func (c *Client) GetContents(ctx context.Context, req *ContentsRequest) (*ContentsResponse, error) {
	var result ContentsResponse
	if err := c.doRequest(ctx, http.MethodPost, "/contents", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

