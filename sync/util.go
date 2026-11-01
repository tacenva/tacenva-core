package sync

import (
	"sort"

	"github.com/tacenva/tacenva-core/entity"
)

func MergePendingChanges[T any](
	baseChanges []entity.SyncChange,
	compChanges []entity.SyncChange,
	baseData map[string]T,
	compData map[string]T,
	shouldUpdateData func(mergedData, compData T) bool,
) ([]entity.SyncChange, map[string]T) {
	if shouldUpdateData == nil {
		shouldUpdateData = func(_, _ T) bool {
			return true
		}
	}

	mergedChanges := make(map[string]entity.SyncChange)
	mergedData := make(map[string]T)

	for _, change := range baseChanges {
		entityID := change.EntityID
		existing, exists := mergedChanges[entityID]

		if exists &&
			(existing.Operation == entity.OperationDelete ||
				(existing.Operation == entity.OperationCreate &&
					change.Operation == entity.OperationUpdate)) {
			continue
		}

		if !exists {
			mergedChanges[entityID] = change
			mergedData[entityID] = baseData[entityID]
			continue
		}

		if change.Operation == entity.OperationDelete {
			mergedChanges[entityID] = change
			delete(mergedData, entityID)
			continue
		}

		if change.ID > existing.ID {
			mergedChanges[entityID] = change
		}
	}

	for _, change := range compChanges {
		entityID := change.EntityID
		existing, exists := mergedChanges[entityID]

		if !exists || existing.Operation == entity.OperationDelete {
			continue
		}

		if change.Operation == entity.OperationDelete {
			delete(mergedChanges, entityID)
			delete(mergedData, entityID)
			continue
		}

		if change.ID > existing.ID &&
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
