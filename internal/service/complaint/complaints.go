package service

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/storage"
	"complaint_server/internal/storage/pg"
	"context"
	"errors"
	"fmt"
)

type ComplaintService struct {
	storage *pg.Storage
}

func NewComplaintsService(strg *pg.Storage) *ComplaintService {
	return &ComplaintService{storage: strg}
}

// CreateComplaint создаёт жалобу с проверкой ограничений.
func (s *ComplaintService) CreateComplaint(ctx context.Context, barcode string, categoryID int, message string) (int, string, error) {
	// Проверяем, можно ли отправить жалобу
	canSubmit, err := s.storage.CheckComplaintLimit(ctx, barcode)
	if err != nil {
		return 0, "", fmt.Errorf("failed to check complaint limit: %w", err)
	}
	if !canSubmit {
		return 0, "", storage.ErrLimitOneComplaintInOneHour
	}

	// Сохраняем жалобу
	complaintID, answer, err := s.storage.SaveComplaint(ctx, barcode, categoryID, message)
	if err != nil {
		return 0, "", fmt.Errorf("failed to save complaint: %w", err)
	}

	return complaintID, answer, nil
}

// GetComplaintById получает жалобу по ID
func (s *ComplaintService) GetComplaintById(ctx context.Context, complaintID int) (domain.Complaint, error) {
	complaint, err := s.storage.GetComplaintById(ctx, complaintID)
	if err != nil {
		return domain.Complaint{}, err
	}
	return complaint, nil
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
func (s *ComplaintService) GetComplaintsByCategoryId(ctx context.Context, categoryId int) ([]domain.Complaint, error) {
	return s.storage.GetComplaintsByCategoryId(ctx, categoryId)
}

// UpdateComplaintStatus обновляет статус жалобы
func (s *ComplaintService) UpdateComplaintStatus(ctx context.Context, complaintID int64, status domain.ComplaintStatus, answer string) error {
	return s.storage.UpdateComplaintStatus(ctx, complaintID, status, answer)
}

// UpdateComplaint обновляет жалобу
func (s *ComplaintService) UpdateComplaint(ctx context.Context, complaintID int64, complaint domain.Complaint) (domain.Complaint, error) {
	return s.storage.UpdateComplaint(ctx, complaintID, complaint)
}

// DeleteComplaintById удаляет жалобу
func (s *ComplaintService) DeleteComplaintById(ctx context.Context, complaintID int) error {
	return s.storage.DeleteComplaint(ctx, complaintID)
}

// CanSubmitByBarcode проверяет может ли отправить
func (s *ComplaintService) CanSubmitByBarcode(ctx context.Context, userID string) (bool, error) {
	return s.storage.CheckComplaintLimit(ctx, userID)
}

// GetComplaintsByBarcode
func (s *ComplaintService) GetComplaintsByBarcode(ctx context.Context, barcode string) ([]domain.Complaint, error) {
	return s.storage.GetComplaintsByBarcode(ctx, barcode)
}
