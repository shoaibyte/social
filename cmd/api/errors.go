package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s, method: %s, path: %s", err, r.Method, r.URL.Path)
	_ = writeJSONError(w, http.StatusInternalServerError, "server encountered internal error")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error: %s, method: %s, path: %s", err, r.Method, r.URL.Path)
	_ = writeJSONError(w, http.StatusBadRequest, "client made a bad request")
}

func (app *application) unprocessableEntityResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("unprocessable entity error: %s, method: %s, path: %s", err, r.Method, r.URL.Path)
	_ = writeJSONError(w, http.StatusUnprocessableEntity, "cannot process the request entity")
}

func (app *application) NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found error: %s, method: %s, path: %s", err, r.Method, r.URL.Path)
	_ = writeJSONError(w, http.StatusNotFound, "resource not found")
}
