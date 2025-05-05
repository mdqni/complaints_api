package categories

import (
	"complaint_server/internal/config"
	"complaint_server/internal/domain"
	serviceAdmin "complaint_server/internal/service/admin"
	serviceCategory "complaint_server/internal/service/category"
	serviceComplaint "complaint_server/internal/service/complaint"
	"complaint_server/internal/shared/api/response"
	"complaint_server/internal/shared/logger/sl"
	"complaint_server/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
)

type Handler struct {
	Log              *slog.Logger
	AdminService     *serviceAdmin.AdminService
	CategoryService  *serviceCategory.CategoryService
	ComplaintService *serviceComplaint.ComplaintService
	Redis            *redis.Client
	Cfg              *config.Config
}

func NewHandler(ctx context.Context, complaintsService *serviceComplaint.ComplaintService, adminService *serviceAdmin.AdminService, categoryService *serviceCategory.CategoryService, log *slog.Logger, redis *redis.Client, cfg *config.Config) *Handler {
	return &Handler{
		AdminService:     adminService,
		CategoryService:  categoryService,
		ComplaintService: complaintsService,
		Log:              log,
		Redis:            redis,
		Cfg:              cfg,
	}
}

// New @Summary Создать категорию
// @Description Создает новую категорию жалоб с необходимыми данными: название, описание и ответ.
// @Tags Categories
// @Accept json
// @Produce json
// @Param request body Request true "Данные категории"
// @Success 200 {object} response.Response "Категория успешно создана"
// @Failure 400 {object} response.Response "Ошибка валидации или декодирования данных"
// @Failure 500 {object} response.Response "Ошибка сервера"
// @Router /admin/categories [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.categories.create.New"
	log := h.Log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req struct {
		Title       string `json:"title" validate:"required"`
		Description string `json:"description" validate:"required"`
		Answer      string `json:"answer" validate:"required"`
	}

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "Failed to decode request body",
			StatusCode: http.StatusBadRequest,
			Data:       nil,
		})
		return
	}
	log.Info("request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			log.Error("validation failed", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Validation failed",
				StatusCode: http.StatusBadRequest,
				Data:       validationErrors,
			})
			return
		}
		log.Error("unknown validation error", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "Unknown validation error",
			StatusCode: http.StatusBadRequest,
			Data:       nil,
		})
		return
	}

	categoryID, err := h.CategoryService.CreateCategory(r.Context(), domain.Category{
		Title:       req.Title,
		Description: req.Description,
		Answer:      req.Answer,
	})
	if errors.Is(err, storage.ErrDBConnection) {
		log.Error("db connection error", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    storage.ErrDBConnection.Error(),
			StatusCode: http.StatusInternalServerError,
			Data:       nil,
		})
	}
	if errors.Is(err, storage.ErrCreateCategory) {
		log.Error("failed to create categories", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "Error ",
			StatusCode: http.StatusConflict,
			Data:       nil,
		})
	}
	if err != nil {
		log.Error("failed to save categories", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "Failed to save categories",
			StatusCode: http.StatusInternalServerError,
			Data:       nil,
		})
		return
	}

	log.Info("categories saved", slog.Any("id", categoryID))

	render.JSON(w, r, response.Response{
		Message:    "Category created successfully",
		StatusCode: http.StatusOK,
		Data:       categoryID,
	})
}

// GetAll New @Summary Получить все категории
// @Description Возвращает список всех категорий жалоб, доступных в системе.
// @Tags Categories
// @Accept json
// @Produce json
// @Success 200 {array} domain.Category "List of all categories"
// @Failure 500 {object} response.Response "Internal server error while fetching categories"
// @Router /categories [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // Убедись, что установил utf-8
	const op = "handlers.categories.get_all.New"
	log := h.Log.With(
		slog.String("op", op),
		slog.String("url", r.URL.String()))
	ctx := r.Context()

	// Получаем категории
	result, err := h.CategoryService.GetCategories(ctx)

	if err != nil {
		log.Error(op, sl.Err(err))
		w.Header().Set("Content-Type", "application/json; charset=utf-8") // Устанавливаем правильный контент
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, response.Response{
			Message:    "internal error",
			StatusCode: http.StatusInternalServerError,
			Data:       nil,
		})
		return
	}
	log.Info("Categories found")

	// Ручная сериализация JSON с ensure_ascii=false
	responseData := response.Response{
		Message:    "Categories fetched successfully",
		StatusCode: http.StatusOK,
		Data:       result,
	}

	encodedResponse, err := json.MarshalIndent(responseData, "", "  ")
	if err != nil {
		log.Error(op, sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, response.Response{
			Message:    "Failed to serialize response",
			StatusCode: http.StatusInternalServerError,
			Data:       nil,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(encodedResponse)
}

// GetById New @Summary Получить категорию по ID
// @Description Возвращает категорию по уникальному идентификатору (ID).
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID (unique identifier of the categories)"
// @Success 200 {object} domain.Category "Category details"
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 500 {object} response.Response "Internal server error while fetching the categories"
// @Router /categories/{id} [get]
func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.categories.get_by_id.New"

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	log := h.Log.With(
		slog.String("op", op),
		slog.String("url", r.URL.String()),
	)
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Error("Missing complaint_id")
		render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "Missing complaint_id"})
		return
	}
	_uuid, err := uuid.Parse(id)
	result, err := h.CategoryService.GetCategoryById(ctx, _uuid)
	if errors.Is(err, storage.ErrCategoryNotFound) {
		log.Error(op, sl.Err(err))
		responseData, _ := json.Marshal(response.Response{
			Message:    fmt.Sprintf("categories with %d not found", _uuid),
			StatusCode: http.StatusNotFound,
			Data:       nil,
		})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responseData)
		return
	}
	if err != nil {
		log.Error(op, sl.Err(err))
		responseData, _ := json.Marshal(response.Response{
			Message:    "internal error",
			StatusCode: http.StatusInternalServerError,
			Data:       nil,
		})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responseData)
		return
	}

	log.Info("Category found", slog.Any("category_id", _uuid))
	responseData, _ := json.Marshal(response.Response{
		Message:    "Category fetched successfully",
		StatusCode: http.StatusOK,
		Data:       result,
	})
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

// GetCategoryComplaints New GetComplaintsByCategoryId godoc
// @Summary Get complaints by categories UUID
// @Description Retrieve all complaints that belong to a specific categories based on its unique identifier (Category ID).
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path string true "Category UUID (unique identifier of the categories)"
// @Success 200 {array} domain.Complaint "List of complaints associated with the given categories"
// @Failure 400 {object} response.Response "Invalid categories ID format"
// @Failure 404 {object} response.Response "No complaints found for the given categories"
// @Failure 500 {object} response.Response "Internal server error while fetching complaints"
// @Router /categories/{id}/complaints [get]
func (h *Handler) GetCategoryComplaints(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.complaints.getByCategoryId.New"

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	log := h.Log.With(
		slog.String("op", op),
		slog.String("url", r.URL.String()),
	)

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Error("Missing categories id")
		render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "Missing complaint_id"})
		return
	}

	categoryUUID, err := uuid.Parse(id)
	if err != nil {
		log.Error(op, sl.Err(err))
		render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: err.Error()})
		return
	}

	result, err := h.ComplaintService.GetComplaintsByCategoryId(r.Context(), categoryUUID)
	if err != nil {
		log.Error("failed to get complaints", sl.Err(err))
		responseData, _ := json.Marshal(response.Response{
			Message:    "no complaints found for the given categories",
			StatusCode: http.StatusNotFound,
			Data:       nil,
		})
		w.WriteHeader(http.StatusNotFound)
		w.Write(responseData)
		return
	}

	log.Info("Complaints found for categories", slog.Any("category_id", categoryUUID))
	responseData, _ := json.Marshal(response.Response{
		Message:    "Complaints fetched successfully",
		StatusCode: http.StatusOK,
		Data:       result,
	})
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)

}

// Update New @Summary Обновить категорию
// @Description Обновляет информацию о категории жалоб. Требуется предоставить ID категории и новые данные (название, описание и ответ).
// @Tags Categories
// @Accept json
// @Produce json
// @Param request body Request true "Данные категории"
// @Success 200 {object} response.Response "Категория успешно обновлена"
// @Failure 400 {object} response.Response "Ошибка валидации или декодирования данных"
// @Failure 500 {object} response.Response "Ошибка сервера"
// @Router /admin/categories/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.categories.update.New"
	log := h.Log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	var req struct {
		Id          uuid.UUID `json:"id"`
		Title       string    `json:"title" validate:"required"`
		Description string    `json:"description" validate:"required"`
		Answer      string    `json:"answer" validate:"required"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "Failed to decode request body",
			StatusCode: http.StatusBadRequest,
			Data:       nil,
		})
		return
	}
	log.Info("request body decoded", slog.Any("request", req))

	if err := validator.New().Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			log.Error("validation failed", sl.Err(err))
			render.JSON(w, r, response.Response{
				Message:    "Validation failed",
				StatusCode: http.StatusBadRequest,
				Data:       validationErrors,
			})
			return
		}
		log.Error("unknown validation error", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "Unknown validation error",
			StatusCode: http.StatusBadRequest,
			Data:       nil,
		})
		return
	}
	categoryId := chi.URLParam(r, "id")
	if categoryId == "" {
		log.Error("Missing categories id")
		render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "Missing categories id"})
		return
	}
	id, err := uuid.Parse(categoryId)
	_, err = h.CategoryService.UpdateCategory(r.Context(), id, domain.Category{
		ID:          req.Id,
		Title:       req.Title,
		Description: req.Description,
		Answer:      req.Answer,
	})
	if errors.Is(err, storage.ErrCategoryNotFound) {
		log.Error(op, sl.Err(err))
		responseData, _ := json.Marshal(response.Response{
			Message:    fmt.Sprintf("categories with %d not found", id),
			StatusCode: http.StatusNotFound,
			Data:       nil,
		})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responseData)
		return
	}
	if err != nil {
		log.Error("failed to update categories", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "Failed to update categories",
			StatusCode: http.StatusInternalServerError,
			Data:       nil,
		})
		return
	}

	render.JSON(w, r, response.Response{
		Message:    "Category updated successfully",
		StatusCode: http.StatusOK,
		Data:       map[string]interface{}{"id": req.Id},
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.complaint.deleteByAdmin.New"

	log := h.Log.With(
		slog.String("op", op),
		slog.String("url", r.URL.String()))

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Info("id can not be empty")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "id can not be empty"})
		return
	}
	uuid_, err := uuid.Parse(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, response.Response{StatusCode: http.StatusInternalServerError, Message: err.Error()})
		return
	}

	err = h.CategoryService.DeleteCategoryById(r.Context(), uuid_)

	if errors.Is(err, storage.ErrCategoryNotFound) {
		log.Error("categories not found", sl.Err(err))
		w.WriteHeader(http.StatusNotFound)
		render.JSON(w, r, response.Response{StatusCode: http.StatusNotFound, Message: err.Error()})
		return
	}
	if errors.Is(err, storage.ErrHasRelatedRows) {
		log.Error("there are related rows", sl.Err(err))
		w.WriteHeader(http.StatusConflict)
		render.JSON(w, r, response.Response{StatusCode: http.StatusConflict, Message: err.Error()})
		return
	}
	if err != nil {
		log.Info("failed to deleteByAdmin categories", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, response.Response{StatusCode: http.StatusInternalServerError, Message: err.Error(), Data: nil})
		return
	}
	log.Info("categories deleted")
	render.JSON(w, r, response.Response{StatusCode: http.StatusOK})
}
