package user

import (
	"context"

	"go-mini-setup/pkg/errors"
)

// UserService defines the business logic interface for users.
type UserService interface {
	GetUsers(ctx context.Context) ([]UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*UserResponse, error)
	CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error)
	UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
}

type userService struct {
	repo UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(repo UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) GetUsers(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return ToUserResponseList(users), nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*UserResponse, error) {
	if id == "" {
		return nil, errors.NewBadRequestError("user id is required")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	res := ToUserResponse(user)
	return &res, nil
}

func (s *userService) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, _ := s.repo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.NewConflictError("user with this email already exists")
	}

	user, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	res := ToUserResponse(user)
	return &res, nil
}

func (s *userService) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*UserResponse, error) {
	if id == "" {
		return nil, errors.NewBadRequestError("user id is required")
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	if req.Email != nil && *req.Email != existing.Email {
		emailConflict, _ := s.repo.FindByEmail(ctx, *req.Email)
		if emailConflict != nil {
			return nil, errors.NewConflictError("another user already uses this email")
		}
	}

	user, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	res := ToUserResponse(user)
	return &res, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return errors.NewBadRequestError("user id is required")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.NewNotFoundError("user not found")
	}

	_, err = s.repo.Delete(ctx, id)
	return err
}
