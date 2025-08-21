package store

import (
	"context"
	"database/sql"
)

type PostsStore struct {
	DB *sql.DB
}

func (p *PostsStore) Create(ctx context.Context) error {
	return nil
}
