package sync

import (
	"fmt"
	"time"

	"github.com/tacenva/tacenva-core/entity"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Transaction(
	fn func(tx *gorm.DB) error,
) error {
	if r.db == nil {
		return fmt.Errorf("database is required")
	}
	if fn == nil {
		return fmt.Errorf("transaction function is required")
	}

	return r.db.Transaction(fn)
}

func (r *Repository) SyncChangeExists(
	tx *gorm.DB,
	id string,
	sotID string,
) (bool, error) {
	var count int64

	err := tx.Model(&entity.SyncChange{}).
		Where("id = ? AND sot_id = ?", id, sotID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("query sync change: %w", err)
	}

	return count > 0, nil
}

func (r *Repository) CreateSyncChange(
	tx *gorm.DB,
	change *entity.SyncChange,
) error {
	if err := tx.Create(change).Error; err != nil {
		return fmt.Errorf("create sync change: %w", err)
	}

	return nil
}

func (r *Repository) UpsertSyncState(
	tx *gorm.DB,
	state *entity.SyncState,
) error {
	return tx.
		Where(
			"sot_id = ? AND entity = ?",
			state.SotID,
			state.Entity,
		).
		Assign(map[string]any{
			"last_change_id": state.LastChangeID,
			"updated_at":     state.UpdatedAt,
		}).
		FirstOrCreate(state).
		Error
}

func (r *Repository) GetChangesAfter(
	sotID string,
	entityName string,
	startID string,
	limit int,
) ([]entity.SyncChange, error) {
	var changes []entity.SyncChange

	err := r.db.
		Model(&entity.SyncChange{}).
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
		return nil, fmt.Errorf("get pending sync changes: %w", err)
	}

	return changes, nil
}

func (r *Repository) GetSyncState(
	sotID string,
	entityName string,
) (entity.SyncState, error) {
	var state entity.SyncState

	err := r.db.
		Where(
			"sot_id = ? AND entity = ?",
			sotID,
			entityName,
		).
		First(&state).
		Error
	if err != nil {
		return entity.SyncState{}, err
	}

	return state, nil
}

func (r *Repository) GetChangeCount(
	sotID string,
	entityName string,
	startID string,
) (int64, error) {
	var count int64

	err := r.db.
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

func (r *Repository) CreateSyncChanges(
	tx *gorm.DB,
	changes []entity.SyncChange,
) error {
	if len(changes) == 0 {
		return nil
	}

	if err := tx.Create(&changes).Error; err != nil {
		return fmt.Errorf("create sync changes: %w", err)
	}

	return nil
}

func (r *Repository) GetLatestSyncChange(
	tx *gorm.DB,
	sotID string,
	entityName string,
) (entity.SyncChange, error) {
	var change entity.SyncChange

	err := tx.
		Where("sot_id = ? AND entity = ?", sotID, entityName).
		Order("id DESC").
		First(&change).
		Error
	if err != nil {
		return entity.SyncChange{}, fmt.Errorf(
			"get latest sync change: %w",
			err,
		)
	}

	return change, nil
}

func (r *Repository) NewSyncState(
	sotID string,
	entityName string,
	lastChangeID string,
) entity.SyncState {
	return entity.SyncState{
		SotID:        sotID,
		Entity:       entityName,
		LastChangeID: lastChangeID,
		UpdatedAt:    time.Now(),
	}
}

func (r *Repository) GetLatestSyncChangeByState(
	state *entity.SyncState,
) (*entity.SyncChange, error) {
	if state == nil {
		return nil, fmt.Errorf("sync state is required")
	}

	if state.LastChangeID == "" {
		return nil, nil
	}

	var change entity.SyncChange

	err := r.db.
		Where(
			"id = ? AND sot_id = ? AND entity = ?",
			state.LastChangeID,
			state.SotID,
			state.Entity,
		).
		First(&change).
		Error

	if err != nil {
		return nil, err
	}

	return &change, nil
}
