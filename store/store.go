package store

import (
	"context"
	"embed"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gen"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

const migrationsDir = "migrations"

type handlerKey struct{}

type Config struct {
	Driver string
	DSN    string
}

type Handler struct {
	*gorm.DB
}

func NewContext(ctx context.Context, h *Handler) context.Context {
	return context.WithValue(ctx, handlerKey{}, h)
}

func WithContext(ctx context.Context) *Handler {
	return ctx.Value(handlerKey{}).(*Handler)
}

func MustInit(cfg Config) *Handler {
	h, err := Init(cfg)
	if err != nil {
		panic(err)
	}

	return h
}

func Init(cfg Config) (*Handler, error) {
	var (
		err error
		db  *gorm.DB
	)
	switch cfg.Driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	default:
		panic("unknown driver")
	}
	if err != nil {
		return nil, err
	}

	if err := goose.SetDialect(cfg.Driver); err != nil {
		return nil, err
	}

	return &Handler{
		DB: db,
	}, err
}

type generateModel struct {
	cfg gen.Config
	f   func(g *gen.Generator)
}

var generateModels []*generateModel

func RegistGenerate(cfg gen.Config, f func(g *gen.Generator)) {
	generateModels = append(generateModels, &generateModel{
		cfg: cfg,
		f:   f,
	})
}

func (h *Handler) Generate() {
	for _, gm := range generateModels {
		if gm.cfg.Mode == 0 {
			gm.cfg.Mode = gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutCRUDMethods
		}
		g := gen.NewGenerator(gm.cfg)
		g.UseDB(h.DB)
		gm.f(g)
		g.Execute()
	}
}

func (h *Handler) MigrationUp() error {
	db, _ := h.DB.DB()
	goose.SetBaseFS(embedMigrations)
	return goose.Up(db, migrationsDir)
}

func (h *Handler) MigrationUpTo(version int64) error {
	db, _ := h.DB.DB()
	goose.SetBaseFS(embedMigrations)
	return goose.UpTo(db, migrationsDir, version)
}

func (h *Handler) MigrationDown() error {
	db, _ := h.DB.DB()
	goose.SetBaseFS(embedMigrations)
	return goose.Down(db, migrationsDir)
}

func (h *Handler) MigrationDownTo(version int64) error {
	db, _ := h.DB.DB()
	goose.SetBaseFS(embedMigrations)
	return goose.DownTo(db, migrationsDir, version)
}

func (h *Handler) MigrationRedo() error {
	db, _ := h.DB.DB()
	goose.SetBaseFS(embedMigrations)
	return goose.Redo(db, migrationsDir)
}

func (h *Handler) MigrationCreate(name string) error {
	db, _ := h.DB.DB()
	goose.SetBaseFS(nil)
	return goose.Create(db, "store/"+migrationsDir, name, "sql")
}

func (h *Handler) MigrationStatus() error {
	db, _ := h.DB.DB()
	goose.SetBaseFS(embedMigrations)
	return goose.Status(db, migrationsDir)
}
