package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// client qt API.
type Client struct {
	BaseURL  string
	TenantID string
	APIKey   string
	Client   *http.Client
}

func NewClient(baseURL, tenantID, apiKey string) *Client {
	return &Client{
		BaseURL:  baseURL,
		TenantID: tenantID,
		APIKey:   apiKey,
		Client:   &http.Client{},
	}
}

func (c *Client) doReq(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", c.TenantID)
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API Error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) Health() (string, error) {
	resp, err := c.doReq("GET", "/health", nil)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *Client) GetProfile() (string, error) {
	resp, err := c.doReq("GET", "/api/v1/profile", nil)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *Client) GenerateTicket(count int, owner string) (string, error) {
	reqBody := map[string]interface{}{
		"count":    count,
		"owner_id": owner,
	}
	resp, err := c.doReq("POST", "/api/v1/tickets/generate", reqBody)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *Client) VerifyTicket(token string) (string, error) {
	reqBody := map[string]interface{}{
		"token": token,
	}
	resp, err := c.doReq("POST", "/api/v1/tickets/verify", reqBody)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *Client) RevokeTicket(ticketID, actorID, reason string) (string, error) {
	reqBody := map[string]interface{}{
		"ticket_id": ticketID,
		"actor_id":  actorID,
		"reason":    reason,
	}
	resp, err := c.doReq("POST", "/api/v1/tickets/revoke", reqBody)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *Client) GetMembers() (string, error) {
	resp, err := c.doReq("GET", "/api/v1/members", nil)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *Client) CreateMember(name, role string) (string, error) {
	reqBody := map[string]interface{}{
		"name": name,
		"role": role,
	}
	resp, err := c.doReq("POST", "/api/v1/members", reqBody)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// batches
func (c *Client) GetBatchStatus(batchID string) (string, error) {
	resp, err := c.doReq("GET", "/api/v1/batches/"+batchID+"/status", nil)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func prettyJSON(raw string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return raw
	}
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return raw
	}
	return string(b)
}
