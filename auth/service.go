package auth

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
	sotHostname        string
}

func NewService(
	userService *user.Service,
	accessunitService *accessunit.Service,
	accessgroupService *accessgroup.Service,
	sotHostname string,
) *Service {
	return &Service{
		userService:        userService,
		accessunitService:  accessunitService,
		accessgroupService: accessgroupService,
		sotHostname:        sotHostname,
	}
}

func (s *Service) Enroll(
	enrollRequest *contract.EnrollRequest,
	sotID string,
	userStatus entity.UserStatus,
) (*contract.EnrollResponse, error) {
	accessUnit, err := s.accessunitService.GetUnitByPublicKey(
		sotID,
		enrollRequest.PublicKey,
	)
	if err != nil {
		return nil, err
	}

	if userStatus == entity.UserStatusPending && accessUnit.AutoApprove {
		userStatus = entity.UserStatusApproved
	}

	user, token, err := s.userService.Create(
		enrollRequest.Hostname,
		accessUnit.ID,
		userStatus,
	)
	if err != nil {
		return nil, err
	}

	response := contract.EnrollResponse{
		SotHostname: s.sotHostname,
		Status:      userStatus,
		AuthToken:   token,
		AccessLevel: accessUnit.Level,
	}

	if userStatus == entity.UserStatusApproved {
		var groups []entity.AccessGroup

		if accessUnit.Level != entity.LevelRead {
			groups, err = s.accessgroupService.List(sotID)
			if err != nil {
				return nil, err
			}
		}

		response.GroupAssignment = &contract.GroupAssignment{
			ID:    user.AccessUnit.GroupID,
			Items: groups,
		}
	}

	return &response, nil
}

func (s *Service) GetUserData(token string) (*entity.User, error) {
	userData, err := s.userService.GetByToken(token)
	if err != nil {
		return nil, err
	}

	if userData.AccessUnit.Revoked {
		// return nil, permission.ErrRevoked
	}

	switch userData.Status {
	case entity.UserStatusApproved:
		return userData, nil

	case entity.UserStatusPending:
		return nil, user.ErrPending

	case entity.UserStatusRevoked:
		return nil, user.ErrRevoked

	default:
		return nil, user.ErrStatusInvalid
	}
}

func (s *Service) GetAuthDataWithGroupList(
	authorizedUser *entity.User,
	sotID string,
	accessGroupLastChangeID string,
) (*contract.AuthDataResp, error) {
	authData := contract.AuthDataResp{
		Status:      authorizedUser.Status,
		AccessLevel: authorizedUser.AccessUnit.Level,
	}

	if authorizedUser.AccessUnit.Revoked {
		authData.Status = entity.UserStatusRevoked
	}

	var (
		groups         []entity.AccessGroup
		lastSyncChange *entity.SyncChange
		groupKeyData   *contract.GroupKeyData
	)
	if authorizedUser.HasWriteAccess() {
		var err error
		groups, err = s.accessgroupService.List(sotID)
		if err != nil {
			return nil, err
		}

		lastSyncChange, err = s.accessgroupService.GetLatestSyncChange(sotID)
		if err != nil {
			return nil, err
		}

		groupKeyData, err = s.accessunitService.GetGroupKeyByUserID(
			authorizedUser.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	authData.GroupAssignment = &contract.GroupAssignment{
		ID:             authorizedUser.AccessUnit.GroupID,
		LastSyncChange: *lastSyncChange,
		GroupKeyData:   *groupKeyData,
		Items:          groups,
	}

	return &authData, nil
}
