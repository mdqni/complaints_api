package categories

import (
	"complaint_server/internal/delivery/http/middleware/admin_only"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.With(admin_only.AdminOnlyMiddleware(h.Log, h.Cfg, h.AdminService)).Post("/admin/", h.Create)
	r.With(admin_only.AdminOnlyMiddleware(h.Log, h.Cfg, h.AdminService)).Put("/admin/{id}", h.Update)
	r.With(admin_only.AdminOnlyMiddleware(h.Log, h.Cfg, h.AdminService)).Delete("/admin/{id}", h.Delete)
	r.Get("/", h.GetAll)
	r.Get("/{id}", h.GetById)
	r.Get("/{id}/complaints", h.GetCategoryComplaints)
}
