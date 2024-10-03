package main

import (
	"context"
	"log"
	"telbot/internal/api"
)

func main() {
	app := api.New(context.Background())

	err := app.Run()
	if err != nil {
		log.Println(err)
	}
}
