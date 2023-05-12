package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/api"
	"github.com/pandodao/botastic/models"
	"gorm.io/gorm"
)

func (h *Handler) CreateTurn(ctx context.Context, turn *models.Turn) error {
	return h.db.WithContext(ctx).Create(turn).Error
}

func (h *Handler) GetTurnCount(ctx context.Context, convId uuid.UUID, status api.TurnStatus) (int64, error) {
	var count int64
	return count, h.db.WithContext(ctx).Model(&models.Turn{}).Where("conv_id = ? AND status = ?", convId, int(status)).Count(&count).Error
}

func (h *Handler) GetTurns(ctx context.Context, convId uuid.UUID, status api.TurnStatus, limit int) ([]*models.Turn, error) {
	var turns []*models.Turn
	if err := h.db.WithContext(ctx).Where("conv_id = ? AND status = ?", convId, int(status)).Limit(limit).Order("created_at").Find(&turns).Error; err != nil {
		return nil, err
	}

	return turns, nil
}

func (h *Handler) GetTurnsByStatus(ctx context.Context, status []api.TurnStatus) ([]*models.Turn, error) {
	var turns []*models.Turn
	ss := []int{}
	for _, s := range status {
		ss = append(ss, int(s))
	}
	if err := h.db.WithContext(ctx).Where("status IN (?)", ss).Order("created_at").Find(&turns).Error; err != nil {
		return nil, err
	}

	return turns, nil
}

func (h *Handler) UpdateTurnToSuccess(ctx context.Context, id uint, response string, promptTokens, completionTokens, totalTokens int, mr models.MiddlewareResults) error {
	return h.db.WithContext(ctx).Model(&models.Turn{}).Where("id = ?", id).Updates(map[string]any{
		"status":             int(api.TurnStatusSuccess),
		"response":           response,
		"prompt_tokens":      promptTokens,
		"completion_tokens":  completionTokens,
		"total_tokens":       totalTokens,
		"middleware_results": mr,
	}).Error
}

func (h *Handler) UpdateTurnToFailed(ctx context.Context, id uint, err *api.TurnError, mr models.MiddlewareResults) error {
	return h.db.WithContext(ctx).Model(&models.Turn{}).Where("id = ?", id).Updates(map[string]any{
		"status":             int(api.TurnStatusFailed),
		"error_code":         err.Code,
		"error_message":      err.Error(),
		"middleware_results": mr,
	}).Error
}

func (h *Handler) UpdateTurnToProcessing(ctx context.Context, id uint) error {
	return h.db.WithContext(ctx).Model(&models.Turn{}).Where("id = ?", id).Update("status", int(api.TurnStatusProcessing)).Error
}

func (h *Handler) GetTurn(ctx context.Context, id uint) (*models.Turn, error) {
	var turn models.Turn
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(&turn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &turn, nil
}
