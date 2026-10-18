// Package auth is a client for the eufit auth API. Its tests are the tester app:
// they run against the real API and fail if the login contract breaks.
package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const loginPath = "/auth/login"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenPair holds the short-lived access JWT and the longer-lived refresh JWT.
// Only the access token is accepted as a bearer.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

type LoginResult struct {
	Tokens TokenPair `json:"tokens"`
	User   UserInfo  `json:"user"`
}

var (
	// ErrInvalidCredentials is returned on 401 (AUTH_004).
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrForbidden is returned on 403: unverified email (AUTH_006),
	// disabled account (AUTHZ_002) or suspended organization (AUTHZ_003).
	ErrForbidden = errors.New("login forbidden")
)

// Login sends email and password to url and returns the tokens and user info from the server.
func Login(client *http.Client, url, email, password string) (*LoginResult, error) {
	body, err := json.Marshal(LoginRequest{Email: email, Password: password})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, ErrInvalidCredentials
	case http.StatusForbidden:
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", ErrForbidden, msg)
	default:
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %s: %s", resp.Status, msg)
	}

	var out LoginResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}
