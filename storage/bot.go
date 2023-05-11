package storage

import (
	"context"
	"errors"

	"github.com/pandodao/botastic/models"
	"gorm.io/gorm"
)

func (h *Handler) CreateBot(ctx context.Context, bot *models.Bot) error {
	return h.db.WithContext(ctx).Create(bot).Error
}

func (h *Handler) UpdateBot(ctx context.Context, id uint, m map[string]any) (int64, error) {
	r := h.db.Model(&models.Bot{}).Where("id = ?", id).Updates(m)
	return r.RowsAffected, r.Error
}

func (h *Handler) DeleteBot(ctx context.Context, id uint) error {
	return h.db.WithContext(ctx).Delete(&models.Bot{}, id).Error
}

func (h *Handler) GetBot(ctx context.Context, id uint) (*models.Bot, error) {
	bot := &models.Bot{}
	if err := h.db.WithContext(ctx).Where("id = ?", id).First(bot).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return bot, nil
}

func (h *Handler) GetBots(ctx context.Context) ([]*models.Bot, error) {
	var bots []*models.Bot
	if err := h.db.WithContext(ctx).Find(&bots).Error; err != nil {
		return nil, err
	}

	return bots, nil
}
