package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/romeulima/devbook-server/internal/models"
)

type PgStore struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) PgStore {
	return PgStore{
		db: db,
	}
}

func (ps *PgStore) CreateUser(ctx context.Context, u models.UserPayload) (*models.UserResponse, error) {
	row := ps.db.QueryRow(ctx,
		`INSERT INTO users (name, nick, email, password) values ($1, $2, $3, $4) RETURNING id, created_at`,
		u.Name, u.Nick, u.Email, u.Password)

	var (
		id        uuid.UUID
		createdAt time.Time
	)

	if err := row.Scan(&id, &createdAt); err != nil {
		return nil, err
	}

	userResponse := models.NewUserResponse(id, u, createdAt)

	return userResponse, nil
}

func (ps *PgStore) GetUsers(ctx context.Context, param string) ([]models.UserResponse, error) {
	q := fmt.Sprintf(`SELECT id, name, nick, email, created_at FROM users WHERE name ILIKE '%%%s%%' OR nick ILIKE '%%%s%%'`, param, param)

	rows, err := ps.db.Query(ctx, q)

	if err != nil {
		return nil, err
	}

	var users []models.UserResponse

	for rows.Next() {
		var u models.UserResponse

		if err = rows.Scan(&u.ID, &u.Name, &u.Nick, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (ps *PgStore) GetUserByID(ctx context.Context, id uuid.UUID) (models.UserResponse, error) {
	row := ps.db.QueryRow(ctx,
		`SELECT id, name, nick, email, created_at FROM users WHERE id = $1`,
		id)

	var user models.UserResponse

	err := row.Scan(&user.ID, &user.Name, &user.Nick, &user.Email, &user.CreatedAt)

	return user, err
}

func (ps *PgStore) UpdateUser(ctx context.Context, id uuid.UUID, payload models.UserPayload) error {
	result, err := ps.db.Exec(ctx,
		`UPDATE users SET 
			name = COALESCE(NULLIF($1, ''), name),
			nick = COALESCE(NULLIF($2, ''), nick),
			email = COALESCE(NULLIF($3, ''), email)
			WHERE id = $4`,
		payload.Name, payload.Nick, payload.Email, id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (ps *PgStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	result, err := ps.db.Exec(ctx,
		`DELETE FROM users WHERE id = $1`, id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (ps *PgStore) GetUserByEmail(ctx context.Context, email string) (uuid.UUID, string, error) {
	row := ps.db.QueryRow(ctx,
		`SELECT id, password FROM users WHERE email = $1`, email)

	var (
		id       uuid.UUID
		password string
	)

	if err := row.Scan(&id, &password); err != nil {
		return uuid.Nil, "", err
	}

	return id, password, nil
}
