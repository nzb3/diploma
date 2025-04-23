package gormpg

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/nzb3/diploma/search/internal/domain/models"
)

func (r *Repository) migrate() error {
	const op = "postgres.Repository.migrate"

	if err := r.db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := r.db.Exec("CREATE EXTENSION IF NOT EXISTS \"vector\"").Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	m := []interface{}{
		&models.Resource{},
		&models.ResourceEmbedding{},
		&models.Collection{},
		&models.Embedding{},
	}

	// Migrate the basic tables
	if err := r.db.AutoMigrate(
		m...,
	); err != nil {
		slog.Error("Failed to run auto migrations", "op", op, "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	// Add owner_id column if it doesn't exist
	// This is a safer approach than relying on AutoMigrate for existing production tables
	if !r.db.Migrator().HasColumn(&models.Resource{}, "owner_id") {
		slog.Info("Adding owner_id column to resources table")
		if err := r.db.Exec("ALTER TABLE resources ADD COLUMN IF NOT EXISTS owner_id VARCHAR(100)").Error; err != nil {
			slog.Error("Failed to add owner_id column", "op", op, "error", err)
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	// Create indexes and constraints
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

		if err := tx.Exec(`
			ALTER TABLE collections ADD CONSTRAINT uni_collections_name UNIQUE (name);
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM pg_constraint 
					WHERE conname = 'fk_resource' 
					AND conrelid = 'resource_embedding'::regclass
				) THEN
					ALTER TABLE resource_embedding 
					ADD CONSTRAINT fk_resource 
					FOREIGN KEY (resource_id) 
					REFERENCES resources(id) 
					ON DELETE CASCADE;
				END IF;
			END $$;
		`).Error; err != nil {
			return err
		}

		// For fk_embedding constraint
		if err := tx.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM pg_constraint 
					WHERE conname = 'fk_resource' 
					AND conrelid = 'resource_embedding'::regclass
				) THEN
					ALTER TABLE resource_embedding 
					ADD CONSTRAINT fk_embedding 
					FOREIGN KEY (embedding_id) 
					REFERENCES embeddings(uuid) 
					ON DELETE CASCADE;
				END IF;
			END $$;
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			CREATE OR REPLACE FUNCTION delete_orphaned_embeddings() RETURNS TRIGGER AS $$
			BEGIN
				DELETE FROM embeddings 
				WHERE uuid = OLD.embedding_id 
				AND NOT EXISTS (
					SELECT 1 FROM resource_embedding 
					WHERE embedding_id = OLD.embedding_id
				);
				RETURN OLD;
			END;
			$$ LANGUAGE plpgsql;
		
	
			CREATE OR REPLACE TRIGGER  trg_delete_orphaned_embeddings
			AFTER DELETE ON resource_embedding
			FOR EACH ROW
			EXECUTE FUNCTION delete_orphaned_embeddings();
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
            CREATE UNIQUE INDEX IF NOT EXISTS idx_resource_embedding_unique 
            ON resource_embedding (resource_id, embedding_id)
        `).Error; err != nil {
			return err
		}

		return nil
	})
}
