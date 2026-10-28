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
