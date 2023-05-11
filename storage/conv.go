package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/models"
	"gorm.io/gorm"
)

func (h *Handler) CreateConv(ctx context.Context, conv *models.Conv) error {
	if conv.ID == uuid.Nil {
		conv.ID = uuid.New()
	}

	return h.db.WithContext(ctx).Create(conv).Error
}

func (h *Handler) UpdateConv(ctx context.Context, id uuid.UUID, m map[string]any) (int64, error) {
	r := h.db.Model(&models.Conv{}).Where("id = ?", id).Updates(m)
	return r.RowsAffected, r.Error
}

func (h *Handler) GetConv(ctx context.Context, id uuid.UUID) (*models.Conv, error) {
	conv := &models.Conv{}
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(conv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return conv, nil
}

func (h *Handler) DeleteConv(ctx context.Context, id uuid.UUID) error {
	return h.db.WithContext(ctx).Delete(&models.Conv{}, id).Error
}
