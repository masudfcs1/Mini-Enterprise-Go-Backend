package auth

import (
	"context"

	"go-mini-setup/internal/database/db"
)

// AuthRepository defines database queries needed for authentication.
type AuthRepository interface {
	FindByEmail(ctx context.Context, email string) (*db.UserModel, error)
	CreateUser(ctx context.Context, email, password string, name *string) (*db.UserModel, error)
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

func (r *authRepository) CreateUser(ctx context.Context, email, password string, name *string) (*db.UserModel, error) {
	var optionalParams []db.UserSetParam
	optionalParams = append(optionalParams, db.User.Password.Set(password))
	if name != nil {
		optionalParams = append(optionalParams, db.User.Name.Set(*name))
	}

	return r.client.User.CreateOne(
		db.User.Email.Set(email),
		optionalParams...,
	).Exec(ctx)
}
