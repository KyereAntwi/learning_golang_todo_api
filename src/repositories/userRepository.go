package repositories

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"todoapi.com/m/src/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(username, password, email, primaryPhone string) (uuid.UUID, error) {
	user, err := domain.NewUser(username, password, email, primaryPhone)

	if err != nil {
		return uuid.Nil, err
	}

	query := `
	INSERT INTO users (id, username, hashed_password, email, primary_phone, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	args := []interface{}{user.GetID(), user.GetUsername(), user.GetHashedPassword(), user.GetEmail(), user.GetPrimaryPhone(), user.GetCreatedAt(), user.GetUpdatedAt()}

	_, err = r.db.Exec(query, args...)

	if err != nil {
		return uuid.Nil, err
	}

	return user.GetID(), nil
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (domain.User, error) {
	query := `
	SELECT id, username, hashed_password, email, primary_phone, created_at, updated_at
	FROM users
	WHERE id = $1`

	row := r.db.QueryRow(query, id)
	var username, email, primaryPhone, hashedPassword string
	var createdAt, updatedAt time.Time

	err := row.Scan(&id, &username, &hashedPassword, &email, &primaryPhone, &createdAt, &updatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return domain.User{}, errors.New("user not found")
		default:
			return domain.User{}, err
		}
	}

	user := domain.ReconstructUser(id, username, hashedPassword, email, primaryPhone, createdAt, updatedAt)

	return user, nil
}

func (r *UserRepository) GetUserByUsername(username string) (domain.User, error) {
	query := `
	SELECT id, username, hashed_password, email, primary_phone, created_at, updated_at
	FROM users
	WHERE username = $1`

	row := r.db.QueryRow(query, username)
	var id uuid.UUID
	var email, primaryPhone, hashedPassword string
	var createdAt, updatedAt time.Time

	err := row.Scan(&id, &username, &hashedPassword, &email, &primaryPhone, &createdAt, &updatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return domain.User{}, errors.New("user not found")
		default:
			return domain.User{}, err
		}
	}

	user := domain.ReconstructUser(id, username, hashedPassword, email, primaryPhone, createdAt, updatedAt)

	return user, nil
}

func (r *UserRepository) IsUsernameTaken(username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = $1`

	var count int
	err := r.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) Authenticate(username, password string) (string, error) {
	user, err := r.GetUserByUsername(username)
	if err != nil {
		return "", err
	}

	if !user.CheckPassword(password) {
		return "", errors.New("invalid credentials")
	}

	return user.GetID().String(), nil
}

func (r *UserRepository) StoreRefreshToken(token domain.RefreshToken) error {
	query := `
	INSERT INTO refresh_tokens (user_id, hashed_token, expires_at, created_at) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id) DO UPDATE 
	SET hashed_token = EXCLUDED.hashed_token, expires_at = EXCLUDED.expires_at, created_at = EXCLUDED.created_at`

	args := []interface{}{token.UserId, token.HashedToken, token.ExpiresAt, token.CreatedAt}

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *UserRepository) GetRefreshTokenByUserID(userID uuid.UUID) (domain.RefreshToken, error) {
	query := `
	SELECT user_id, hashed_token, expires_at, created_at
	FROM refresh_tokens
	WHERE user_id = $1`

	row := r.db.QueryRow(query, userID)
	var hashedToken string
	var expiresAt, createdAt time.Time

	err := row.Scan(&userID, &hashedToken, &expiresAt, &createdAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return domain.RefreshToken{}, errors.New("refresh token not found")
		default:
			return domain.RefreshToken{}, err
		}
	}

	refreshToken, err := base64.StdEncoding.DecodeString(hashedToken)
	if err != nil {
		return domain.RefreshToken{}, err
	}

	return domain.RefreshToken{
		UserId:      userID,
		HashedToken: string(refreshToken),
		ExpiresAt:   expiresAt,
		CreatedAt:   createdAt,
	}, nil
}

func (r *UserRepository) DeleteRefreshToken(userID uuid.UUID) error {
	query := `
	DELETE FROM refresh_tokens
	WHERE user_id = $1`

	_, err := r.db.Exec(query, userID)
	return err
}
