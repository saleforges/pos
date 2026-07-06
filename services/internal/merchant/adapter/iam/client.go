package iam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/saleforge/pos/services/internal/merchant/port"
)

type tokenValidator struct {
	baseURL    string
	httpClient *http.Client
}

func NewTokenValidator(baseURL string) port.TokenValidator {
	return &tokenValidator{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type introspectRequest struct {
	Token string `json:"token"`
}

type introspectResponse struct {
	Active      bool     `json:"active"`
	UserID      string   `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

func (v *tokenValidator) Validate(tokenString string) (*port.TokenClaims, error) {
	body := introspectRequest{Token: tokenString}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := v.httpClient.Post(v.baseURL+"/api/v1/auth/introspect", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to introspect token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token validation failed with status %d", resp.StatusCode)
	}

	var result introspectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode introspect response: %w", err)
	}

	if !result.Active {
		return nil, fmt.Errorf("token is not active")
	}

	return &port.TokenClaims{
		UserID:      result.UserID,
		Roles:       result.Roles,
		Permissions: result.Permissions,
	}, nil
}
