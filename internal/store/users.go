package store

import (
	"context"
	"database/sql"
)

type User struct {
	ID int64 `json:"id"`
}

type UserStore struct {
	db *sql.DB
}

func (s UserStore) Create(ctx context.Context, user *User) error {
	return nil
}
