package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/models"
)

func (h *Handler) CreateTurn(ctx context.Context, turn *models.Turn) error {
	return h.db.WithContext(ctx).Create(turn).Error
}

func (h *Handler) GetTurnCount(ctx context.Context, convId uuid.UUID, status api.TurnStatus) (int64, error) {
	var count int64
	return count, h.db.WithContext(ctx).Model(&models.Turn{}).Where("conv_id = ? AND status = ?", convId, status).Count(&count).Error
}
