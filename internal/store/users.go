package store

import (
	"context"
	"database/sql"
)

type UsersStore struct {
	DB *sql.DB
}

func (u UsersStore) Create(ctx context.Context) error {
	return nil
}
