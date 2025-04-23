package gormpg

import (
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) (*Repository, error) {
	const op = "postgres.NewRepository"

	r := &Repository{
		db: db,
	}

	err := r.migrate()

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Repository{
		db: db,
	}, nil
}
