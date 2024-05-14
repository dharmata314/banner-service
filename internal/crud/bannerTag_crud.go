package crud

import (
	errMsg "banner-serivce/internal/api/err"
	"banner-serivce/internal/structs"
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BannerTagRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewBannerTagRepository(db *pgxpool.Pool, log *slog.Logger) *BannerTagRepository {
	return &BannerTagRepository{db, log}
}

func (btr *BannerTagRepository) CreateBannerTag(ctx context.Context, bannerTag *structs.BannerTag) error {
	_, err := btr.db.Exec(ctx, `INSERT INTO banner_tags (banner_id, tag_id) VALUES ($1, $2)`, bannerTag.BannerID, bannerTag.TagID)
	if err != nil {
		btr.log.Error("Failed to create BannerTag", errMsg.Err(err))
		return err
	}

	return nil
}

func (btr *BannerRepository) FindBannerTagsByBannerID(ctx context.Context, bannerID int) ([]structs.BannerTag, error) {
	rows, err := btr.db.Query(ctx, `SELECT * FROM banner_tags WHERE banner_id = $1`, bannerID)
	if err != nil {
		btr.log.Error("Failed to find BannerTags by Banner ID", errMsg.Err(err))
		return nil, err
	}
	defer rows.Close()

	var bannerTagsArr []structs.BannerTag

	for rows.Next() {
		var bannerTag structs.BannerTag
		err := rows.Scan(&bannerTag.BannerID, &bannerTag.TagID)
		if err != nil {
			btr.log.Error("failed to scan bannerTags", errMsg.Err(err))
			return nil, err
		}
		bannerTagsArr = append(bannerTagsArr, bannerTag)
	}

	return bannerTagsArr, nil
}
