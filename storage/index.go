package storage

import (
	"context"
	"errors"

	"github.com/pandodao/botastic/models"
	"gorm.io/gorm"
)

func (h *Handler) UpsertIndexes(ctx context.Context, indexes []*models.Index, callback func() error) error {
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, index := range indexes {
			if index.ID == 0 {
				if err := tx.Create(index).Error; err != nil {
					return err
				}
			} else {
				update := map[string]interface{}{
					"properties": index.Properties,
				}
				if len(index.Vector) > 0 {
					update["vector"] = index.Vector
				}
				if err := tx.Model(index).UpdateColumns(update).Error; err != nil {
					return err
				}
			}
		}

		return callback()
	})
}

func (h *Handler) GetIndexesByGroupKey(ctx context.Context, groupKey string) ([]*models.Index, error) {
	var indexes []*models.Index
	if err := h.db.WithContext(ctx).Where("group_key = ?", groupKey).Find(&indexes).Error; err != nil {
		return nil, err
	}
	return indexes, nil
}

func (h *Handler) GetIndexes(ctx context.Context, ids []uint) ([]*models.Index, error) {
	var indexes []*models.Index
	if err := h.db.WithContext(ctx).Where("id IN (?)", ids).Find(&indexes).Error; err != nil {
		return nil, err
	}
	return indexes, nil
}

func (h *Handler) GetIndex(ctx context.Context, id uint) (*models.Index, error) {
	index := &models.Index{}
	if err := h.db.WithContext(ctx).First(index, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return index, nil
}
