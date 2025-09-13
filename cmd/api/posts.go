package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"social/internal/store"
	"strconv"
	"strings"
)

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.unprocessableEntityResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
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
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseInt(strings.TrimSpace(chi.URLParam(r, "postID")), 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	post, err := app.store.Posts.Get(ctx, postID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.NotFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	comments, err := app.store.Comments.GetByPostID(ctx, postID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
