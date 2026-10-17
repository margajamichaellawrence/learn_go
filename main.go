package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

var ErrInvalidCredentials = errors.New("invalid username or password")

// Login sends username and password to url and returns the token from the server.
func Login(client *http.Client, url, username, password string) (string, error) {
	body, err := json.Marshal(LoginRequest{Username: username, Password: password})
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrInvalidCredentials
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var out LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return out.Token, nil
}

func main() {
	client := &http.Client{Timeout: 10 * time.Second}

	token, err := Login(client, "http://localhost:8080/login", "demo", "secret")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("logged in, token:", token)
}
