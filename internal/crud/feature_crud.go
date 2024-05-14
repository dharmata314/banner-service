package crud

import (
	errMsg "banner-serivce/internal/api/err"
	"banner-serivce/internal/structs"
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FeatureRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewFeatureRepository(db *pgxpool.Pool, log *slog.Logger) *FeatureRepository {
	return &FeatureRepository{db, log}
}

func (fr *FeatureRepository) CreateFeature(ctx context.Context, feature *structs.Feature) error {
	err := fr.db.QueryRow(ctx, `INSERT INTO features (name) VALUES ($1) RETURNING id`, feature.Name).Scan(&feature.ID)
	if err != nil {
		fr.log.Error("failed creating feature", errMsg.Err(err))
		return err
	}
	return nil
}

func (fr *FeatureRepository) FindFeatureById(ctx context.Context, id int) (structs.Feature, error) {
	var feature structs.Feature

	row := fr.db.QueryRow(ctx, `SELECT id, name FROM features WHERE id = $1`, id)

	err := row.Scan(&feature.ID, feature.Name)

	if err != nil {
		fr.log.Error("Failed to find Feature by ID", errMsg.Err(err))
		return structs.Feature{}, err
	}

	return feature, nil
}

func (fr *FeatureRepository) FindFeatureByName(ctx context.Context, name string) (structs.Feature, error) {
	query, err := fr.db.Query(ctx, `SELECT * FROM features WHERE name = $1`, name)
	if err != nil {
		fr.log.Error("Feature not found", errMsg.Err(err))
		return structs.Feature{}, err
	}
	row := structs.Feature{}
	if !query.Next() {
		fr.log.Error("Feature not found")
		return structs.Feature{}, fmt.Errorf("Feature not found")
	} else {
		err := query.Scan(&row.ID, &row.Name)
		if err != nil {
			fr.log.Error("Feature not found", errMsg.Err(err))
		}
	}
	return row, nil

}
