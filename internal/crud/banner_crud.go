package crud

import (
	errMsg "banner-serivce/internal/api/err"
	bannerhandlers "banner-serivce/internal/handlers/banner_handlers"
	"banner-serivce/internal/structs"
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strconv"
)

type BannerRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewBannerRepository(db *pgxpool.Pool, log *slog.Logger) *BannerRepository {
	return &BannerRepository{db, log}
}

func (br *BannerRepository) CreateBanner(ctx context.Context, banner *structs.Banner) error {

	err := br.db.QueryRow(ctx,
		`INSERT INTO banners
		(
					feature_id,
					content,
					is_active,
					created_at,
					updated_at
		)
		VALUES
		(
					$1,
					$2,
					$3,
					$4,
					$5
		)
		returning id`,
		banner.FeatureID,
		banner.Content,
		banner.IsActive,
		banner.CreatedAt,
		banner.UpdatedAt,
	).Scan(&banner.ID)
	if err != nil {
		br.log.Error("failed to create banner", errMsg.Err(err))
		return err
	}
	return nil
}

func (br *BannerRepository) FindBannerByID(ctx context.Context, id int) (structs.Banner, error) {

	var banner structs.Banner

	row := br.db.QueryRow(ctx, `SELECT & FROM banners WHERE id = $1`, id)

	err := row.Scan(&banner.ID, &banner.FeatureID, &banner.Content, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return structs.Banner{}, err
		}
		br.log.Error("failed to find banner row", errMsg.Err(err))
	}

	return banner, nil

}

func (br *BannerRepository) FindBannerByFeatureID(ctx context.Context, feature_id int) ([]structs.Banner, error) {
	query, err := br.db.Query(ctx, `SELECT * FROM banners WHERE feature_id = $1`, feature_id)
	if err != nil {
		br.log.Error("Error querying banners", errMsg.Err(err))
		return nil, err
	}

	defer query.Close()

	var bannersArr []structs.Banner

	for query.Next() {
		var bannerRow structs.Banner

		err := query.Scan(&bannerRow.ID, &bannerRow.FeatureID, &bannerRow.Content, &bannerRow.IsActive, &bannerRow.CreatedAt, &bannerRow.UpdatedAt)
		if err != nil {
			br.log.Error("failed to scan banners", errMsg.Err(err))
			return nil, err
		}
		bannersArr = append(bannersArr, bannerRow)
	}

	if len(bannersArr) == 0 {
		br.log.Info("No banners were found for feature ID:", feature_id)
		return []structs.Banner{}, nil
	}

	return bannersArr, nil
}

func (br *BannerRepository) FindBannerByTagID(ctx context.Context, tag_id int) ([]structs.Banner, error) {

	query, err := br.db.Query(ctx, `SELECT * FROM banner_tags WHERE tag_id = $1`, tag_id)
	if err != nil {
		br.log.Error("Error querying banners", errMsg.Err(err))
		return nil, err
	}

	defer query.Close()

	var bannersArr []structs.Banner

	for query.Next() {
		var bannerRow structs.Banner

		err := query.Scan(&bannerRow.ID, &bannerRow.FeatureID, &bannerRow.Content, &bannerRow.IsActive, &bannerRow.CreatedAt, &bannerRow.UpdatedAt)
		if err != nil {
			br.log.Error("failed to scan banners", errMsg.Err(err))
			return nil, err
		}
		bannersArr = append(bannersArr, bannerRow)
	}

	if len(bannersArr) == 0 {
		br.log.Info("No banners were found for tag ID:", tag_id)
		return []structs.Banner{}, nil
	}

	return bannersArr, nil
}

func (br *BannerRepository) FindBannerByFeatureTag(ctx context.Context, featureID, tagID int) (*structs.Banner, error) {

	query := `SELECT b.id,
	b.feature_id,
	b.content,
	b.is_active,
	b.created_at,
	b.updated_at
FROM   banners b
	INNER JOIN banner_tags bt
			ON b.id = bt.banner_id
WHERE  b.feature_id = $1
	AND bt.tag_id = $2
	AND b.is_active = true`

	row := br.db.QueryRow(ctx, query, featureID, tagID)

	var banner structs.Banner

	err := row.Scan(&banner.ID, &banner.FeatureID, &banner.Content, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt)
	if err != nil {
		br.log.Error("Error with database", errMsg.Err(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		br.log.Error("Failed to find banner", errMsg.Err(err))
		return nil, err
	}

	return &banner, nil
}

func (br *BannerRepository) DeleteBannerByID(ctx context.Context, id int) error {
	_, err := br.db.Exec(ctx, `DELETE FROM banners WHERE id = $1`, id)
	if err != nil {
		br.log.Error("failed to delete banner", errMsg.Err(err))
		return err
	}
	return nil
}

func (br *BannerRepository) FindBannersByParameters(ctx context.Context, params bannerhandlers.RequestGetBanners) ([]structs.Banner, error) {
	query := "SELECT b.id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at, array_agg(bt.tag_id) AS tag_ids FROM banners b LEFT JOIN banner_tags bt ON b.id = bt.banner_id WHERE 1=1"
	args := []interface{}{}

	if params.FeatureID != nil {
		query += " AND b.feature_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, *params.FeatureID)
	}

	if params.TagID != nil {
		query += " AND b.id IN (SELECT banner_id FROM banner_tags WHERE tag_id = $" + strconv.Itoa(len(args)+1) + ")"
		args = append(args, *params.TagID)
	}

	query += " GROUP BY b.id"

	if params.Limit != nil {
		query += " LIMIT $" + strconv.Itoa(len(args)+1)
		args = append(args, *params.Limit)
	}

	if params.Offset != nil {
		query += " OFFSET $" + strconv.Itoa(len(args)+1)
		args = append(args, *params.Offset)
	}

	rows, err := br.db.Query(ctx, query, args...)
	if err != nil {
		br.log.Error("Failed to query banners", errMsg.Err(err))
		return nil, err
	}
	defer rows.Close()

	var banners []structs.Banner
	for rows.Next() {
		var banner structs.Banner
		var tagIDs []int
		if err := rows.Scan(&banner.ID, &banner.FeatureID, &banner.Content, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt, &tagIDs); err != nil {
			br.log.Error("Failed to scan banner row", errMsg.Err(err))
			return nil, err
		}
		banner.TagIDs = tagIDs
		banners = append(banners, banner)
	}

	if err := rows.Err(); err != nil {
		br.log.Error("Error occurred while iterating banner rows", errMsg.Err(err))
		return nil, err
	}

	return banners, nil
}

func (br *BannerRepository) UpdateBanner(ctx context.Context, banner *structs.Banner) error {

	tx, err := br.db.Begin(ctx)
	if err != nil {
		br.log.Error("Failed to begin transaction", errMsg.Err(err))
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM banner_tags WHERE banner_id = $1`, banner.ID)
	if err != nil {
		br.log.Error("failed to delete old tags for banner", errMsg.Err(err))
		return err
	}

	for _, tagID := range banner.TagIDs {
		_, err = tx.Exec(ctx, `INSERT INTO banner_tags (banner_id, tag_id) VALUES ($1, $2)`, banner.ID, tagID)
		if err != nil {
			br.log.Error("failed to insert tag for banner", errMsg.Err(err))
			return err
		}
	}

	_, err = tx.Exec(ctx,
		`UPDATE banners SET feature_id = $1, content = $2, is_active = $3, updated_at = $4 WHERE id = $5`,
		banner.FeatureID, banner.Content, banner.IsActive, banner.UpdatedAt, banner.ID)
	if err != nil {
		br.log.Error("Failed to update banner", errMsg.Err(err))
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		br.log.Error("Failed to commit transaction", errMsg.Err(err))
		return err
	}

	return nil
}
