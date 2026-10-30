package sync

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/tacenva/tacenva-core/entity"
	"github.com/tacenva/tacenva-core/internal/user"
	"gorm.io/gorm"
)

type Service struct {
	db          *gorm.DB
	userService *user.Service
}

func NewService(
	db *gorm.DB,
	userService *user.Service,
) *Service {
	return &Service{
		db:          db,
		userService: userService,
	}
}

func (s *Service) RecordChangeWithoutTx(
	syncChangeID *string,
	entity string,
	entityID string,
	operation entity.SyncOperation,
	sotID string,
	changedBy string,
) error {
	return s.Transaction(func(tx *gorm.DB) error {
		return s.RecordChange(
			tx,
			syncChangeID,
			entity,
			entityID,
			operation,
			sotID,
			changedBy,
		)
	})
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

	switch operation {
	case entity.OperationCreate,
		entity.OperationUpdate,
		entity.OperationDelete,
		entity.OperationChange:
	default:
		return fmt.Errorf(
			"invalid sync operation %q",
			operation,
		)
	}

	if tx == nil {
		return fmt.Errorf("sync transaction is required")
	}

	setSyncStateToBeLastest := false
	if syncChangeID != nil && *syncChangeID != "" {
		setSyncStateToBeLastest = true

		var count int64
		if err := tx.Model(&entity.SyncChange{}).
			Where("id = ? AND sot_id = ?", *syncChangeID, sotID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("query sync change: %w", err)
		}

		if count == 0 {
			return fmt.Errorf(
				"sync change %q not found for SoT %q",
				*syncChangeID,
				sotID,
			)
		}
	}

	if syncChangeID == nil {
		syncChangeID = new(string)
	}

	now := time.Now()
	change := entity.SyncChange{
		Entity:    entityName,
		EntityID:  entityID,
		Operation: operation,
		SotID:     sotID,

		CreatedAt: now,
		ChangedBy: changedBy,
	}

	if *syncChangeID != "" {
		syncChangeID = &change.ID
	}

	if err := tx.Create(&change).Error; err != nil {
		return fmt.Errorf(
			"create sync change: %w",
			err,
		)
	}

	if setSyncStateToBeLastest {
		state := entity.SyncState{
			SotID:        sotID,
			Entity:       entityName,
			LastChangeID: change.ID,
			UpdatedAt:    now,
		}

		if err := tx.
			Where(
				"sot_id = ? AND entity = ?",
				sotID,
				entityName,
			).
			Assign(state).
			FirstOrCreate(&state).Error; err != nil {
			return fmt.Errorf(
				"update sync state: %w",
				err,
			)
		}
	}

	return nil
}

func (s *Service) GetChangesAfter(
	sotID string,
	entityName string,
	startID string,
	limit int,
) ([]entity.SyncChange, error) {
	if sotID == "" {
		return nil, fmt.Errorf("sot id is required")
	}

	if entityName == "" {
		return nil, fmt.Errorf("sync entity is required")
	}

	if limit <= 0 {
		limit = 500
	}

	var changes []entity.SyncChange

	err := s.db.
		Table("sync_changes").
		Where(
			"sot_id = ? AND entity = ? AND id > ?",
			sotID,
			entityName,
			startID,
		).
		Order("id ASC").
		Limit(limit).
		Find(&changes).
		Error

	if err != nil {
		return nil, fmt.Errorf(
			"get pending sync changes: %w",
			err,
		)
	}

	return changes, nil
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

	var state entity.SyncState

	err := s.db.
		Where(
			"sot_id = ? AND entity = ?",
			sotID,
			entityName,
		).
		First(&state).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.SyncState{}, nil
		}

		return entity.SyncState{}, fmt.Errorf(
			"get sync state: %w",
			err,
		)
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

	var count int64

	err := s.db.
		Model(&entity.SyncChange{}).
		Where(
			"sot_id = ? AND entity = ? AND id > ?",
			sotID,
			entityName,
			startID,
		).
		Count(&count).
		Error

	if err != nil {
		return 0, fmt.Errorf(
			"get sync change count after %q: %w",
			startID,
			err,
		)
	}

	return count, nil
}

func (s *Service) Transaction(
	fn func(tx *gorm.DB) error,
) error {
	if fn == nil {
		return fmt.Errorf("transaction function is required")
	}

	return s.db.Transaction(fn)
}

func (s *Service) InsertSyncChanges(
	sotID string,
	entity string,
	syncChanges []entity.SyncChange,
) error {
	if len(syncChanges) == 0 {
		return nil
	}

	return s.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&syncChanges).Error; err != nil {
			return err
		}

		return s.SetSyncStateToBeLastest(tx, sotID, entity)
	})
}

func (s *Service) SetSyncStateToBeLastest(
	tx *gorm.DB,
	sotID string,
	entityName string,
) error {
	var lastChange entity.SyncChange

	err := tx.
		Where("sot_id = ? AND entity = ?", sotID, entityName).
		Order("id DESC").
		First(&lastChange).
		Error
	if err != nil {
		return fmt.Errorf("get latest sync change: %w", err)
	}

	state := entity.SyncState{
		SotID:        sotID,
		Entity:       entityName,
		LastChangeID: lastChange.ID,
		UpdatedAt:    time.Now(),
	}

	return tx.
		Where("sot_id = ? AND entity = ?", sotID, entityName).
		Assign(map[string]any{
			"last_change_id": state.LastChangeID,
			"updated_at":     state.UpdatedAt,
		}).
		FirstOrCreate(&state).
		Error
}

func MergePendingChanges[T any](
	baseChanges []entity.SyncChange,
	compChanges []entity.SyncChange,
	baseData map[string]T,
	compData map[string]T,
	shouldUpdateData func(mergedData, compData T) bool,
) ([]entity.SyncChange, map[string]T) {
	if shouldUpdateData == nil {
		shouldUpdateData = func(mergedData, compData T) bool {
			return true
		}
	}

	mergedChanges := make(map[string]entity.SyncChange)
	mergedData := make(map[string]T)

	for _, change := range baseChanges {
		entityID := change.EntityID
		existing, exists := mergedChanges[entityID]

		if exists && (existing.Operation == entity.OperationDelete ||
			(existing.Operation == entity.OperationCreate &&
				change.Operation == entity.OperationUpdate)) {
			continue
		}

		if !exists {
			mergedChanges[entityID] = change
			mergedData[entityID] = baseData[entityID]
		} else if change.Operation == entity.OperationDelete {
			mergedChanges[entityID] = change
			delete(mergedData, entityID)
		} else if change.ID > existing.ID {
			mergedChanges[entityID] = change
		}
	}

	for _, change := range compChanges {
		entityID := change.EntityID
		existing, exists := mergedChanges[entityID]

		if !exists {
			continue
		}

		if existing.Operation == entity.OperationDelete {
			continue
		}

		if change.Operation == entity.OperationDelete {
			delete(mergedChanges, entityID)
			delete(mergedData, entityID)
		} else if change.ID > existing.ID &&
			shouldUpdateData(
				mergedData[entityID],
				compData[entityID],
			) {
			mergedData[entityID] = compData[entityID]
		}
	}

	sortedChanges := make([]entity.SyncChange, 0, len(mergedChanges))
	for _, change := range mergedChanges {
		sortedChanges = append(sortedChanges, change)
	}

	sort.Slice(sortedChanges, func(i, j int) bool {
		return sortedChanges[i].ID < sortedChanges[j].ID
	})

	return sortedChanges, mergedData
}
