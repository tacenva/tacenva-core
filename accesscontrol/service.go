package accesscontrol

import (
	"errors"

	"github.com/tacenva/tacenva-core/contract"
	"github.com/tacenva/tacenva-core/entity"
	"github.com/tacenva/tacenva-core/genericerror"
	"github.com/tacenva/tacenva-core/internal/accessgroup"
	"github.com/tacenva/tacenva-core/internal/accessunit"
	"github.com/tacenva/tacenva-core/internal/user"
)

type Service struct {
	userService        *user.Service
	accessunitService  *accessunit.Service
	accessgroupService *accessgroup.Service

	Internal *Internal
}

func NewService(
	userService *user.Service,
	accessunitService *accessunit.Service,
	accessgroupService *accessgroup.Service,
) *Service {
	return &Service{
		userService:        userService,
		accessunitService:  accessunitService,
		accessgroupService: accessgroupService,

		Internal: NewInternalService(
			userService,
			accessunitService,
			accessgroupService,
		),
	}
}

func (s *Service) CreateGroup(
	authorizer *entity.User,
	accessGroupReq *contract.AccessGroupSetupRequest,
) error {
	if err := verifyAdminAccess(authorizer); err != nil {
		return err
	}

	if accessGroupReq == nil {
		return genericerror.ErrEmptyValue
	}
	if accessGroupReq.ParentGroupID == "" {
		return errors.New("No selected parent group")
	}

	return s.accessgroupService.CreateGroup(
		authorizer.AccessUnit.SotID,
		authorizer.AccessUnit.GroupID,
		authorizer.ID+":"+authorizer.Hostname,
		accessGroupReq,
	)
}

func (s *Service) UpdateGroup(
	authorizer *entity.User,
	accessGroup *entity.AccessGroup,
) error {
	if err := s.verifyGroupAccess(
		authorizer,
		authorizer.AccessUnit.GroupID,
		accessGroup.ID,
		false,
	); err != nil {
		return err
	}

	return s.accessgroupService.UpdateGroup(
		authorizer.AccessUnit.SotID,
		authorizer.AccessUnit.GroupID,
		authorizer.ID+":"+authorizer.Hostname,
		accessGroup,
	)
}

func (s *Service) DeleteGroup(
	authorizer *entity.User,
	id string,
) error {
	if err := s.verifyGroupAccess(
		authorizer,
		authorizer.AccessUnit.GroupID,
		id,
		true,
	); err != nil {
		return err
	}

	return s.accessgroupService.DeleteGroup(
		authorizer.AccessUnit.SotID,
		authorizer.AccessUnit.GroupID,
		authorizer.ID+":"+authorizer.Hostname,
		id,
	)
}

func (s *Service) GroupWithChildren(
	authorizer *entity.User,
	id string,
) ([]contract.AccessGroupTree, error) {
	if err := verifyAdminAccess(authorizer); err != nil {
		return nil, err
	}
	return s.accessgroupService.GroupWithChildren(
		authorizer.AccessUnit.SotID,
		id,
	)
}

func (s *Service) List(
	sotID string,
) ([]entity.AccessGroup, error) {
	return s.accessgroupService.List(
		sotID,
	)
}

func (s *Service) PublicKeyOfGroups(
	authorizer *entity.User,
	groupIDS []string,
) (map[string]string, error) {
	if authorizer.HasWriteAccess() {
		return nil, errors.New("only users with write access are allowed to perform this action")

	}
	return s.accessgroupService.PublicKeyList(
		authorizer.AccessUnit.SotID,
		groupIDS,
	)
}

func (s *Service) SyncGroups(
	sotID string,
	groups []entity.AccessGroup,
) error {
	return s.accessgroupService.SyncGroups(sotID, groups)
}

func (s *Service) CreateUnit(
	authorizer *entity.User,
	accessUnitReq *contract.AccessUnitSetupReq,
) error {
	if err := s.verifyGroupAccess(
		authorizer,
		authorizer.AccessUnit.GroupID,
		accessUnitReq.AccessUnit.GroupID,
		false,
	); err != nil {
		return err
	}
	return s.accessunitService.CreateUnit(
		authorizer.AccessUnit.SotID,
		accessUnitReq,
	)
}

func (s *Service) UpdateUnit(
	authorizer *entity.User,
	accessUnit *entity.AccessUnit,
) error {
	if err := s.verifyUnitAccess(
		authorizer,
		accessUnit.ID,
		false,
	); err != nil {
		return err
	}
	return s.accessunitService.UpdateUnit(
		authorizer.AccessUnit.SotID,
		accessUnit,
	)
}

func (s *Service) DeleteUnit(
	authorizer *entity.User,
	unitID string,
) error {
	if err := s.verifyUnitAccess(
		authorizer,
		unitID,
		true,
	); err != nil {
		return err
	}
	return s.accessunitService.DeleteUnit(
		authorizer.AccessUnit.SotID,
		unitID,
	)
}

func (s *Service) SetRevokeUnit(
	authorizer *entity.User,
	unitID string,
	isRevoked bool,
) error {
	if err := s.verifyUnitAccess(
		authorizer,
		unitID,
		true,
	); err != nil {
		return err
	}
	return s.accessunitService.SetRevokeUnit(
		authorizer.AccessUnit.SotID,
		unitID,
		isRevoked,
	)
}

func (s *Service) UnitList(
	authorizer *entity.User,
	groupID string,
) ([]contract.AccessUnitWithUserCount, error) {
	if err := verifyAdminAccess(authorizer); err != nil {
		return nil, err
	}
	return s.accessunitService.UnitsByGroup(authorizer.AccessUnit.SotID, groupID)
}

func (s *Service) SetApprovalUser(
	authorizer *entity.User,
	userID string,
	isRevoked bool,
) error {
	if err := s.verifyUserAccess(
		authorizer,
		userID,
	); err != nil {
		return err
	}
	return s.accessunitService.SetApprovalUser(
		userID,
		isRevoked,
	)
}

func (s *Service) DeleteUser(
	authorizer *entity.User,
	userID string,
) error {
	if err := s.verifyUserAccess(
		authorizer,
		userID,
	); err != nil {
		return err
	}
	return s.accessunitService.DeleteUser(
		userID,
	)
}
