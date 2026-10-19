package main

import (
	"log"

	"go_auth/api"
)

func main() {
	log.Fatal(api.New().Listen(":3000"))
}
