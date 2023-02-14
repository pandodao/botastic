package handler

import (
	"fmt"
	"net/http"

	"botastic/core"
	"botastic/handler/asset"
	"botastic/handler/echo"
	"botastic/handler/render"

	"github.com/go-chi/chi"
)

func New(cfg Config, assets core.AssetStore) Server {
	return Server{cfg: cfg, assets: assets}
}

type (
	Config struct {
	}

	Server struct {
		cfg    Config
		assets core.AssetStore
	}
)

func (s Server) HandleRest() http.Handler {
	r := chi.NewRouter()
	r.Use(render.WrapResponse(true))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, http.StatusNotFound, fmt.Errorf("not found"))
	})

	r.Route("/echo", func(r chi.Router) {
		r.Get("/{msg}", echo.Get())
		r.Post("/", echo.Post())
	})

	r.Route("/assets", func(r chi.Router) {
		r.Get("/", asset.GetAssets(s.assets))
		r.Get("/{assetID}", asset.GetAsset(s.assets))
	})

	return r
}
