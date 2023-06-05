package storage

import (
	"context"
	"errors"
	"sort"

	"github.com/pandodao/botastic/internal/utils"
	"github.com/pandodao/botastic/models"
	"gorm.io/gorm"
)

func (h *Handler) UpsertIndexes(ctx context.Context, indexes []*models.Index) ([]uint, error) {
	newCreatedIndexes := []uint{}
	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, index := range indexes {
			if index.ID == 0 {
				if err := tx.Create(index).Error; err != nil {
					return err
				}
				newCreatedIndexes = append(newCreatedIndexes, index.ID)
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

		return nil
	}); err != nil {
		return nil, err
	}

	return newCreatedIndexes, nil
}

func (h *Handler) DeleteIndexes(ctx context.Context, ids []uint) error {
	return h.db.WithContext(ctx).Where("id IN (?)", ids).Delete(&models.Index{}).Error
}

func (h *Handler) GetIndexesByGroupKey(ctx context.Context, groupKey string) ([]*models.Index, error) {
	var indexes []*models.Index
	if err := h.db.WithContext(ctx).Where("group_key = ?", groupKey).Find(&indexes).Error; err != nil {
		return nil, err
	}
	return indexes, nil
}

func (h *Handler) SearchIndexes(ctx context.Context, groupKey string, data []float32, limit int) ([]*models.Index, error) {
	indexes, err := h.GetIndexesByGroupKey(ctx, groupKey)
	if err != nil {
		return nil, err
	}
	scoreMap := make(map[float64]*models.Index, len(indexes))
	scores := make([]float64, 0, len(indexes))
	for _, index := range indexes {
		score := utils.CosineSimilarity(data, index.Vector)
		scores = append(scores, score)
		index.Score = score
		scoreMap[score] = index
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i] > scores[j]
	})

	var result []*models.Index
	if limit > len(scores) {
		limit = len(scores)
	}
	for _, score := range scores[:limit] {
		index := scoreMap[score]
		result = append(result, index)
	}
	return result, nil
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
