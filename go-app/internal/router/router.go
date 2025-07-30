package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/vmm2136/besu_challenge/go-app/internal/controller"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter configura e retorna um novo router HTTP
func NewRouter(c *controller.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/value", c.GetValueHandler)
	r.Post("/value", c.SetValueHandler)
	r.Post("/sync", c.SyncValueHandler)
	r.Get("/check", c.CheckValueHandler)

	return r
}
