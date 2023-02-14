package hc

import (
	"net/http"
	"time"

	"botastic/handler/render"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Handle(version string) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.NoCache)
	r.Handle("/", handle(version))
	return r
}

func handle(version string) http.HandlerFunc {
	b := time.Now()
	return func(w http.ResponseWriter, r *http.Request) {
		uptime := time.Since(b).Truncate(time.Millisecond)
		render.JSON(w, render.H{
			"uptime":  uptime.String(),
			"version": version,
		})
	}
}
