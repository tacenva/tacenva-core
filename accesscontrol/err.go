package accesscontrol

import (
	"errors"

	"github.com/tacenva/tacenva-core/entity"
	"github.com/tacenva/tacenva-core/genericerror"
)

var (
	errAdminOnly         = errors.New("only administrators are allowed to perform this action")
	errGroupAccessDenied = errors.New("you do not have access to this group")
	errSelfAccessDenied  = errors.New("you are not allowed to perform this action on your own resources")
	errInvalidAuthorizer = genericerror.ErrEmptyValue
)

func verifyAdminAccess(authorizer *entity.User) error {
	if authorizer == nil {
		return errInvalidAuthorizer
	}

	if !authorizer.HasAdminAccess() {
		return errAdminOnly
	}

	return nil
}

func (s *Service) verifyGroupAccess(
	authorizer *entity.User,
	parentGroupID string,
	childGroupID string,
	selfAccessdenied bool,
) error {
	if err := verifyAdminAccess(authorizer); err != nil {
		return err
	}

	if selfAccessdenied {
		if childGroupID == authorizer.AccessUnit.GroupID {
			return errSelfAccessDenied
		}
	}

	isAccessible, err := s.accessgroupService.CanAccessGroup(
		authorizer.AccessUnit.SotID,
		parentGroupID,
		childGroupID,
	)
	if err != nil {
		return err
	}

	if !isAccessible {
		return errGroupAccessDenied
	}

	return nil
}

func (s *Service) verifyUnitAccess(
	authorizer *entity.User,
	unitID string,
	selfAccessdenied bool,
) error {
	if err := verifyAdminAccess(authorizer); err != nil {
		return err
	}

	toAccessGroup, err := s.accessgroupService.GetGroupByUnitID(
		authorizer.AccessUnit.SotID,
		unitID,
	)
	if err != nil {
		return err
	}

	if selfAccessdenied {
		if toAccessGroup.ID == authorizer.AccessUnit.GroupID {
			return errSelfAccessDenied
		}
	}

	isAccessible, err := s.accessgroupService.CanAccessGroup(
		authorizer.AccessUnit.SotID,
		authorizer.AccessUnit.GroupID,
		toAccessGroup.ID,
	)
	if err != nil {
		return err
	}

	if !isAccessible {
		return errGroupAccessDenied
	}

	return nil
}

func (s *Service) verifyUserAccess(
	authorizer *entity.User,
	userID string,
) error {
	if err := verifyAdminAccess(authorizer); err != nil {
		return err
	}

	if authorizer.ID == userID {
		return errSelfAccessDenied
	}

	toAccessGroup, err := s.accessgroupService.GetGroupByUserID(
		authorizer.AccessUnit.SotID,
		userID,
	)
	if err != nil {
		return err
	}

	isAccessible, err := s.accessgroupService.CanAccessGroup(
		authorizer.AccessUnit.SotID,
		authorizer.AccessUnit.GroupID,
		toAccessGroup.ID,
	)
	if err != nil {
		return err
	}

	if !isAccessible {
		return errGroupAccessDenied
	}

	return nil
}
