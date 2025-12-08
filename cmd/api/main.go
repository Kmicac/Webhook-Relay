package main

import (
	"github.com/Kmicac/Webhook-Relay/internal/api"
)

func main() {
	e := api.NewServer()

	e.Logger.Fatal(e.Start(":8080"))
}
