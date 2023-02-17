package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/pandodao/botastic/core"
	util "github.com/pandodao/botastic/util"
)

func New(
	cfg Config,
	apps core.AppStore,
) *service {
	return &service{
		cfg:  cfg,
		apps: apps,
	}
}

type Config struct {
	SecretKey string
}

type service struct {
	cfg  Config
	apps core.AppStore
}

func (s *service) CreateApp(ctx context.Context) (*core.App, error) {
	var err error
	app := &core.App{}
	app.AppID = uuid.New().String()
	app.AppSecret, err = util.GenerateSecret()
	if err != nil {
		return nil, err
	}

	if err := app.Encrypt(s.cfg.SecretKey); err != nil {
		return nil, err
	}

	id, err := s.apps.CreateApp(ctx, app.AppID, app.AppSecretEncrypted)
	if err != nil {
		return nil, err
	}

	app.ID = id

	return app, nil
}

func (s *service) GetAppByAppID(ctx context.Context, appID string) (*core.App, error) {
	app, err := s.apps.GetAppByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}

	if err := app.Decrypt(s.cfg.SecretKey); err != nil {
		return nil, err
	}

	return app, nil
}
