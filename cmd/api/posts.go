package main

import (
	"net/http"
	"social/internal/store"
)

type CreatePostPayload struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		if err := writeJSONError(w, http.StatusUnprocessableEntity, err.Error()); err != nil {
			return
		}
		return
	}

	userID := 1

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		UserID:  int64(userID),
		Tags:    payload.Tags,
	}

	ctx := r.Context()

	if err := app.store.Posts.Create(ctx, post); err != nil {
		if err := writeJSONError(w, http.StatusInternalServerError, err.Error()); err != nil {
			return
		}
		return
	}

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		if err := writeJSONError(w, http.StatusInternalServerError, err.Error()); err != nil {
			return
		}
		return
	}
}
