package sync

import (
	"errors"
	"fmt"

	"github.com/tacenva/tacenva-core/entity"
	"github.com/tacenva/tacenva-core/genericerror"
	"github.com/tacenva/tacenva-core/internal/user"
	"gorm.io/gorm"
)

const defaultSyncChangeLimit = 500

type Service struct {
	repository  *Repository
	userService *user.Service
}

func NewService(
	db *gorm.DB,
	userService *user.Service,
) *Service {
	return &Service{
		repository:  NewRepository(db),
		userService: userService,
	}
}

func (s *Service) RecordChange(
	tx *gorm.DB,
	syncChangeID *string,
	entityName string,
	entityID string,
	operation entity.SyncOperation,
	sotID string,
	changedBy string,
) error {
	if entityName == "" {
		return fmt.Errorf("sync entity is required")
	}
	if entityID == "" {
		return fmt.Errorf("sync entity id is required")
	}
	if sotID == "" {
		return fmt.Errorf("sot id is required")
	}
	if tx == nil {
		return fmt.Errorf("sync transaction is required")
	}

	switch operation {
	case entity.OperationCreate,
		entity.OperationUpdate,
		entity.OperationDelete,
		entity.OperationChange:
	default:
		return fmt.Errorf("invalid sync operation %q", operation)
	}

	updateState := syncChangeID != nil && *syncChangeID != ""
	if updateState {
		exists, err := s.repository.SyncChangeExists(
			tx,
			*syncChangeID,
			sotID,
		)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf(
				"sync change %q not found for SoT %q",
				*syncChangeID,
				sotID,
			)
		}
	}

	change := entity.SyncChange{
		Entity:    entityName,
		EntityID:  entityID,
		Operation: operation,
		SotID:     sotID,
		ChangedBy: changedBy,
	}

	if err := s.repository.CreateSyncChange(tx, &change); err != nil {
		return err
	}

	if !updateState {
		return nil
	}

	state := s.repository.NewSyncState(
		sotID,
		entityName,
		change.ID,
	)

	if err := s.repository.UpsertSyncState(tx, &state); err != nil {
		return fmt.Errorf("update sync state: %w", err)
	}

	return nil
}

func (s *Service) getChangesAfter(
	sotID string,
	entityName string,
	startID string,
	total int64,
) ([]entity.SyncChange, error) {
	if total <= 0 {
		return nil, nil
	}

	const maxLimit int64 = 1<<31 - 1

	limit := total
	if limit > maxLimit {
		limit = maxLimit
	}

	batch, err := s.repository.GetChangesAfter(
		sotID,
		entityName,
		startID,
		int(limit),
	)
	if err != nil {
		return nil, err
	}

	if len(batch) == 0 {
		return nil, nil
	}

	if int64(len(batch)) >= total {
		return batch, nil
	}

	remaining, err := s.getChangesAfter(
		sotID,
		entityName,
		batch[len(batch)-1].ID,
		total-int64(len(batch)),
	)
	if err != nil {
		return nil, err
	}

	return append(batch, remaining...), nil
}

func (s *Service) GetChangesAfter(
	sotID string,
	entityName string,
	startID string,
	limit int64,
) ([]entity.SyncChange, error) {
	if sotID == "" {
		return nil, fmt.Errorf("sot id is required")
	}
	if entityName == "" {
		return nil, fmt.Errorf("sync entity is required")
	}
	if limit <= 0 {
		limit = defaultSyncChangeLimit
	}

	return s.getChangesAfter(
		sotID,
		entityName,
		startID,
		limit,
	)
}

func (s *Service) GetSyncState(
	sotID string,
	entityName string,
) (entity.SyncState, error) {
	if sotID == "" {
		return entity.SyncState{}, fmt.Errorf("sot id is required")
	}
	if entityName == "" {
		return entity.SyncState{}, fmt.Errorf("sync entity is required")
	}

	state, err := s.repository.GetSyncState(sotID, entityName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.SyncState{}, nil
		}
		return entity.SyncState{}, fmt.Errorf("get sync state: %w", err)
	}

	return state, nil
}

func (s *Service) GetChangeCount(
	sotID string,
	entityName string,
	startID string,
) (int64, error) {
	if sotID == "" {
		return 0, fmt.Errorf("sot id is required")
	}
	if entityName == "" {
		return 0, fmt.Errorf("sync entity is required")
	}
	// if startID == "" {
	// 	return 0, nil
	// }
	return s.repository.GetChangeCount(
		sotID,
		entityName,
		startID,
	)
}

func (s *Service) Transaction(
	fn func(tx *gorm.DB) error,
) error {
	return s.repository.Transaction(fn)
}

func (s *Service) InsertSyncChanges(
	tx *gorm.DB,
	sotID string,
	entityName string,
	changes []entity.SyncChange,
) error {
	if tx == nil {
		return fmt.Errorf("sync transaction is required")
	}
	if sotID == "" {
		return fmt.Errorf("sot id is required")
	}
	if entityName == "" {
		return fmt.Errorf("sync entity is required")
	}
	if len(changes) == 0 {
		return nil
	}

	for i := range changes {
		changes[i].SotID = sotID
	}

	if err := s.repository.CreateSyncChanges(tx, changes); err != nil {
		return err
	}

	return s.SetSyncStateToBeLatest(tx, sotID, entityName)
}

func (s *Service) SetSyncStateToBeLatest(
	tx *gorm.DB,
	sotID string,
	entityName string,
) error {
	if tx == nil {
		return fmt.Errorf("sync transaction is required")
	}
	if sotID == "" {
		return fmt.Errorf("sot id is required")
	}
	if entityName == "" {
		return fmt.Errorf("sync entity is required")
	}

	lastChange, err := s.repository.GetLatestSyncChange(
		tx,
		sotID,
		entityName,
	)
	if err != nil {
		return err
	}

	state := s.repository.NewSyncState(
		sotID,
		entityName,
		lastChange.ID,
	)

	if err := s.repository.UpsertSyncState(tx, &state); err != nil {
		return fmt.Errorf("update sync state: %w", err)
	}

	return nil
}

func (s *Service) GetLatestSyncChange(
	sotID string,
	entityName string,
) (*entity.SyncChange, error) {
	if err := genericerror.ValidateRequired(map[string]any{
		"sot id":      sotID,
		"entity name": entityName,
	}); err != nil {
		return nil, err
	}

	state, err := s.repository.GetSyncState(sotID, entityName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return s.repository.GetLatestSyncChangeByState(&state)
}
