package complaints

import (
	"complaint_server/internal/delivery/http/middleware/admin_only"
	"complaint_server/internal/delivery/http/middleware/cache"

	"github.com/go-chi/chi/v5"
	"time"
)

const CacheDuration = 5 * time.Minute

func RegisterRoutes(r chi.Router, h *Handler) {
	r.With(cache.CacheMiddleware(h.Redis, CacheDuration, h.Log)).
		Get("/", h.GetAll)

	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByComplaintId)
	r.Get("/can-submit", h.CanSubmit)
	r.Get("/by-token", h.GetComplaintsByToken)
	r.Delete("/{id}", h.DeleteByOwner)
	r.With(admin_only.AdminOnlyMiddleware(h.Log, h.Cfg, h.AdminService)).Post("/admin/", h.Create)
	r.With(admin_only.AdminOnlyMiddleware(h.Log, h.Cfg, h.AdminService)).Put("/admin/{id}", h.Update)
	r.With(admin_only.AdminOnlyMiddleware(h.Log, h.Cfg, h.AdminService)).Delete("/admin/{id}", h.DeleteByAdmin)

}
