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

	// Очищаем кеш после изменения БД
	//redis.Del(ctx, "cache:/complaints")
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
	//key := "cache:/complaints"
	//// Проверяем кеш
	//cachedData, err := redis.Get(ctx, key).Result()
	//if err == nil {
	//	// Данные найдены в кеше, декодируем их
	//	var complaints []domain.Complaint
	//	err := json.Unmarshal([]byte(cachedData), &complaints)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return complaints, nil
	//}
	complaints, err := s.storage.GetComplaints()
	if errors.Is(err, storage.ErrComplaintNotFound) {
		return nil, storage.ErrComplaintNotFound
	}
	if err != nil {
		return []domain.Complaint{}, err
	}

	//// Сохраняем в Redis на 5 минут
	//jsonData, _ := json.Marshal(complaints)
	//redis.Set(ctx, key, jsonData, 5*time.Minute)

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
func (s *ComplaintService) DeleteComplaint(ctx context.Context, complaintID int) error {
	return s.storage.DeleteComplaint(ctx, complaintID)
}
