package complaintService

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/storage"
	"complaint_server/internal/storage/pg"
	"context"
	"fmt"
)

// ComplaintStorage интерфейс для хранилища жалоб
type ComplaintStorage interface {
	SaveComplaint(ctx context.Context, userUUID string, categoryID int, message string) (string, error)
	GetComplaintById(ctx context.Context, complaintID int) (domain.Complaint, error)
	GetComplaintsByCategoryId(ctx context.Context, categoryId int) ([]domain.Complaint, error)
	UpdateComplaintStatus(ctx context.Context, complaintID int64, status domain.ComplaintStatus, answer string) error
	CheckComplaintLimit(ctx context.Context, userUUID string) (bool, error)
	DeleteComplaint(ctx context.Context, id int) error
	UpdateComplaint(ctx context.Context, complaintID int64, complaint domain.Complaint) error
}
type ComplaintService struct {
	storage *pg.Storage
}

func New(strg *pg.Storage) *ComplaintService {
	return &ComplaintService{storage: strg}
}

// CreateComplaint создаёт жалобу с проверкой ограничений.
func (s *ComplaintService) CreateComplaint(ctx context.Context, userUUID string, categoryID int, message string) (string, error) {
	// Проверяем, можно ли отправить жалобу
	canSubmit, err := s.storage.CheckComplaintLimit(ctx, userUUID)
	if err != nil {
		return "", fmt.Errorf("failed to check complaint limit: %w", err)
	}
	if !canSubmit {
		return "", storage.ErrLimitOneComplaintInOneHour
	}

	// Сохраняем жалобу
	answer, err := s.storage.SaveComplaint(ctx, userUUID, categoryID, message)
	if err != nil {
		return "", fmt.Errorf("failed to save complaint: %w", err)
	}

	return answer, nil
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
	if err != nil {
		return []domain.Complaint{}, err
	}
	return complaints, nil
}

// GetComplaintsByCategoryId получает жалобы по категории
func (s *ComplaintService) GetComplaintsByCategoryId(ctx context.Context, categoryId int) ([]domain.Complaint, error) {
	return s.storage.GetComplaintsByCategoryId(ctx, categoryId)
}

// UpdateComplaint обновляет жалобу
func (s *ComplaintService) UpdateComplaint(ctx context.Context, complaintID int64, complaint domain.Complaint) error {
	return s.storage.UpdateComplaint(ctx, complaintID, complaint)
}

// UpdateComplaintStatus обновляет статус жалобы
func (s *ComplaintService) UpdateComplaintStatus(ctx context.Context, complaintID int64, status domain.ComplaintStatus, answer string) error {
	return s.storage.UpdateComplaintStatus(ctx, complaintID, status, answer)
}

// DeleteComplaint удаляет жалобу
func (s *ComplaintService) DeleteComplaint(ctx context.Context, complaintID int) error {
	return s.storage.DeleteComplaint(ctx, complaintID)
}
