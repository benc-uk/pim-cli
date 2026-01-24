package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

const (
	graphAPIScope   = "https://graph.microsoft.com/.default"
	graphAPIBaseURL = "https://graph.microsoft.com/beta"
)

// Client wraps HTTP client with Azure authentication for Microsoft Graph API
type Client struct {
	cred       azcore.TokenCredential
	httpClient *http.Client
}

// NewClient creates a new Graph API client with the given Azure credential
func NewClient(cred azcore.TokenCredential) *Client {
	return &Client{
		cred:       cred,
		httpClient: http.DefaultClient,
	}
}

// Request performs an authenticated request to the Microsoft Graph API
func (c *Client) Request(ctx context.Context, method, url string, body []byte, result any) error {
	token, err := c.cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{graphAPIScope},
	})
	if err != nil {
		return fmt.Errorf("failed to get Graph API token: %w", err)
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Graph API: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("graph API error: %s - %s", resp.Status, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
