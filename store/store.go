package store

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gen"
	"gorm.io/gorm"
)

type Config struct {
	Driver string
	DSN    string
}

type Handler struct {
	*gorm.DB
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

	return &Handler{
		DB: db,
	}, err
}

type generateModel struct {
	outputPath string
	f          func(g *gen.Generator)
}

var generateModels []*generateModel

func RegistGenerate(outputPath string, f func(g *gen.Generator)) {
	generateModels = append(generateModels, &generateModel{
		outputPath: outputPath,
		f:          f,
	})
}

func (h *Handler) Generate() {
	for _, gm := range generateModels {
		g := gen.NewGenerator(gen.Config{
			OutPath: gm.outputPath,
			Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		})

		g.UseDB(h.DB)
		gm.f(g)
		g.Execute()
	}
}
