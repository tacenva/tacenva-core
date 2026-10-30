package accesscontrol

import (
	"github.com/tacenva/tacenva-core/contract"
	"github.com/tacenva/tacenva-core/entity"
	"github.com/tacenva/tacenva-core/internal/accessgroup"
	"github.com/tacenva/tacenva-core/internal/accessunit"
	"github.com/tacenva/tacenva-core/internal/user"
	"github.com/tacenva/tacenva-core/util/keyring"
)

type Internal struct {
	userService        *user.Service
	accessunitService  *accessunit.Service
	accessgroupService *accessgroup.Service
}

func NewInternalService(
	userService *user.Service,
	accessunitService *accessunit.Service,
	accessgroupService *accessgroup.Service,
) *Internal {
	return &Internal{
		userService:        userService,
		accessunitService:  accessunitService,
		accessgroupService: accessgroupService,
	}
}

func (s *Internal) InitAccess(
	sotID string,
) (keyPairOfUnit, keyPairOfGroup *keyring.KeyPair, err error) {
	keyPairOfGroup, err = keyring.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}

	accessGroup := entity.AccessGroup{
		SotID:     sotID,
		Name:      "Tacenva",
		PublicKey: keyPairOfGroup.PublicKey,
	}

	keyPairOfUnit, err = keyring.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}

	adminUnit := entity.AccessUnit{
		Name:        accessGroup.Name,
		SotID:       sotID,
		Revoked:     false,
		AutoApprove: false,
		PublicKey:   keyPairOfUnit.PublicKey,
	}

	wrappedGroupKey, err := keyPairOfUnit.Seal(
		[]byte(keyPairOfGroup.PrivateKey),
	)
	if err != nil {
		return nil, nil, err
	}

	err = s.accessgroupService.CreateGroup(
		sotID,
		"",
		"internal",
		&contract.AccessGroupSetupRequest{
			AccessGroup:     &accessGroup,
			AdminUnit:       &adminUnit,
			WrappedGroupKey: wrappedGroupKey,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return keyPairOfUnit, keyPairOfGroup, nil
}

func (s *Internal) AccessUnitList(
	sotID string,
) ([]entity.AccessUnit, error) {
	return s.accessunitService.List(sotID)
}

func (s *Internal) UserList(
	sotID,
	unitID string,
) ([]entity.User, error) {
	return s.accessunitService.UsersByUnit(
		sotID,
		unitID,
	)
}

func (s *Internal) SetUserStatus(
	userID string,
	status entity.UserStatus,
) error {
	return s.userService.UpdateStatus(
		userID,
		status,
	)
}
