package complaints

import (
	"complaint_server/internal/config"
	"complaint_server/internal/domain"
	serviceAdmin "complaint_server/internal/service/admin"
	serviceCategory "complaint_server/internal/service/category"
	serviceComplaint "complaint_server/internal/service/complaint"
	"complaint_server/internal/shared/api/response"
	jwt2 "complaint_server/internal/shared/jwt"
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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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
	}
}

// GetAll New GetComplaintById godoc
// @Summary Get a complaint by ID
// @Description Retrieve a complaint using its unique identifier. The ID must be an integer that corresponds to a valid complaint in the database.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path int true "Complaint ID (unique identifier of the complaint)"
// @Success 200 {object} domain.Complaint "Complaint details"
// @Failure 400 {object} response.Response "Invalid request, incorrect ID format"
// @Failure 404 {object} response.Response "Complaint with the given ID not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints [get]
func (h Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.complaint.getAllComplaints.New"

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	log := h.Log.With(
		slog.String("op", op),
		slog.String("url", r.URL.String()),
	)

	result, err := h.ComplaintService.GetAllComplaints(r.Context())
	if errors.Is(err, storage.ErrComplaintNotFound) {
		log.Error("complaint not found", sl.Err(err))
		responseData, _ := json.Marshal(response.Response{
			StatusCode: http.StatusOK,
			Data:       []domain.Complaint{},
			Message:    storage.ErrComplaintNotFound.Error(),
		})
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
		return
	}
	if err != nil {
		log.Error(op, sl.Err(err))
		responseData, _ := json.Marshal(response.Response{
			StatusCode: http.StatusInternalServerError,
			Data:       nil,
			Message:    "internal error",
		})
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(responseData)
		return
	}
	w.WriteHeader(http.StatusOK)
	responseData, _ := json.Marshal(response.Response{
		StatusCode: http.StatusOK,
		Data:       result,
	})
	w.Write(responseData)
}

// Create New CreateComplaint godoc
// @Summary Create a new complaint
// @Description Create a new complaint for a specific user and categories. Only one complaint can be submitted per hour for the same user.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param Request body Request true "Complaint details" // Request body with message, category_id, and barcode
// @Success 200 {object} response.Response "Success response with complaint ID and answer"
// @Failure 400 {object} response.Response "Invalid request, bad input or validation error"
// @Failure 429 {object} response.Response "Limit of one complaint per hour exceeded"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints [post]
func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.complaints.register.New"
	log := h.Log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req := struct {
		Message    string    `json:"message" validate:"required"`
		CategoryID uuid.UUID `json:"category_id" validate:"required"`
		Barcode    int       `json:"barcode"`
	}{}
	err := render.DecodeJSON(r.Body, &req)
	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "failed to decode request",
			StatusCode: http.StatusBadRequest,
			Data:       nil,
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("request body decoded", slog.Any("request", req))
	if err := validator.New().Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		errors.As(err, &validationErrors)
		log.Error("failed to validate request", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "validation error",
			StatusCode: http.StatusBadRequest,
			Data:       validationErrors,
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	message := req.Message
	categoryID := req.CategoryID
	barcode := req.Barcode

	complaintID, answer, err := h.ComplaintService.CreateComplaint(r.Context(), barcode, categoryID, message)

	if errors.Is(err, storage.ErrLimitOneComplaintInOneHour) {
		log.Error("failed to register complaint due to rate limit", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "You can only submit one complaint per hour. Please try again later.",
			StatusCode: http.StatusTooManyRequests,
			Data:       nil,
		})
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	if err != nil {
		log.Error("failed to register complaint", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "failed to save complaint",
			StatusCode: http.StatusInternalServerError,
			Data:       nil,
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	render.Status(r, http.StatusOK)
	render.JSON(w, r, response.Response{
		Message:    answer,
		StatusCode: http.StatusOK,
		Data:       complaintID,
	})
}

// GetByComplaintId New GetComplaintById godoc
// @Summary Get a complaint by ID
// @Description Retrieve a complaint using its unique identifier. The UUID must be a string that corresponds to a valid complaint in the database.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path string true "Complaint ID (unique identifier of the complaint)"
// @Success 200 {object} domain.Complaint "Complaint details"
// @Failure 400 {object} response.Response "Invalid request, incorrect ID format"
// @Failure 404 {object} response.Response "Complaint with the given ID not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints/{id} [get]
func (h Handler) GetByComplaintId(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.complaint.get_by_complaint_id.New"

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	log := h.Log.With(
		slog.String("op", op),
		slog.String("url", r.URL.String()))

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Error("Missing complaint_id")
		render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "Missing complaint_id"})
		return
	}
	complaintUUID, err := uuid.Parse(id)
	if err != nil {
		log.Error("Invalid complaint_id")
		render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: err.Error()})
	}
	result, err := h.ComplaintService.GetComplaintByUUID(r.Context(), complaintUUID)
	if errors.Is(err, storage.ErrComplaintNotFound) {
		log.Error("complaint not found", sl.Err(err))
		render.JSON(w, r, response.Error("complaint with this id not found", http.StatusNotFound))
		return
	}
	if errors.Is(err, storage.ErrScanFailure) {
		log.Error("complaint scan failure", sl.Err(err))
		render.JSON(w, r, response.Response{StatusCode: http.StatusInternalServerError, Message: err.Error()})
		return
	}
	if err != nil {
		log.Error(op, sl.Err(err))
		render.JSON(w, r, response.Error("internal error", http.StatusInternalServerError))
		return
	}
	responseData, _ := json.Marshal(response.Response{
		StatusCode: http.StatusOK,
		Data:       result,
	})
	w.Write(responseData)
}

// CanSubmit New godoc
// @Summary      Проверка возможности отправки жалобы
// @Description  Проверяет, может ли пользователь отправить новую жалобу (прошел ли час с последнего запроса)
// @Tags         Complaints
// @Accept       json
// @Produce      json
// @Param        barcode  query     int  true  "Barcode пользователя"
// @Success      200    {object}  response.Response  "true/false"
// @Failure      400    {object}  response.Response  "missing barcode"
// @Failure      500    {object}  response.Response  "internal error"
// @Router       /complaints/can-submit [get]
func (h Handler) CanSubmit(w http.ResponseWriter, r *http.Request) {
	barcode := r.URL.Query().Get("barcode")
	log := h.Log
	if barcode == "" {
		log.Error("missing barcode", http.StatusBadRequest)
		http.Error(w, "missing barcode", http.StatusBadRequest)
		return
	}
	bcode, err := strconv.Atoi(barcode)
	if err != nil {
		log.Error("bad barcode", err, http.StatusBadRequest)
		render.JSON(w, r, &response.Response{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		})
	}
	canSubmit, err := h.ComplaintService.CanSubmitByBarcode(r.Context(), bcode)
	if errors.Is(err, storage.ErrLimitOneComplaintInOneHour) {
		log.Error("Can submit error: ", storage.ErrLimitOneComplaintInOneHour.Error())
		responseData, _ := json.Marshal(response.Response{
			StatusCode: http.StatusTooManyRequests,
			Data:       map[string]interface{}{"canSubmit": false},
		})
		w.Write(responseData)
		return

	}
	responseData, _ := json.Marshal(response.Response{
		StatusCode: http.StatusOK,
		Data:       map[string]interface{}{"canSubmit": canSubmit},
	})
	w.Write(responseData)
	return
}

// GetComplaintsByToken New creates a handler that returns the complaints of a user by their token.
// @Summary Get user complaints by token
// @Description Retrieves all complaints associated with a user based on the provided token.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param token query string true "User token"
// @Success 200 {object} response.Response "List of complaints"
// @Failure 400 {object}  response.Response "Token required"
// @Failure 401 {object} response.Response "Invalid token or failed to fetch profile"
// @Failure 500 {object} response.Response "Failed to serialize complaints"
// @Router /complaints/by-token [get]
func (h Handler) GetComplaintsByToken(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	log := h.Log
	if token == "" {
		log.Error("token required", http.StatusBadRequest)
		render.JSON(w, r, response.Response{
			Message:    "token required",
			StatusCode: http.StatusBadRequest,
		})
		return
	}
	log.Info("token", token)
	profile, err := jwt2.EncodeJWT(h.Cfg.JwtSecret, token)
	log.Info("Profile: ", profile)
	if err != nil {
		log.Error("invalid token or failed to fetch profile", "Err", response.Response{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		})
		render.JSON(w, r, response.Response{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	complaints, err := h.ComplaintService.GetComplaintsByBarcode(r.Context(), profile.Barcode)
	if errors.Is(err, storage.ErrComplaintNotFound) {
		render.JSON(w, r, response.Response{
			Message:    "User has no complaints",
			StatusCode: http.StatusOK})
	}
	if err != nil {
		log.Error("failed to get complaints", "err", err)
		render.JSON(w, r, response.Response{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	responseData, err := json.Marshal(response.Response{Data: complaints, StatusCode: http.StatusOK})
	if err != nil {
		render.JSON(w, r, response.Response{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

// DeleteByOwner @Summary Delete a complaint
// @Description Delete a complaint by its ID by owner. If the complaint is not found or user is not owner, an error is returned.
// @Tags Complaints
// @Param id path string true "Complaint ID"
// @Success 200 {object} response.Response "Complaint successfully deleted"
// @Failure 400 {object} response.Response "Invalid request or complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /complaints/{id} [delete]
func (h Handler) DeleteByOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authHeader := r.Header.Get("Authorization")
	log := h.Log
	if authHeader == "" {
		log.Error("No Authorization header found")
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		log.Error("Invalid token format")
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		log.Info("Token found", slog.String("token", tokenString))
		return []byte(h.Cfg.JwtSecret), nil
	})

	if err != nil || !token.Valid {
		log.Error("Invalid token", slog.String("token", tokenString))
		fmt.Println(err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Error("Invalid token claims")
		render.JSON(w, r, response.Response{
			Message:    "Invalid token claims",
			StatusCode: http.StatusInternalServerError,
			Data:       nil,
		})
		return
	}

	barcodeFloat, ok := claims["barcode"].(float64)
	if !ok {
		log.Error("Invalid token barcode type")
		render.JSON(w, r, response.Response{
			Message:    "Invalid token barcode",
			StatusCode: http.StatusForbidden,
			Data:       nil,
		})
		return
	}
	barcode := int(barcodeFloat)

	id := chi.URLParam(r, "id")
	complaintID, err := uuid.Parse(id)
	if err != nil {
		log.Error("Invalid complaint id", slog.String("complaintID", id))
		render.JSON(w, r, response.Response{
			Message:    "Invalid complaint id",
			StatusCode: http.StatusForbidden,
			Data:       nil,
		})
	}
	can, err := h.ComplaintService.CanUserDeleteComplaintById(ctx, complaintID, barcode)
	if errors.Is(err, storage.ErrComplaintNotFound) {
		log.Error("complaints not found", slog.String("barcode", claims["barcode"].(string)))
		render.JSON(w, r, response.Response{
			Message:    "complaints not found",
			StatusCode: http.StatusNotFound,
			Data:       nil,
		})
		return
	}
	if err != nil {
		log.Error("Error on finding owner", "err", err)
		render.JSON(w, r, response.Response{
			Message:    "Error on finding owner",
			StatusCode: http.StatusNotFound,
			Data:       nil,
		})
		return
	}
	if !can {
		log.Error("err", response.Response{
			Message:    "User is not owner of complaint",
			StatusCode: http.StatusForbidden,
			Data:       nil,
		})
		render.JSON(w, r, response.Response{
			Message:    "User is not owner of complaint",
			StatusCode: http.StatusForbidden,
			Data:       nil,
		})
		return
	}
	err = h.ComplaintService.DeleteComplaintById(ctx, complaintID)
	if err != nil {
		log.Error("Error on deleting complaint", "err", err)
		render.JSON(w, r, response.Response{Message: "Error on deleting complaint", Data: nil, StatusCode: http.StatusInternalServerError})
		return
	}
	log.Info("complaint successfully deleted")

	render.JSON(w, r, response.Response{
		Message:    "Complaint successfully deleted",
		StatusCode: http.StatusOK,
		Data:       nil,
	})
	return
}

// Update New @Summary Update a complaint
// @Description Updates an existing complaint based on the provided complaint ID and new data.
// @Tags Complaints
// @Accept json
// @Produce json
// @Param id path string true "Complaint ID"
// @Param request body Request true "Complaint resolution details"
// @Success 200 {object} Request "Complaint updated successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 404 {object} response.Response "Complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/complaints/{id} [put]
func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.complaint.update.New"

	ctx := r.Context()

	log := h.Log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	complaintID := chi.URLParam(r, "id")
	if complaintID == "" {
		log.Error("Missing complaint_id")
		render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "Missing complaint_id"})
		return
	}
	id, err := uuid.Parse(complaintID)
	if err != nil {
		log.Error("invalid complaint ID", sl.Err(err))
		render.JSON(w, r, response.Response{StatusCode: http.StatusBadRequest, Message: "invalid complaint ID"})
		return
	}

	req := struct {
		Complaint domain.Complaint `json:"data"`
	}{}

	err = render.DecodeJSON(r.Body, &req)

	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))
		render.JSON(w, r, response.Response{Message: "failed to decode request", StatusCode: http.StatusBadRequest})
		return
	}
	log.Info("request body decoded", slog.Any("request", req))

	log.Info("Complaints:", req.Complaint)
	complaint, err := h.ComplaintService.UpdateComplaint(ctx, id, req.Complaint)
	if err != nil {
		log.Error("failed to update complaint", sl.Err(err))
		render.JSON(w, r, response.Response{Message: "failed to update complaint", StatusCode: http.StatusInternalServerError})
		return
	}

	log.Info("complaint updated", slog.Any("id", id))
	render.JSON(w, r,
		response.Response{Message: complaintID, StatusCode: http.StatusOK, Data: complaint},
	)
}

// DeleteByAdmin New @Summary Delete a complaint
// @Description Delete a complaint by its ID. If the complaint is not found, an error is returned.
// @Tags Complaints
// @Param id path string true "Complaint ID"
// @Success 200 {object} response.Response "Complaint successfully deleted"
// @Failure 400 {object} response.Response "Invalid request or complaint not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/complaints/{id} [delete]
func (h Handler) DeleteByAdmin(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.complaint.deleteByAdmin.New"
	log := h.Log.With(
		slog.String("op", op),
		slog.String("url", r.URL.String()))

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Info("id can not be empty")
		render.JSON(w, r, response.Response{
			Message:    "Complaint ID is required",
			StatusCode: http.StatusBadRequest,
			Data:       nil,
		})
		return
	}

	complaintUUID, err := uuid.Parse(id)
	if err != nil {
		log.Error("invalid complaint ID format", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "Invalid complaint ID format",
			StatusCode: http.StatusBadRequest,
			Data:       nil,
		})
		return
	}
	ctx := r.Context()
	err = h.ComplaintService.DeleteComplaintById(ctx, complaintUUID)
	if errors.Is(err, storage.ErrComplaintNotFound) {
		log.Error("complaint not found", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "Complaint not found",
			StatusCode: http.StatusNotFound,
			Data:       nil,
		})
		return
	}

	if err != nil {
		log.Error("failed to deleteByAdmin complaint", sl.Err(err))
		render.JSON(w, r, response.Response{
			Message:    "Internal error while deleting complaint",
			StatusCode: http.StatusInternalServerError,
			Data:       nil,
		})
		return
	}

	log.Info("complaint successfully deleted")

	render.JSON(w, r, response.Response{
		Message:    "Complaint successfully deleted",
		StatusCode: http.StatusOK,
		Data:       nil,
	})
}
