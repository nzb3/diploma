package gormpg

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/nzb3/diploma/resource-service/internal/repository/gormpg/dto"
)

func (r *Repository) migrate() error {
	const op = "postgres.Repository.migrate"

	if err := r.db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	m := []interface{}{
		&dto.Resource{},
	}

	// Migrate the basic tables
	if err := r.db.AutoMigrate(
		m...,
	); err != nil {
		slog.Error("Failed to run auto migrations", "op", op, "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
            CREATE INDEX IF NOT EXISTS idx_collections_name
            ON collections USING HASH (name)
        `).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
            CREATE INDEX IF NOT EXISTS idx_embeddings_collection
            ON embeddings (collection_id)
        `).Error; err != nil {
			return err
		}

		// Resource indexes
		if err := tx.Exec(`
            CREATE INDEX IF NOT EXISTS idx_resources_status
            ON resources USING HASH (status)
        `).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
            CREATE INDEX IF NOT EXISTS idx_resources_type
            ON resources USING HASH (type)
        `).Error; err != nil {
			return err
		}

		return nil
	})
}
