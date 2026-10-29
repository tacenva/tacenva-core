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
	if accessGroupReq == nil {
		return genericerror.ErrEmptyValue
	}
	if accessGroupReq.ParentGroupID == "" {
		return errors.New("No selected parent group")
	}

	return s.accessgroupService.CreateGroup(
		authorizer.AccessUnit.SotID,
		authorizer.AccessUnit.GroupID,
		accessGroupReq,
	)
}

func (s *Service) UpdateGroup(
	authorizer *entity.User,
	accessGroup *entity.AccessGroup,
) error {
	return s.accessgroupService.UpdateGroup(
		authorizer.AccessUnit.SotID,
		authorizer.AccessUnit.GroupID,
		accessGroup,
	)
}

func (s *Service) DeleteGroup(
	authorizer *entity.User,
	id string,
) error {
	return s.accessgroupService.DeleteGroup(
		authorizer.AccessUnit.SotID,
		authorizer.AccessUnit.GroupID,
		id,
	)
}

func (s *Service) GroupWithChildren(
	authorizer *entity.User,
	id string,
) ([]contract.AccessGroupTree, error) {
	// if authorizer
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
	sotID string,
	accessUnitReq *contract.AccessUnitSetupReq,
) error {
	return s.accessunitService.CreateUnit(
		sotID,
		accessUnitReq,
	)
}

func (s *Service) UpdateUnit(
	sotID string,
	accessUnit *entity.AccessUnit,
) error {
	return s.accessunitService.UpdateUnit(
		sotID,
		accessUnit,
	)
}

func (s *Service) DeleteUnit(
	sotID string,
	unitID string,
) error {
	return s.accessunitService.DeleteUnit(
		sotID,
		unitID,
	)
}

func (s *Service) SetRevokeUnit(
	sotID string,
	unitID string,
	isRevoked bool,
) error {
	return s.accessunitService.SetRevokeUnit(
		sotID,
		unitID,
		isRevoked,
	)
}

func (s *Service) SetRevokeStatus(
	sotID string,
	unitID string,
	isRevoked bool,
) error {
	return s.accessunitService.SetRevokeUnit(
		sotID,
		unitID,
		isRevoked,
	)
}

func (s *Service) UnitList(
	sotID string,
	groupID string,
) ([]contract.AccessUnitWithUserCount, error) {
	return s.accessunitService.UnitsByGroup(sotID, groupID)
}

func (s *Service) SetApprovalUser(
	userID string,
	isRevoked bool,
) error {
	return s.accessunitService.SetApprovalUser(
		userID,
		isRevoked,
	)
}

func (s *Service) DeleteUser(
	userID string,
) error {
	return s.accessunitService.DeleteUser(
		userID,
	)
}
