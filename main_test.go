package main

import (
	"bufio"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// These tests hit the real dev API at loginURL.
// Put EUFIT_EMAIL and EUFIT_PASSWORD in .env (git-ignored, see .env.example)
// or set them in your shell to run TestLoginPass.

func TestMain(m *testing.M) {
	loadDotEnv(".env")
	os.Exit(m.Run())
}

// loadDotEnv reads KEY=VALUE lines from path into the environment.
// A missing file is ignored, and variables already set in the shell win.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}

func newClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

func TestLoginPass(t *testing.T) {
	email, password := os.Getenv("EUFIT_EMAIL"), os.Getenv("EUFIT_PASSWORD")
	if email == "" || password == "" {
		t.Skip("EUFIT_EMAIL and EUFIT_PASSWORD not set")
	}

	result, err := Login(newClient(), loginURL, email, password)
	if err != nil {
		t.Fatalf("expected login to pass, got error: %v", err)
	}
	if result.Tokens.AccessToken == "" {
		t.Error("expected an access token")
	}
	if result.Tokens.RefreshToken == "" {
		t.Error("expected a refresh token")
	}
	if result.User.Email != email {
		t.Errorf("user email = %q, want %q", result.User.Email, email)
	}
}

func TestLoginFail(t *testing.T) {
	// Uses an address that shouldn't exist so a real account is never hit with a bad password.
	result, err := Login(newClient(), loginURL, "no-such-user@example.invalid", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
	if result != nil {
		t.Errorf("expected no result, got %+v", result)
	}
}
