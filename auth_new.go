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

func Login(client *http.Client, url, email, password string) (*LoginResult, error) {
	body, err := json.Marshal(LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return nil, ErrInvalidCredentials
	case http.StatusForbidden:
		msg, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("%w: read body: %w", ErrForbidden, err)
		}
		return nil, fmt.Errorf("%w: %s", ErrForbidden, msg)
	default:
		msg, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("unexpected status %s: read body: %w", response.Status, err)
		}
		return nil, fmt.Errorf("unexpected status %s: %s", response.Status, msg)
	}

	var out LoginResult
	// you can reuse error variables because:
	//	- ":=" declares a new one

	// don't do this
	// if err = json.NewDecoder(response.Body).Decode(&out); err != nil {

	// do this, "=" only.. because error already declared above
	if err = json.NewDecoder(response.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}
