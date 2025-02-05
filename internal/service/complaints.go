package service

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/storage"
	"complaint_server/internal/storage/pg"
	"fmt"
)

// ComplaintStorage интерфейс для хранилища жалоб
type ComplaintStorage interface {
	SaveComplaint(userUUID string, categoryID int, message string) (string, error)
	GetComplaintById(complaintID int) (domain.Complaint, error)
	GetComplaintsByCategoryId(categoryId int) ([]domain.Complaint, error)
	UpdateComplaintStatus(complaintID int64, status domain.ComplaintStatus, answer string) error
	CheckComplaintLimit(userUUID string) (bool, error)
	DeleteComplaint(id int) error
}
type ComplaintService struct {
	storage *pg.Storage
}

func New(strg *pg.Storage) *ComplaintService {
	return &ComplaintService{storage: strg}
}

// CreateComplaint создаёт жалобу с проверкой ограничений.
func (s *ComplaintService) CreateComplaint(userUUID string, categoryID int, message string) (string, error) {
	// Проверяем, можно ли отправить жалобу
	canSubmit, err := s.storage.CheckComplaintLimit(userUUID)
	if err != nil {
		return "", fmt.Errorf("failed to check complaint limit: %w", err)
	}
	if !canSubmit {
		return "", storage.ErrLimitOneComplaintInOneHour
	}

	// Сохраняем жалобу
	answer, err := s.storage.SaveComplaint(userUUID, categoryID, message)
	if err != nil {
		return "", fmt.Errorf("failed to save complaint: %w", err)
	}

	return answer, nil
}

// GetComplaintById получает жалобу по ID
func (s *ComplaintService) GetComplaintById(complaintID int) (domain.Complaint, error) {
	complaint, err := s.storage.GetComplaintById(complaintID)
	if err != nil {
		return domain.Complaint{}, err
	}
	return complaint, nil
}

// GetAllComplaints получает все жалобы
func (s *ComplaintService) GetAllComplaints() ([]domain.Complaint, error) {
	complaints, err := s.storage.GetComplaints()
	if err != nil {
		return []domain.Complaint{}, err
	}
	return complaints, nil
}

// GetComplaintsByCategoryId получает жалобы по категории
func (s *ComplaintService) GetComplaintsByCategoryId(categoryId int) ([]domain.Complaint, error) {
	return s.storage.GetComplaintsByCategoryId(categoryId)
}

// UpdateComplaintStatus обновляет статус жалобы
func (s *ComplaintService) UpdateComplaintStatus(complaintID int64, status domain.ComplaintStatus, answer string) error {
	return s.storage.UpdateComplaintStatus(complaintID, status, answer)
}

// DeleteComplaint удаляет жалобу
func (s *ComplaintService) DeleteComplaint(complaintID int) error {
	return s.storage.DeleteComplaint(complaintID)
}
