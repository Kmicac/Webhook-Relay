package main

import (
	"github.com/Kmicac/webhookrelay/internal/api"
)

func main() {
	e := api.NewServer()

	e.Logger.Fatal(e.Start(":8080"))
}
