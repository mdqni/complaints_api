package complaints

import (
	"complaint_server/internal/delivery/http/middleware/admin"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/", h.GetAll)

	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByComplaintId)
	r.Get("/can-submit", h.CanSubmit)
	r.Get("/by-token", h.GetComplaintsByToken)
	r.Delete("/{id}", h.DeleteByOwner)
	r.With(admin.AdminOnlyMiddleware(h.Log, h.Cfg, h.AdminService)).Post("/admin/", h.Create)
	r.With(admin.AdminOnlyMiddleware(h.Log, h.Cfg, h.AdminService)).Put("/admin/{id}", h.Update)
	r.With(admin.AdminOnlyMiddleware(h.Log, h.Cfg, h.AdminService)).Delete("/admin/{id}", h.DeleteByAdmin)

}
