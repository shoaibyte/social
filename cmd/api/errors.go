package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s, method: %s, path: %s", err, r.Method, r.URL.Path)
	_ = writeJSONError(w, http.StatusInternalServerError, err.Error())
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error: %s, method: %s, path: %s", err, r.Method, r.URL.Path)
	_ = writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) unprocessableEntityResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("unprocessable entity error: %s, method: %s, path: %s", err, r.Method, r.URL.Path)
	_ = writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
}

func (app *application) NotFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found error: %s, method: %s, path: %s", err, r.Method, r.URL.Path)
	_ = writeJSONError(w, http.StatusNotFound, err.Error())
}
