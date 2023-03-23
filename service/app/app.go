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

func (s *service) ReplaceStore(apps core.AppStore) core.AppService {
	return New(s.cfg, apps)
}

func (s *service) CreateApp(ctx context.Context, userID uint64, name string) (*core.App, error) {
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

	id, err := s.apps.CreateApp(ctx, app.AppID, app.AppSecretEncrypted, userID, name)
	if err != nil {
		return nil, err
	}

	app, err = s.apps.GetApp(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := app.Decrypt(s.cfg.SecretKey); err != nil {
		return nil, err
	}

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

func (s *service) GetAppsByUser(ctx context.Context, userID uint64) ([]*core.App, error) {
	apps, err := s.apps.GetAppsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if err := app.Decrypt(s.cfg.SecretKey); err != nil {
			return nil, err
		}
	}

	return apps, nil
}

func (s *service) UpdateApp(ctx context.Context, id uint64, name string) error {
	if err := s.apps.UpdateAppName(ctx, id, name); err != nil {
		return err
	}
	return nil
}

func (s *service) DeleteApp(ctx context.Context, id uint64) error {
	if err := s.apps.DeleteApp(ctx, id); err != nil {
		return err
	}
	return nil
}
