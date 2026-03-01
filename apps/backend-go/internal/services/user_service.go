package services

import (
	"context"
	"database/sql"
	"errors"

	sqliteeev "backend-go/internal/db/sqlite/generated"

	"github.com/google/uuid"
)

type UserService struct {
	q *sqliteeev.Queries
}

func NewUserService(q *sqliteeev.Queries) *UserService {
	return &UserService{q: q}
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*sqliteeev.User, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &sqliteeev.User{
		Email:     user.Email,
		ID:        user.ID,
		AvatarUrl: user.AvatarUrl,
		Source:    user.Source,
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*sqliteeev.User, error) {
	user, err := s.q.GetUserById(ctx, id.String())
	if err != nil {
		return nil, err
	}

	return &sqliteeev.User{
		Email:     user.Email,
		ID:        user.ID,
		AvatarUrl: user.AvatarUrl,
		Source:    user.Source,
	}, nil
}

func (s *UserService) CreateUser(ctx context.Context, email string, source string, avatar *string) (*sqliteeev.User, error) {
	id := uuid.New()

	user, err := s.q.CreateUser(ctx, sqliteeev.CreateUserParams{
		ID:        id.String(),
		Email:     email,
		Source:    source,
		AvatarUrl: sql.NullString{String: *avatar, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	if user.ID == "" {
		return nil, errors.New("failed to create user")
	}

	return &sqliteeev.User{
		Email:     user.Email,
		ID:        user.ID,
		AvatarUrl: user.AvatarUrl,
		Source:    user.Source,
	}, nil
}
