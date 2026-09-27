package user

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/vtech/our-sell/internal/platform/apperror"
	"github.com/vtech/our-sell/internal/platform/validator"
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

func (s *Service) List(ctx context.Context, search, role string, page, pageSize int) (UserList, error) {
	search = strings.TrimSpace(search)
	if len(search) > 100 {
		return UserList{}, apperror.NewValidation(map[string]string{"search": "must be at most 100 characters"})
	}
	if page < 1 || page > 10000 {
		return UserList{}, apperror.NewValidation(map[string]string{"page": "must be between 1 and 10000"})
	}
	if pageSize < 1 || pageSize > 100 {
		return UserList{}, apperror.NewValidation(map[string]string{"page_size": "must be between 1 and 100"})
	}
	if role != "" && role != string(RoleUser) && role != string(RoleAdmin) {
		return UserList{}, apperror.NewValidation(map[string]string{"role": "must be user or admin"})
	}
	offset := int64(page-1) * int64(pageSize)
	if offset > math.MaxInt32 {
		return UserList{}, apperror.NewValidation(map[string]string{"page": "is too large"})
	}
	adminRepository, ok := s.repository.(AdminRepository)
	if !ok {
		return UserList{}, apperror.Wrap(errors.New("admin repository is not configured"), "list users")
	}
	accounts, total, err := adminRepository.List(ctx, search, role, int32(pageSize), int32(offset))
	if err != nil {
		return UserList{}, apperror.Wrap(err, "list users")
	}
	public := make([]PublicUser, 0, len(accounts))
	for _, account := range accounts {
		public = append(public, account.Public())
	}
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return UserList{Users: public, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages}, nil
}

func (s *Service) UpdateRole(ctx context.Context, actorID, targetID string, nextRole string) (User, error) {
	role := Role(strings.TrimSpace(nextRole))
	if role != RoleUser && role != RoleAdmin {
		return User{}, apperror.NewValidation(map[string]string{"role": "must be user or admin"})
	}
	actor, err := validator.UUID(actorID)
	if err != nil {
		return User{}, apperror.ErrUnauthorized
	}
	target, err := validator.UUID(targetID)
	if err != nil {
		return User{}, apperror.ErrNotFound
	}
	if actor == target && role != RoleAdmin {
		return User{}, apperror.NewValidation(map[string]string{"role": "you cannot remove your own admin access"})
	}
	adminRepository, ok := s.repository.(AdminRepository)
	if !ok {
		return User{}, apperror.Wrap(errors.New("admin repository is not configured"), "update user role")
	}
	previous, err := adminRepository.ByID(ctx, target)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return User{}, apperror.ErrNotFound
		}
		return User{}, apperror.Wrap(err, "load user before role update")
	}
	updated, err := adminRepository.UpdateRole(ctx, target, role)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return User{}, apperror.ErrNotFound
		}
		return User{}, apperror.Wrap(err, "update user role")
	}
	if target != actor && previous.Role != role {
		if err := adminRepository.RevokeSessions(ctx, target); err != nil {
			return User{}, apperror.Wrap(err, "revoke sessions after role update")
		}
	}
	return updated, nil
}
