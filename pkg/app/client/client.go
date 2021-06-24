// Package client is an app client library
package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/grokloc/grokloc-go/pkg/app"
	"github.com/grokloc/grokloc-go/pkg/jwt"
	"github.com/grokloc/grokloc-go/pkg/security"
)

// Client is an app client
type Client struct {
	Host      string // Without trailing /
	ID        string
	APISecret string
	h         *http.Client
	token     *app.Token
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

// getToken retrieves the jwt for the calling user
func (c *Client) getToken() error {
	req, err := http.NewRequest(http.MethodPut, c.Host+app.TokenRoute, nil)
	if err != nil {
		return err
	}
	req.Header.Add(app.IDHeader, c.ID)
	req.Header.Add(app.TokenRequestHeader, security.EncodedSHA256(c.ID+c.APISecret))
	resp, body, err := c.makeRequest(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token response code: %d", resp.StatusCode)
	}
	if body == nil {
		return errors.New("token response body should be non-nil")
	}
	token := app.Token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		return err
	}
	c.token = &token
	return nil
}

// authedRequest will refresh the token for a regular user instance if it is nil
// or set to expire in 30 seconds
func (c *Client) authedRequest(req *http.Request) (*http.Response, []byte, error) {
	if c.token == nil || (c.token.Expires+30) > time.Now().Unix() {
		err := c.getToken()
		if err != nil {
			return nil, nil, err
		}
	}
	req.Header.Add(app.IDHeader, c.ID)
	req.Header.Add(jwt.Authorization, jwt.ToHeaderVal(c.token.Bearer))
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

// Ok calls the /ok endpoint
func (c *Client) Ok() (*http.Response, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.Host+app.OkRoute, nil)
	if err != nil {
		return nil, nil, err
	}
	return c.makeRequest(req)
}

// Status calls the /status endpoint
func (c *Client) Status() (*http.Response, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.Host+app.StatusRoute, nil)
	if err != nil {
		return nil, nil, err
	}
	return c.authedRequest(req)
}
