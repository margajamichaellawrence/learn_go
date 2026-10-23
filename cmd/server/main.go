package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"go_auth/api"
)

func main() {
	backendURL := flag.String("backend-url", "", "auth backend base URL, e.g. https://dev-backend-platform.eufit.io")
	flag.Parse()
	if *backendURL == "" {
		log.Fatal("-backend-url is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	log.Fatal(api.New(client, *backendURL).Listen(":3000"))
}
