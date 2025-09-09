package main

import (
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
		if err := writeJSONError(w, http.StatusInternalServerError, err.Error()); err != nil {
			return
		}
		return
	}
}
