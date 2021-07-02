// Package client is an app client library
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/grokloc/grokloc-go/pkg/app"
	"github.com/grokloc/grokloc-go/pkg/jwt"
	"github.com/grokloc/grokloc-go/pkg/models"
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

// org related

// CreateOrg creates an org
func (c *Client) CreateOrg(name string) (*http.Response, []byte, error) {
	bs, err := json.Marshal(app.CreateOrgMsg{Name: name})
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.Host+app.OrgRoute, bytes.NewBuffer(bs))
	if err != nil {
		return nil, nil, err
	}
	return c.authedRequest(req)
}

// ReadOrg reads an org
func (c *Client) ReadOrg(id string) (*http.Response, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.Host+app.OrgRoute+"/"+id, nil)
	if err != nil {
		return nil, nil, err
	}
	return c.authedRequest(req)
}

// UpdateOrgOwner updates an org owner
func (c *Client) UpdateOrgOwner(id, owner string) (*http.Response, []byte, error) {
	bs, err := json.Marshal(app.UpdateOrgOwnerMsg{Owner: owner})
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest(http.MethodPut, c.Host+app.OrgRoute+"/"+id, bytes.NewBuffer(bs))
	if err != nil {
		return nil, nil, err
	}
	return c.authedRequest(req)
}

// UpdateOrgStatus updates an org status
func (c *Client) UpdateOrgStatus(id string, status models.Status) (*http.Response, []byte, error) {
	bs, err := json.Marshal(app.UpdateStatusMsg{Status: status})
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest(http.MethodPut, c.Host+app.OrgRoute+"/"+id, bytes.NewBuffer(bs))
	if err != nil {
		return nil, nil, err
	}
	return c.authedRequest(req)
}

// user related

// CreateUser creates a user
func (c *Client) CreateUser(displayName, email, org, password string) (*http.Response, []byte, error) {
	bs, err := json.Marshal(app.CreateUserMsg{
		DisplayName: displayName,
		Email:       email,
		Org:         org,
		Password:    password,
	})
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.Host+app.UserRoute, bytes.NewBuffer(bs))
	if err != nil {
		return nil, nil, err
	}
	return c.authedRequest(req)
}

// ReadUser reads a user
func (c *Client) ReadUser(id string) (*http.Response, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.Host+app.UserRoute+"/"+id, nil)
	if err != nil {
		return nil, nil, err
	}
	return c.authedRequest(req)
}

// UpdateUserDisplayName updates a user display name
func (c *Client) UpdateUserDisplayName(id, displayName string) (*http.Response, []byte, error) {
	bs, err := json.Marshal(app.UpdateUserDisplayNameMsg{DisplayName: displayName})
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest(http.MethodPut, c.Host+app.UserRoute+"/"+id, bytes.NewBuffer(bs))
	if err != nil {
		return nil, nil, err
	}
	return c.authedRequest(req)
}

// UpdateUserPassword updates a user password
func (c *Client) UpdateUserPassword(id, password string) (*http.Response, []byte, error) {
	bs, err := json.Marshal(app.UpdateUserPasswordMsg{Password: password})
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest(http.MethodPut, c.Host+app.UserRoute+"/"+id, bytes.NewBuffer(bs))
	if err != nil {
		return nil, nil, err
	}
	return c.authedRequest(req)
}

// UpdateUserStatus updates a user status
func (c *Client) UpdateUserStatus(id string, status models.Status) (*http.Response, []byte, error) {
	bs, err := json.Marshal(app.UpdateStatusMsg{Status: status})
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest(http.MethodPut, c.Host+app.UserRoute+"/"+id, bytes.NewBuffer(bs))
	if err != nil {
		return nil, nil, err
	}
	return c.authedRequest(req)
}
