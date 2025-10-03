package main

import (
	"log"
	"net/http"
	"time"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}

	if err := writeJSON(w, http.StatusOK, data); err != nil {
		log.Printf("health check: failed to write JSON response: %v", err)
		return
	}

	return
}
