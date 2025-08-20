package main

import "net/http"

func (app *application) healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Alive..."))
}
