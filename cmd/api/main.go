package main

import (
	"log"
	"social/internal/env"
	"social/internal/store"
)

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
	}

	storage := store.NewStorage(nil)

	app := &application{
		config: cfg,
		store:  storage,
	}

	mux := app.mount()

	log.Fatal(app.run(mux))
}
