package postgresql

import (
	"banner-serivce/internal/auth"
	"banner-serivce/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"sync"
)

type Postgres struct {
	Db     *pgxpool.Pool
	log    *slog.Logger
	Config *config.Config
}

var (
	pgInstance *Postgres
	pgOnce     sync.Once
)

func NewPG(ctx context.Context, connString string, log *slog.Logger, cfg *config.Config) (*Postgres, error) {
	var err error

	pgOnce.Do(func() {
		var db *pgxpool.Pool
		db, err = pgxpool.New(ctx, connString)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}

		pgInstance = &Postgres{db, log, cfg}
		if err = CreateTables(ctx, db, log, cfg); err != nil {
			return
		}
	})

	if err != nil {
		return nil, err
	}
	return pgInstance, nil
}

func CreateTables(ctx context.Context, db *pgxpool.Pool, log *slog.Logger, cfg *config.Config) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS banners (
			id SERIAL PRIMARY KEY,
			feature_id INTEGER,
			content JSONB,
			is_active BOOLEAN,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create banners table: %w", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tags (
			id SERIAL PRIMARY KEY,
			name TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tags table: %w", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS banner_tags (
			banner_id INTEGER,
			tag_id INTEGER,
			PRIMARY KEY (banner_id, tag_id),
			FOREIGN KEY (banner_id) REFERENCES banners(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create banner_tags table: %w", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS features (
			id SERIAL PRIMARY KEY,
			name TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create features table: %w", err)
	}

	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
		    id SERIAL PRIMARY KEY ,
			username TEXT UNIQUE,
			password TEXT,
			role TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	hashPass, err := auth.HashPassword(cfg.DefaultAdminPass)
	_, err = db.Exec(ctx, `INSERT INTO users (username, password, role) VALUES ($1,$2,$3)`, "admin", hashPass, "admin")
	log.Info("Tables created (or updated)")
	return nil
}

func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.Db.Ping(ctx)
}

func (pg *Postgres) Close() {
	pg.Db.Close()
}
