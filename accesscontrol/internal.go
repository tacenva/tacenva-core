package accesscontrol

import (
	"github.com/tacenva/tacenva-core/entity"
	"github.com/tacenva/tacenva-core/util/keyring"
)

func (s *Service) InitAccess(
	sotID string,
) (*keyring.KeyPair, error) {
	keyPairOfGroup, err := keyring.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	accessGroup := entity.AccessGroup{
		SotID:     sotID,
		Name:      "Tacenva",
		PublicKey: keyPairOfGroup.PublicKey,
	}

	keyPairOfUnit, err := keyring.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	adminUnit := entity.AccessUnit{
		Name:        accessGroup.Name,
		SotID:       sotID,
		Revoked:     false,
		AutoApprove: false,
		PublicKey:   keyPairOfUnit.PublicKey,
	}

	wrappedGroupKey, err := keyPairOfUnit.Seal([]byte(keyPairOfGroup.PrivateKey))
	if err != nil {
		return nil, err
	}

	if err = s.accessgroupService.CreateGroup(
		sotID,
		"",
		&accessGroup,
		&adminUnit,
		"",
		wrappedGroupKey,
		"",
	); err != nil {
		return nil, err
	}

	return keyPairOfUnit, nil
}
