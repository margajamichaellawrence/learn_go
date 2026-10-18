package auth

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// These tests hit the real API at EUFIT_BASE_URL.
// Put EUFIT_BASE_URL, EUFIT_EMAIL and EUFIT_PASSWORD in .env (git-ignored, see .env.example)
// or set them in your shell.

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

// loginURL returns the login endpoint built from EUFIT_BASE_URL, skipping the test if it isn't set.
func loginURL(t *testing.T) string {
	t.Helper()
	baseURL := strings.TrimSuffix(os.Getenv("EUFIT_BASE_URL"), "/")
	if baseURL == "" {
		t.Skip("EUFIT_BASE_URL not set")
	}
	return baseURL + loginPath
}

func newClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

// TestLogin
// - run all tests EUFIT_BASE_URL=https://dev-backend-platform.eufit.io EUFIT_EMAIL=dembygenesis@gmail.com EUFIT_PASSWORD=<REDACTED> go test ./...
// - run this test EUFIT_BASE_URL=https://dev-backend-platform.eufit.io EUFIT_EMAIL=dembygenesis@gmail.com EUFIT_PASSWORD=<REDACTED> go test -run '^TestLogin$' -
func TestLogin(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		wantErr   error
		wantEmail string
	}{
		{
			name:      "success",
			email:     os.Getenv("EUFIT_EMAIL"),
			password:  os.Getenv("EUFIT_PASSWORD"),
			wantEmail: os.Getenv("EUFIT_EMAIL"),
		},
		{
			name:     "invalid credentials",
			email:    "no-such-user@example.invalid",
			password: "wrong-password",
			wantErr:  ErrInvalidCredentials,
		},
	}

	url := loginURL(t)
	if os.Getenv("EUFIT_EMAIL") == "" || os.Getenv("EUFIT_PASSWORD") == "" {
		t.Skip("EUFIT_EMAIL and EUFIT_PASSWORD not set")
	}
	client := newClient()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Login(client, url, tt.email, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Login() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if got != nil {
					t.Errorf("Login() = %+v, want nil", got)
				}
				return
			}

			if got.Tokens.AccessToken == "" {
				t.Error("Login() access token is empty")
			}
			if got.Tokens.RefreshToken == "" {
				t.Error("Login() refresh token is empty")
			}
			if got.User.Email != tt.wantEmail {
				t.Errorf("Login() user email = %q, want %q", got.User.Email, tt.wantEmail)
			}
		})
	}
}

func TestLoginPass(t *testing.T) {
	url := loginURL(t)
	email, password := os.Getenv("EUFIT_EMAIL"), os.Getenv("EUFIT_PASSWORD")
	if email == "" || password == "" {
		t.Skip("EUFIT_EMAIL and EUFIT_PASSWORD not set")
	}
	// doeson't reach it
	fmt.Println("url", url)

	result, err := Login(newClient(), url, email, password)
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
	result, err := Login(newClient(), loginURL(t), "no-such-user@example.invalid", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
	if result != nil {
		t.Errorf("expected no result, got %+v", result)
	}
}
