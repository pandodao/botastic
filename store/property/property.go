package property

import (
	"botastic/core"
	"context"
	_ "embed"

	"github.com/jmoiron/sqlx"
)

func New(db *sqlx.DB) core.PropertyStore {
	return &store{
		db: db,
	}
}

type store struct {
	db *sqlx.DB
}

func (s *store) Get(ctx context.Context, key string) (core.PropertyValue, error) {
	query, args, err := s.db.BindNamed("", map[string]interface{}{
		"key": key,
	})

	if err != nil {
		return "", err
	}

	rows, err := s.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return "", err
	}

	pp := &core.Property{}

	return pp.Value, nil
}

func (s *store) Set(ctx context.Context, key string, value interface{}) error {
	query, args, err := s.db.BindNamed("", map[string]interface{}{
		"key":   key,
		"value": value,
	})

	if err != nil {
		return err
	}

	if _, err = s.db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}
