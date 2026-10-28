package accessunit

import (
	"github.com/tacenva/tacenva-core/entity"
	"gorm.io/gorm"
)

type Service struct {
	repository *repository
}

func NewService(repository *repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreateUnit(
	sotID string,
	authorizerGroupID string,
	accessUnit *entity.AccessUnit,
	wrappedGroupKey string,
) error {
	if accessUnit == nil {
		// return
	}

	if accessUnit.GroupID != authorizerGroupID {
		// return
	}

	accessUnit.SotID = sotID
	return s.repository.Transaction(func(tx *gorm.DB) error {
		err := s.repository.CreateUnit(tx, accessUnit)
		if err != nil {
			return err
		}

		accessGroupKey := entity.AccessGroupKey{
			SotID:         sotID,
			AccessGroupID: accessUnit.GroupID,
			AccessUnitID:  accessUnit.ID,
			WrappedKey:    wrappedGroupKey,
		}

		err = s.repository.CreateGroupKey(tx, &accessGroupKey)
		if err != nil {
			return err
		}

		return nil
	})
}

// func (s *Service) Simulate() error {
// 	keyPairForAccessGroup, _ := keyring.GenerateKeyPair()
// 	accessGroup := entity.AccessGroup{
// 		Name:      "Somethin",
// 		PublicKey: keyPairForAccessGroup.PublicKey,
// 	}

// 	keyPairForAccessUnit, _ := keyring.GenerateKeyPair()
// 	adminUnit := entity.AccessUnit{
// 		Name:        accessGroup.Name + " admin",
// 		Level:       entity.LevelAdmin,
// 		PublicKey:   keyPairForAccessUnit.PublicKey,
// 		Revoked:     false,
// 		AutoApprove: false,
// 	}

// 	pura2nyaparent, _ := keyring.GenerateKeyPair()
// 	wrappedGroupKey, _ := keyPairForAccessUnit.Seal([]byte(keyPairForAccessGroup.PrivateKey))
// 	wrappedChildKey, _ := pura2nyaparent.Seal([]byte(keyPairForAccessGroup.PrivateKey))

// 	return s.CreateGroup(
// 		"",
// 		"",
// 		&accessGroup,
// 		&adminUnit,
// 		"",
// 		wrappedGroupKey,
// 		wrappedChildKey,
// 	)
// }

func (s *Service) UpdateUnit(
	sotID string,
	authorizerGroupID string,
	accessUnit *entity.AccessUnit,
) error {
	if accessUnit == nil {
		// return
	}

	return s.repository.Transaction(func(tx *gorm.DB) error {
		current, err := s.repository.GetUnitByID(tx, accessUnit.ID, sotID)
		if err != nil {
			return err
		}

		if accessUnit.PublicKey != current.PublicKey && accessUnit.Revoked != current.Revoked {
			// return
		}
		accessUnit.Revoked = current.Revoked
		err = s.repository.UpdateUnit(tx, accessUnit)
		return err
	})
}

func (s *Service) DeleteUnit(
	sotID string,
	authorizerGroupID string,
	id string,
) error {
	return s.repository.Transaction(func(tx *gorm.DB) error {
		return s.repository.DeleteUnit(tx, id, sotID)
	})
}

func (s *Service) UnitsByGroup(
	sotID string,
	authorizerGroupID string,
	id string,
) ([]AccessUnitWithUserCount, error) {
	var units []AccessUnitWithUserCount

	err := s.repository.Transaction(func(tx *gorm.DB) error {
		result, err := s.repository.GetUnitsByGroup(
			tx,
			id,
			sotID,
		)
		if err != nil {
			return err
		}

		units = result
		return nil
	})
	if err != nil {
		return nil, err
	}

	return units, nil
}

func (s *Service) UsersByUnit(
	sotID string,
	authorizerGroupID string,
	unitID string,
) ([]entity.User, error) {
	var users []entity.User

	err := s.repository.Transaction(func(tx *gorm.DB) error {
		result, err := s.repository.GetUsersByUnit(
			tx,
			unitID,
			sotID,
		)
		if err != nil {
			return err
		}

		users = result
		return nil
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Service) SetRevokeUnit(
	sotID string,
	unitID string,
	isRevoked bool,
) error {
	return s.repository.Transaction(func(tx *gorm.DB) error {
		current, err := s.repository.GetUnitByID(tx, unitID, sotID)
		if err != nil {
			return err
		}
		current.Revoked = isRevoked
		err = s.repository.UpdateUnit(tx, current)
		return err
	})
}

func (s *Service) GetUnitByPublicKey(
	sotID string,
	publicKey string,
) (*entity.AccessUnit, error) {
	return s.repository.GetUnitByPublicKey(
		sotID,
		publicKey,
	)
}
