package user

import (
	"errors"
	"strings"

	"github.com/tacenva/tacenva-core/entity"
	"github.com/tacenva/tacenva-core/util/credential"
)

var (
	ErrNotFound         = errors.New("user not found")
	ErrSotIdentityEmpty = errors.New("user sot identity cannot be empty")
	ErrTokenEmpty       = errors.New("token cannot be empty")
	ErrStatusInvalid    = errors.New("invalid user status")
)

const tokenSize = 32

type Service struct {
	repository *repository
}

func NewService(repository *repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(
	sotIdentity string,
	accessUnitID string,
	userStatus entity.UserStatus,
) (*entity.User, string, error) {
	sotIdentity = strings.ToLower(strings.TrimSpace(sotIdentity))
	accessUnitID = strings.TrimSpace(accessUnitID)

	if sotIdentity == "" {
		return nil, "", ErrSotIdentityEmpty
	}

	token, err := credential.Generate(tokenSize)
	if err != nil {
		return nil, "", err
	}

	user := &entity.User{
		AccessUnitID: accessUnitID,
		SotIdentity:  sotIdentity,
		Token:        credential.Hash(token),
		Status:       userStatus,
	}

	if err := s.repository.Create(user); err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *Service) GetByToken(token string) (*entity.User, error) {
	token = strings.TrimSpace(token)

	if token == "" {
		return nil, ErrTokenEmpty
	}

	tokenHash := credential.Hash(token)

	user, err := s.repository.FindByToken(tokenHash)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrNotFound
	}

	return user, nil
}

func (s *Service) Update(
	id string,
	sotIdentity string,
) (*entity.User, error) {
	id = strings.TrimSpace(id)
	sotIdentity = strings.ToLower(strings.TrimSpace(sotIdentity))

	if id == "" {
		return nil, ErrNotFound
	}

	if sotIdentity == "" {
		return nil, ErrSotIdentityEmpty
	}

	user, err := s.repository.FindByID(id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrNotFound
	}

	user.SotIdentity = sotIdentity

	if err := s.repository.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdateStatus(
	id string,
	status entity.UserStatus,
) error {
	id = strings.TrimSpace(id)

	if id == "" {
		return ErrNotFound
	}

	switch status {
	case entity.UserStatusPending,
		entity.UserStatusApproved,
		entity.UserStatusRevoked:

	default:
		return ErrStatusInvalid
	}

	user, err := s.repository.FindByID(id)
	if err != nil {
		return err
	}

	if user == nil {
		return ErrNotFound
	}

	user.Status = status

	if err := s.repository.Update(user); err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(id string) error {
	id = strings.TrimSpace(id)

	if id == "" {
		return ErrNotFound
	}

	user, err := s.repository.FindByID(id)
	if err != nil {
		return err
	}

	if user == nil {
		return ErrNotFound
	}

	return s.repository.Delete(id)
}

func (s *Service) FindAllIDsExcept(userID string) ([]string, error) {
	return s.repository.FindAllIDsExcept(userID)
}
