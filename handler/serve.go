package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pandodao/botastic/handler/render"
)

func New(cfg Config) Server {
	return Server{cfg: cfg}
}

type (
	Config struct {
	}

	Server struct {
		cfg Config
	}
)

func (s Server) HandleRest() http.Handler {
	r := chi.NewRouter()
	r.Use(render.WrapResponse(true))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		render.Error(w, http.StatusNotFound, fmt.Errorf("not found"))
	})

	return r
}
