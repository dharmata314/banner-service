package crud

import (
	errMsg "banner-serivce/internal/api/err"
	"banner-serivce/internal/structs"
	"fmt"
	"log/slog"

	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TagRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewTagRepository(db *pgxpool.Pool, log *slog.Logger) *TagRepository {
	return &TagRepository{db, log}
}

func (tr *TagRepository) CreateTag(ctx context.Context, tag *structs.Tag) error {
	err := tr.db.QueryRow(ctx,
		`INSERT INTO tags (name)
		VALUES ($1)
		RETURNING id`, tag.Name).Scan(&tag.ID)
	if err != nil {
		tr.log.Error("Failed to create tag", errMsg.Err(err))
		return err
	}
	return nil
}

func (tr *TagRepository) FindTagById(ctx context.Context, id int) (structs.Tag, error) {
	var tag structs.Tag
	err := tr.db.QueryRow(ctx, `SELECT id, name FROM tags WHERE id = $1`, id).Scan(&tag.ID, &tag.Name)
	if err != nil {
		tr.log.Error("Failed to find Tag by ID", errMsg.Err(err))
		return structs.Tag{}, err
	}
	return tag, nil
}

func (tr *TagRepository) FindTagByName(ctx context.Context, name string) (structs.Tag, error) {
	query, err := tr.db.Query(ctx,
		`SELECT * FROM tags WHERE name = $1`, name)
	if err != nil {
		tr.log.Error("Tag not found", errMsg.Err(err))
		return structs.Tag{}, err
	}
	row := structs.Tag{}
	if !query.Next() {
		tr.log.Error("Tag not found")
		return structs.Tag{}, err
	} else {
		err := query.Scan(&row.ID, &row.Name)
		if err != nil {
			tr.log.Error("Tag not found", errMsg.Err(err))
			return structs.Tag{}, fmt.Errorf("Tag not found")
		}
	}
	return row, nil
}