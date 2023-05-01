package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/api"
	"gorm.io/gorm"
)

type Conv struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	BotID     uint      `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (c Conv) API() api.Conv {
	return api.Conv{
		ID:        c.ID,
		BotID:     c.BotID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
