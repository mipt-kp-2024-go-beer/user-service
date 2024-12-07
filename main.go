package main

import (
	"context"
	"log"

	"github.com/mipt-kp-2024-go-beer/user-service/internal/app"
)

func main() {
	ctx := context.Background()

	config, err := app.NewConfig("configs/config.yml")
	if err != nil {
		log.Fatal(err)
	}

	app, err := app.New(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	if err = app.Setup(ctx, config.DB); err != nil {
		log.Fatal(err)
	}

	if err = app.Start(); err != nil {
		log.Fatal(err)
	}
}
