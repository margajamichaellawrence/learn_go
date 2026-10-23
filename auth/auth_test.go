package auth

import (
	"errors"
	"flag"
	"net/http"
	"testing"
	"time"
)

// These tests hit the real API. Pass the URL and test account on the command line:
//
//	go test -v ./... -url=https://dev-backend-platform.eufit.io -email=you@example.com -password=yourpass
var (
	baseURL  = flag.String("url", "", "API base URL, e.g. https://dev-backend-platform.eufit.io")
	email    = flag.String("email", "", "test account email")
	password = flag.String("password", "", "test account password")
)

func newClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

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
			email:     *email,
			password:  *password,
			wantEmail: *email,
		},
		{
			// Fake account on purpose, so a real account is never hit with a bad password.
			name:     "invalid credentials",
			email:    "no-such-user@example.invalid",
			password: "wrong-password",
			wantErr:  ErrInvalidCredentials,
		},
	}

	if *baseURL == "" || *email == "" || *password == "" {
		t.Skip("pass -url, -email and -password to run")
	}
	client := newClient()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Login(client, *baseURL, tt.email, tt.password)
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
