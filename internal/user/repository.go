package user

import (
	"context"

	"go-mini-setup/internal/database/db"
)

// UserRepository defines the database operations for users.
type UserRepository interface {
	FindAll(ctx context.Context) ([]UserModel, error)
	FindByID(ctx context.Context, id string) (*UserModel, error)
	FindByEmail(ctx context.Context, email string) (*UserModel, error)
	Create(ctx context.Context, req CreateUserRequest) (*UserModel, error)
	Update(ctx context.Context, id string, req UpdateUserRequest) (*UserModel, error)
	Delete(ctx context.Context, id string) (*UserModel, error)
}

type userRepository struct {
	client *db.PrismaClient
}

// NewUserRepository creates a new UserRepository backed by Prisma.
func NewUserRepository(client *db.PrismaClient) UserRepository {
	return &userRepository{
		client: client,
	}
}

func (r *userRepository) FindAll(ctx context.Context) ([]UserModel, error) {
	return r.client.User.FindMany().Exec(ctx)
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*UserModel, error) {
	return r.client.User.FindUnique(db.User.ID.Equals(id)).Exec(ctx)
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*UserModel, error) {
	return r.client.User.FindUnique(db.User.Email.Equals(email)).Exec(ctx)
}

func (r *userRepository) Create(ctx context.Context, req CreateUserRequest) (*UserModel, error) {
	var optionalParams []db.UserSetParam
	if req.Name != nil {
		optionalParams = append(optionalParams, db.User.Name.Set(*req.Name))
	}
	if req.Password != nil {
		optionalParams = append(optionalParams, db.User.Password.Set(*req.Password))
	}
	if req.Role != nil {
		optionalParams = append(optionalParams, db.User.Role.Set(db.Role(*req.Role)))
	}

	return r.client.User.CreateOne(
		db.User.Email.Set(req.Email),
		optionalParams...,
	).Exec(ctx)
}

func (r *userRepository) Update(ctx context.Context, id string, req UpdateUserRequest) (*UserModel, error) {
	var params []db.UserSetParam
	if req.Email != nil {
		params = append(params, db.User.Email.Set(*req.Email))
	}
	if req.Name != nil {
		params = append(params, db.User.Name.Set(*req.Name))
	}
	if req.Role != nil {
		params = append(params, db.User.Role.Set(db.Role(*req.Role)))
	}

	return r.client.User.FindUnique(
		db.User.ID.Equals(id),
	).Update(params...).Exec(ctx)
}

func (r *userRepository) Delete(ctx context.Context, id string) (*UserModel, error) {
	return r.client.User.FindUnique(
		db.User.ID.Equals(id),
	).Delete().Exec(ctx)
}
