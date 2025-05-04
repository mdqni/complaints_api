package serviceComplaint

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/storage"
	"complaint_server/internal/storage/pg"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
)

type ComplaintService struct {
	storage *pg.Storage
}

func NewComplaintsService(strg *pg.Storage) *ComplaintService {
	return &ComplaintService{storage: strg}
}

// CreateComplaint создаёт жалобу с проверкой ограничений.
func (s *ComplaintService) CreateComplaint(ctx context.Context, barcode int, categoryID uuid.UUID, message string) (uuid.UUID, string, error) {
	// Проверяем, можно ли отправить жалобу
	canSubmit, err := s.storage.CheckComplaintLimit(ctx, barcode)
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("failed to check complaint limit: %w", err)
	}
	if !canSubmit {
		return uuid.UUID{}, "", storage.ErrLimitOneComplaintInOneHour
	}

	// Сохраняем жалобу
	complaintID, answer, err := s.storage.SaveComplaint(ctx, barcode, categoryID, message)
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("failed to save complaint: %w", err)
	}

	return complaintID, answer, nil
}

// GetComplaintById получает жалобу по ID
func (s *ComplaintService) GetComplaintByUUID(ctx context.Context, complaintID uuid.UUID) (domain.Complaint, error) {
	complaint, err := s.storage.GetComplaintByUUID(ctx, complaintID)
	return complaint, err
}

// GetAllComplaints получает все жалобы
func (s *ComplaintService) GetAllComplaints(ctx context.Context) ([]domain.Complaint, error) {
	complaints, err := s.storage.GetComplaints(ctx)
	if errors.Is(err, storage.ErrComplaintNotFound) {
		return nil, storage.ErrComplaintNotFound
	}
	if err != nil {
		return []domain.Complaint{}, err
	}

	return complaints, nil
}

// GetComplaintsByCategoryId получает жалобы по категории
func (s *ComplaintService) GetComplaintsByCategoryId(ctx context.Context, categoryId uuid.UUID) ([]domain.Complaint, error) {
	return s.storage.GetComplaintsByCategoryId(ctx, categoryId)
}

// UpdateComplaintStatus обновляет статус жалобы
func (s *ComplaintService) UpdateComplaintStatus(ctx context.Context, complaintID uuid.UUID, status domain.ComplaintStatus, answer string) error {
	return s.storage.UpdateComplaintStatus(ctx, complaintID, status, answer)
}

// UpdateComplaint обновляет жалобу
func (s *ComplaintService) UpdateComplaint(ctx context.Context, complaintID uuid.UUID, complaint domain.Complaint) (domain.Complaint, error) {
	return s.storage.UpdateComplaint(ctx, complaintID, complaint)
}

// DeleteComplaintById удаляет жалобу
func (s *ComplaintService) DeleteComplaintById(ctx context.Context, complaintID uuid.UUID) error {
	return s.storage.DeleteComplaint(ctx, complaintID)
}

// CanSubmitByBarcode проверяет может ли отправить
func (s *ComplaintService) CanSubmitByBarcode(ctx context.Context, barcode int) (bool, error) {
	return s.storage.CheckComplaintLimit(ctx, barcode)
}

func (s *ComplaintService) CanUserDeleteComplaintById(ctx context.Context, complaintID uuid.UUID, barcode int) (bool, error) {
	can, err := s.storage.IsOwnerOfComplaint(ctx, complaintID, barcode)
	return can, err
}

// GetComplaintsByBarcode
func (s *ComplaintService) GetComplaintsByBarcode(ctx context.Context, barcode int) ([]domain.Complaint, error) {
	complaints, err := s.storage.GetComplaintsByBarcode(ctx, barcode)
	if errors.Is(err, storage.ErrComplaintNotFound) {
		fmt.Println("❌ scan error:", err)
		return nil, storage.ErrComplaintNotFound
	}
	if err != nil {
		fmt.Println("❌ scan error:", err)
		return []domain.Complaint{}, err
	}
	return complaints, err
}
