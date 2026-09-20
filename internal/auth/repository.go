package auth

import (
	"context"
	"time"

	"go-mini-setup/internal/database/db"
)

// AuthRepository defines database queries needed for authentication and tokens.
type AuthRepository interface {
	FindByEmail(ctx context.Context, email string) (*db.UserModel, error)
	FindByID(ctx context.Context, id string) (*db.UserModel, error)
	CreateUser(ctx context.Context, email, password string, name, role *string) (*db.UserModel, error)
	CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) (*db.RefreshTokenModel, error)
	FindRefreshToken(ctx context.Context, token string) (*db.RefreshTokenModel, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteUserRefreshTokens(ctx context.Context, userID string) error
}

type authRepository struct {
	client *db.PrismaClient
}

// NewAuthRepository creates a new AuthRepository instance.
func NewAuthRepository(client *db.PrismaClient) AuthRepository {
	return &authRepository{
		client: client,
	}
}

func (r *authRepository) FindByEmail(ctx context.Context, email string) (*db.UserModel, error) {
	return r.client.User.FindUnique(db.User.Email.Equals(email)).Exec(ctx)
}

func (r *authRepository) FindByID(ctx context.Context, id string) (*db.UserModel, error) {
	return r.client.User.FindUnique(db.User.ID.Equals(id)).Exec(ctx)
}

func (r *authRepository) CreateUser(ctx context.Context, email, password string, name, role *string) (*db.UserModel, error) {
	var optionalParams []db.UserSetParam
	optionalParams = append(optionalParams, db.User.Password.Set(password))
	if name != nil {
		optionalParams = append(optionalParams, db.User.Name.Set(*name))
	}
	if role != nil {
		optionalParams = append(optionalParams, db.User.Role.Set(db.Role(*role)))
	}

	return r.client.User.CreateOne(
		db.User.Email.Set(email),
		optionalParams...,
	).Exec(ctx)
}

func (r *authRepository) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) (*db.RefreshTokenModel, error) {
	return r.client.RefreshToken.CreateOne(
		db.RefreshToken.Token.Set(token),
		db.RefreshToken.User.Link(db.User.ID.Equals(userID)),
		db.RefreshToken.ExpiresAt.Set(expiresAt),
	).Exec(ctx)
}

func (r *authRepository) FindRefreshToken(ctx context.Context, token string) (*db.RefreshTokenModel, error) {
	return r.client.RefreshToken.FindUnique(
		db.RefreshToken.Token.Equals(token),
	).Exec(ctx)
}

func (r *authRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.client.RefreshToken.FindUnique(
		db.RefreshToken.Token.Equals(token),
	).Delete().Exec(ctx)
	return err
}

func (r *authRepository) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.client.RefreshToken.FindMany(
		db.RefreshToken.UserID.Equals(userID),
	).Delete().Exec(ctx)
	return err
}
