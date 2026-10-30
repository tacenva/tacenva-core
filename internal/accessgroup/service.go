package accessgroup

import (
	"errors"

	"github.com/tacenva/tacenva-core/contract"
	"github.com/tacenva/tacenva-core/debug"
	"github.com/tacenva/tacenva-core/entity"
	"github.com/tacenva/tacenva-core/genericerror"
	"github.com/tacenva/tacenva-core/sync"
	"gorm.io/gorm"
)

type Service struct {
	repository *repository
	createUnit func(
		tx *gorm.DB,
		sotID string,
		accessUnit *entity.AccessUnit,
		wrappedGroupKey string,
	) error

	syncService    *sync.Service
	syncentityname string
}

func NewService(
	repository *repository,
	createUnit func(
		tx *gorm.DB,
		sotID string,
		accessUnit *entity.AccessUnit,
		wrappedGroupKey string,
	) error,
	syncService *sync.Service,
) *Service {
	return &Service{
		repository:     repository,
		createUnit:     createUnit,
		syncService:    syncService,
		syncentityname: "access_group",
	}
}

func (s *Service) GetLastChangeID(
	sotID string,
) (string, error) {
	syncState, err := s.syncService.GetSyncState(
		sotID,
		s.syncentityname,
	)
	if err != nil {
		return "", err
	}

	return syncState.LastChangeID, nil
}

func (s *Service) CreateGroup(
	sotID string,
	authorizerGroupID string,
	changeBy string,
	accessGroupRequest *contract.AccessGroupSetupRequest,
) error {
	debug.LogJSON("access group", accessGroupRequest)
	if accessGroupRequest.AdminUnit == nil || accessGroupRequest.AccessGroup == nil {
		return errAccessAssigned
	}

	return s.repository.db.Transaction(func(tx *gorm.DB) error {
		// 1: create group first
		accessGroupRequest.AccessGroup.SotID = sotID
		err := s.repository.CreateGroup(tx, accessGroupRequest.AccessGroup)
		if err != nil {
			return err
		}

		// 2: create hierarchy
		if accessGroupRequest.ParentGroupID != "" {
			accessGroupHierarchy := entity.AccessGroupHierarchy{
				SotID:         sotID,
				ParentGroupID: accessGroupRequest.ParentGroupID,
				ChildGroupID:  accessGroupRequest.AccessGroup.ID,
			}

			if accessGroupRequest.WrappedHierarchyKey != "" { // parent can access child
				accessGroupHierarchy.WrappedChildKey = accessGroupRequest.WrappedHierarchyKey
			}

			err = s.repository.CreateGroupHierarchy(tx, &accessGroupHierarchy)
			if err != nil {
				return err
			}
		}

		// 3: create admin (unit & groupkey)
		accessGroupRequest.AdminUnit.GroupID = accessGroupRequest.AccessGroup.ID
		accessGroupRequest.AdminUnit.Level = entity.LevelAdmin
		err = s.createUnit(
			tx,
			sotID,
			accessGroupRequest.AdminUnit,
			accessGroupRequest.WrappedGroupKey,
		)
		if err != nil {
			return err
		}

		// record change
		err = s.syncService.RecordChange(
			tx,
			nil,
			s.syncentityname,
			accessGroupRequest.AccessGroup.ID,
			entity.OperationChange,
			sotID,
			changeBy,
		)
		if err != nil {
			return err
		}

		err = s.syncService.SetSyncStateToBeLastest(
			tx,
			sotID,
			s.syncentityname,
		)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) UpdateGroup(
	sotID string,
	authorizerGroupID string,
	changeBy string,
	accessGroup *entity.AccessGroup,
) error {
	if accessGroup == nil {
		return genericerror.ErrEmptyValue
	}

	return s.repository.db.Transaction(func(tx *gorm.DB) error {
		current, err := s.repository.GetGroupByID(tx, accessGroup.ID, sotID)
		if err != nil {
			return err
		}

		if accessGroup.PublicKey != "" && accessGroup.PublicKey != current.PublicKey {
			return genericerror.ErrImutableField
		}

		current.Name = accessGroup.Name
		err = s.repository.UpdateGroup(tx, current)
		if err != nil {
			return err
		}

		// record change
		err = s.syncService.RecordChange(
			tx,
			nil,
			s.syncentityname,
			current.ID,
			entity.OperationChange,
			sotID,
			changeBy,
		)
		if err != nil {
			return err
		}

		err = s.syncService.SetSyncStateToBeLastest(
			tx,
			sotID,
			s.syncentityname,
		)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) DeleteGroup(
	sotID string,
	authorizerGroupID string,
	changeBy string,
	id string,
) error {
	return s.repository.db.Transaction(func(tx *gorm.DB) error {
		err := s.repository.DeleteGroup(tx, id, sotID)
		if err != nil {
			return err
		}

		// record change
		err = s.syncService.RecordChange(
			tx,
			nil,
			s.syncentityname,
			id,
			entity.OperationChange,
			sotID,
			changeBy,
		)
		if err != nil {
			return err
		}

		err = s.syncService.SetSyncStateToBeLastest(
			tx,
			sotID,
			s.syncentityname,
		)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) GroupWithChildren(
	sotID string,
	id string,
) ([]contract.AccessGroupTree, error) {
	var groups []contract.AccessGroupTree

	err := s.repository.db.Transaction(func(tx *gorm.DB) error {
		result, err := s.repository.GetGroupWithChildren(
			tx,
			id,
			sotID,
		)
		if err != nil {
			return err
		}

		groups = result
		return nil
	})
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (s *Service) List(
	sotID string,
) ([]entity.AccessGroup, error) {
	var groups []entity.AccessGroup
	err := s.repository.db.Transaction(func(tx *gorm.DB) error {
		result, err := s.repository.FindAll(tx, sotID)
		if err != nil {
			return err
		}
		groups = result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (s *Service) PublicKeyList(
	sotID string,
	groupIDs []string,
) (map[string]string, error) {
	publicKeys := make(map[string]string, len(groupIDs))

	err := s.repository.db.Transaction(func(tx *gorm.DB) error {
		groups, err := s.repository.FindKeysByIDs(tx, sotID, groupIDs)
		if err != nil {
			return err
		}

		for _, group := range groups {
			publicKeys[group.ID] = group.PublicKey
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return publicKeys, nil
}

func (s *Service) SyncGroups(
	sotID string,
	groups []entity.AccessGroup,
) error {
	if sotID == "" {
		return errors.New("sot ID cannot be empty")
	}

	// Validasi input dan cegah ID duplikat.
	incoming := make(map[string]entity.AccessGroup, len(groups))

	for _, group := range groups {
		if group.ID == "" {
			return errors.New("group ID cannot be empty")
		}
		if group.Name == "" {
			return errors.New("group name cannot be empty")
		}
		if group.PublicKey != "" {
			return errors.New("group public key must be empty")
		}
		if _, exists := incoming[group.ID]; exists {
			return errors.New("duplicate group ID: " + group.ID)
		}

		incoming[group.ID] = group
	}

	return s.repository.db.Transaction(func(tx *gorm.DB) error {
		existing, err := s.repository.FindAll(tx, sotID)
		if err != nil {
			return err
		}

		existingByID := make(map[string]entity.AccessGroup, len(existing))
		for _, group := range existing {
			existingByID[group.ID] = group
		}

		var toCreate []entity.AccessGroup
		var toUpdate []entity.AccessGroup
		var toDelete []string

		for id, group := range incoming {
			current, exists := existingByID[id]
			if !exists {
				toCreate = append(toCreate, group)
				continue
			}

			if current.Name != group.Name ||
				current.PublicKey != group.PublicKey {
				toUpdate = append(toUpdate, group)
			}
		}

		for id := range existingByID {
			if _, exists := incoming[id]; !exists {
				toDelete = append(toDelete, id)
			}
		}

		if err := s.repository.CreateGroupsBulk(tx, sotID, toCreate); err != nil {
			return err
		}

		if err := s.repository.UpdateGroupsBulk(tx, sotID, toUpdate); err != nil {
			return err
		}

		if err := s.repository.DeleteGroupsBulk(tx, sotID, toDelete); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) CanAccessGroup(
	sotID string,
	parentGroupID string,
	childGroupID string,
) (bool, error) {
	if sotID == "" || parentGroupID == "" || childGroupID == "" {
		return false, errors.New("sot id and group ids are required")
	}

	if parentGroupID == childGroupID {
		return true, nil
	}

	return s.repository.CanAccessGroup(
		sotID,
		parentGroupID,
		childGroupID,
	)
}

func (s *Service) GetGroupByUnitID(
	sotID string,
	unitID string,
) (*entity.AccessGroup, error) {
	return s.repository.GetGroupByUnitID(sotID, unitID)
}

func (s *Service) GetGroupByID(
	sotID string,
	groupID string,
) (*entity.AccessGroup, error) {
	var group *entity.AccessGroup
	err := s.repository.db.Transaction(func(tx *gorm.DB) error {
		found, err := s.repository.GetGroupByID(tx, groupID, sotID)
		if err != nil {
			return err
		}
		group = found
		return nil
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *Service) GetGroupByUserID(
	sotID string,
	userID string,
) (*entity.AccessGroup, error) {
	return s.repository.GetGroupByUserID(sotID, userID)
}
