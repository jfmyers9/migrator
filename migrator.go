package migrator

import (
	"database/sql"
	"fmt"
)

type Migrator struct {
	db         *sql.DB
	migrations []Migration
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Migration

type Migration interface {
	Name() string
	Version() int
	Up(db *sql.Tx) error
	Down(db *sql.Tx) error
}

func NewMigrator(db *sql.DB, migrations ...Migration) *Migrator {
	return &Migrator{
		db:         db,
		migrations: migrations,
	}
}

func (m *Migrator) Setup() error {
	rows, err := m.db.Query("SELECT * FROM schema_migrations")
	if err != nil {
		return m.createSchemaMigrationsTable()
	}
	defer rows.Close()

	return nil
}

func (m *Migrator) createSchemaMigrationsTable() error {
	_, err := m.db.Exec(`CREATE TABLE schema_migrations (
		name VARCHAR (50) NOT NULL,
		version INTEGER UNIQUE NOT NULL
	)`)

	return err
}

func (m *Migrator) Migrate() error {
	currentVersion, err := m.getCurrentMigrationVersion()
	if err != nil {
		return err
	}

	for _, migration := range m.migrations {
		if migration.Version() <= currentVersion {
			continue
		}

		if err := m.runMigration(migration); err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) getCurrentMigrationVersion() (int, error) {
	rows, err := m.db.Query("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, nil
	}

	var version int
	err = rows.Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

func (m *Migrator) runMigration(migration Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := migration.Up(tx); err != nil {
		return err
	}

	result, err := tx.Exec("INSERT INTO schema_migrations VALUES ($1, $2)", migration.Name(), migration.Version())
	if err != nil {
		return err
	}

	if count, err := result.RowsAffected(); count != 1 || err != nil {
		return fmt.Errorf("invalid rows affected when inserting migration: expected 1, got %d", count)
	}

	return tx.Commit()
}

// TODO: Implement rollback
func (m *Migrator) Rollback() error {
	return nil
}
