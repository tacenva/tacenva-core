package accessunit

import "github.com/tacenva/tacenva-core/entity"

func (s *Service) UsersByUnit(
	sotID string,
	unitID string,
) ([]entity.User, error) {
	result, err := s.repository.GetUsersByUnit(
		unitID,
		sotID,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) SetApprovalUser(
	userID string,
	isRevoked bool,
) error {
	err := s.repository.SetApprovalUser(
		userID,
		isRevoked,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteUser(
	userID string,
) error {
	err := s.repository.DeleteUser(
		userID,
	)
	if err != nil {
		return err
	}

	return nil
}
