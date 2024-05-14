package crud

import (
	errMsg "banner-serivce/internal/api/err"
	"banner-serivce/internal/structs"
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewUserRepository(db *pgxpool.Pool, log *slog.Logger) *UserRepository {
	return &UserRepository{db, log}
}

func (u *UserRepository) CreateUser(ctx context.Context, user *structs.User) error {
	err := u.db.QueryRow(ctx, `INSERT INTO users (username, password, role) VALUES ($1, $2, $3) RETURNING id`, user.Username, user.Password, user.Role).Scan(&user.ID)
	if err != nil {
		u.log.Error("Failed to create user", errMsg.Err(err))
		return err
	}
	return nil
}

func (u *UserRepository) FindUserByName(ctx context.Context, username string) (structs.User, error) {
	query, err := u.db.Query(ctx, `SELECT * FROM users WHERE username = $1`, username)
	if err != nil {
		u.log.Error("Error querying users", errMsg.Err(err))
		return structs.User{}, err
	}
	row := structs.User{}
	defer query.Close()
	if !query.Next() {
		u.log.Error("User not found")
		return structs.User{}, fmt.Errorf("User not found")
	} else {
		err := query.Scan(&row.ID, &row.Username, &row.Password, &row.Role)
		if err != nil {
			u.log.Error("Error scanning users", errMsg.Err(err))
			return structs.User{}, err
		}
	}
	return row, nil
}

func (ur *UserRepository) FindUserById(ctx context.Context, id int) (structs.User, error) {
	query, err := ur.db.Query(ctx,
		`SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		ur.log.Error("Error querying users", errMsg.Err(err))
		return structs.User{}, err
	}
	defer query.Close()
	rowArray := structs.User{}
	if !query.Next() {
		ur.log.Error("User not found")
		return structs.User{}, fmt.Errorf("User not found")

	} else {
		err := query.Scan(&rowArray.ID, &rowArray.Username, &rowArray.Password, &rowArray.Role)
		if err != nil {
			ur.log.Error("Error scanning users", errMsg.Err(err))
			return structs.User{}, err
		}
	}
	return rowArray, nil
}
