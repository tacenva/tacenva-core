package accesscontrol

import (
	"github.com/tacenva/tacenva-core/contract"
	"github.com/tacenva/tacenva-core/entity"
	"github.com/tacenva/tacenva-core/internal/accessgroup"
	"github.com/tacenva/tacenva-core/internal/accessunit"
	"github.com/tacenva/tacenva-core/internal/user"
)

type Service struct {
	userService        *user.Service
	accessunitService  *accessunit.Service
	accessgroupService *accessgroup.Service
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
	}
}

func (s *Service) CreateGroup(
	authorizer *entity.User,
	accessGroup *entity.AccessGroup,
	adminUnit *entity.AccessUnit,
	parentGroupID string,
	wrappedGroupKey string,
	wrappedChildKey string, // optional
) error {
	if parentGroupID == "" {
		// return
	}

	return s.accessgroupService.CreateGroup(
		authorizer.AccessUnit.SotID,
		authorizer.AccessUnit.GroupID,
		accessGroup,
		adminUnit,
		parentGroupID,
		wrappedGroupKey,
		wrappedChildKey,
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
