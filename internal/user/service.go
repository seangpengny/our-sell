package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/vtech/our-sell/internal/platform/apperror"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Get(ctx context.Context, id uuid.UUID) (User, error) {
	account, err := s.repository.ByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	return account, nil
}

func (s *Service) GetAuthenticated(ctx context.Context, id string) (User, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return User{}, apperror.ErrUnauthorized
	}
	return s.Get(ctx, parsed)
}
