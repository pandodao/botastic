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

func (h *Handler) GetTurns(ctx context.Context, convId uuid.UUID, status api.TurnStatus, limit int) ([]*models.Turn, error) {
	var turns []*models.Turn
	if err := h.db.WithContext(ctx).Where("conv_id = ? AND status = ?", convId, status).Limit(limit).Order("created_at DESC").Find(&turns).Error; err != nil {
		return nil, err
	}

	return turns, nil
}

func (h *Handler) UpdateTurnToSuccess(ctx context.Context, id uint, response string, promptTokens, completionTokens, totalTokens int, mr models.MiddlewareResults) error {
	return h.db.WithContext(ctx).Model(&models.Turn{}).Where("id = ?", id).Updates(map[string]any{
		"status":             api.TurnStatusSuccess,
		"response":           response,
		"prompt_tokens":      promptTokens,
		"completion_tokens":  completionTokens,
		"total_tokens":       totalTokens,
		"middleware_results": mr,
	}).Error
}

func (h *Handler) UpdateTurnToFailed(ctx context.Context, id uint, err *api.TurnError, mr models.MiddlewareResults) error {
	return h.db.WithContext(ctx).Model(&models.Turn{}).Where("id = ?", id).Updates(map[string]any{
		"status":             api.TurnStatusFailed,
		"error_code":         err.Code,
		"error_message":      err.Error(),
		"middleware_results": mr,
	}).Error
}
