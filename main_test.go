package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newFakeAuthServer starts a test server that accepts only demo/secret.
func newFakeAuthServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if req.Username != "demo" || req.Password != "secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(LoginResponse{Token: "test-token"})
	}))
	t.Cleanup(server.Close)
	return server
}

func TestLoginPass(t *testing.T) {
	server := newFakeAuthServer(t)

	token, err := Login(server.Client(), server.URL, "demo", "secret")
	if err != nil {
		t.Fatalf("expected login to pass, got error: %v", err)
	}
	if token != "test-token" {
		t.Errorf("token = %q, want test-token", token)
	}
}

func TestLoginFail(t *testing.T) {
	server := newFakeAuthServer(t)

	token, err := Login(server.Client(), server.URL, "demo", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
	if token != "" {
		t.Errorf("expected no token, got %q", token)
	}
}
