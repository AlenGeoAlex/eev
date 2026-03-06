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

type TargetUserEmails struct {
	Email     string `json:"email"`
	IsStarred bool   `json:"is_starred"`
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

func (s *UserService) GetTargetEmailsOfUser(ctx context.Context, userId uuid.UUID, searchTerm *string) ([]TargetUserEmails, error) {
	search := ""
	if searchTerm != nil {
		search = *searchTerm
	}

	targetUsers, err := s.q.GetTargetEmailsForUser(ctx, sqliteeev.GetTargetEmailsForUserParams{
		UserID:  userId.String(),
		Column2: sql.NullString{String: search, Valid: true},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return make([]TargetUserEmails, 0), nil
	}

	result := make([]TargetUserEmails, len(targetUsers))
	for i, user := range targetUsers {
		starred := false
		if user.Starred == 1 {
			starred = true
		}
		result[i] = TargetUserEmails{
			Email:     user.TargetEmail,
			IsStarred: starred,
		}
	}

	return result, nil
}
