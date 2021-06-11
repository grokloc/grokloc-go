package app

import (
	"io"
	"net/http"
)

// Client is an app client
type Client struct {
	Host      string // Without trailing /
	ID        string
	APISecret string
	h         *http.Client
	Token     *Token
}

// NewClient returns a new Client instance
func NewClient(host, id, apiSecret string) (*Client, error) {
	return &Client{
		Host:      host,
		ID:        id,
		APISecret: apiSecret,
		h:         &http.Client{},
	}, nil
}

// Ok calls the /ok endpoint
func (c *Client) Ok() (*http.Response, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.Host+OkRoute, nil)
	if err != nil {
		return nil, nil, err
	}
	return c.makeRequest(req)
}

// makeRequest is a convenience wrapper to run requests
func (c *Client) makeRequest(req *http.Request) (*http.Response, []byte, error) {
	resp, err := c.h.Do(req)
	if err != nil {
		return nil, nil, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	return resp, respBody, nil
}
