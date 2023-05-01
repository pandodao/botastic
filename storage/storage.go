package storage

import (
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/models"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Handler struct {
	cfg config.DBConfig
	db  *gorm.DB
	m   *gormigrate.Gormigrate
}

func Init(cfg config.DBConfig) (*Handler, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case config.DBSqlite:
		dialector = sqlite.Open(cfg.DSN)
	case config.DBMysql:
		dialector = mysql.Open(cfg.DSN)
	case config.DBPostgres:
		dialector = postgres.Open(cfg.DSN)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if cfg.Debug {
		db = db.Debug()
	}

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{})
	m.InitSchema(func(tx *gorm.DB) error {
		err := tx.AutoMigrate(&models.Conv{}, &models.Turn{}, &models.Bot{})
		if err != nil {
			return err
		}

		return nil
	})

	return &Handler{
		cfg: cfg,
		db:  db,
		m:   m,
	}, nil
}

func (h *Handler) Migrate() error {
	return h.m.Migrate()
}
