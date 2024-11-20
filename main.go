package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	config, err := app.NewConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	app, err := app.New(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	if err = app.Setup(ctx, config.DB.DSN); err != nil {
		log.Fatal(err)
	}

	if err = app.Start(); err != nil {
		log.Fatal(err)
	}
}
