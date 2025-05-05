package serviceComplaint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"complaint_server/internal/domain"
	"complaint_server/internal/repository"
	"complaint_server/internal/storage"

	"github.com/google/uuid"
)

const cacheKey = "cache:/complaints"

type ComplaintService struct {
	repo  repository.ComplaintRepository
	cache storage.Cache
}

func NewComplaintsService(repo repository.ComplaintRepository, cache storage.Cache) *ComplaintService {
	return &ComplaintService{repo: repo, cache: cache}
}

func (s *ComplaintService) CreateComplaint(ctx context.Context, barcode int, categoryID uuid.UUID, message string) (uuid.UUID, string, error) {
	canSubmit, err := s.repo.CheckComplaintLimit(ctx, barcode)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to check complaint limit: %w", err)
	}
	if !canSubmit {
		return uuid.Nil, "", storage.ErrLimitOneComplaintInOneHour
	}

	complaintID, answer, err := s.repo.SaveComplaint(ctx, barcode, categoryID, message)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to save complaint: %w", err)
	}

	_ = s.cache.Delete(ctx, cacheKey)
	return complaintID, answer, nil
}

func (s *ComplaintService) GetAllComplaints(ctx context.Context) ([]domain.Complaint, error) {
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil && len(cachedData) > 0 {
		var cachedComplaints []domain.Complaint
		if err := json.Unmarshal([]byte(cachedData), &cachedComplaints); err == nil {
			return cachedComplaints, nil
		}
	}

	complaints, err := s.repo.GetComplaints(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrComplaintNotFound) {
			return nil, storage.ErrComplaintNotFound
		}
		return nil, err
	}

	data, err := json.Marshal(complaints)
	if err == nil {
		_ = s.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return complaints, nil
}

func (s *ComplaintService) GetComplaintByUUID(ctx context.Context, complaintID uuid.UUID) (domain.Complaint, error) {
	return s.repo.GetComplaintByUUID(ctx, complaintID)
}

func (s *ComplaintService) GetComplaintsByCategoryId(ctx context.Context, categoryID uuid.UUID) ([]domain.Complaint, error) {
	return s.repo.GetComplaintsByCategoryId(ctx, categoryID)
}

func (s *ComplaintService) GetComplaintsByBarcode(ctx context.Context, barcode int) ([]domain.Complaint, error) {
	complaints, err := s.repo.GetComplaintsByBarcode(ctx, barcode)
	if errors.Is(err, storage.ErrComplaintNotFound) {
		return nil, storage.ErrComplaintNotFound
	}
	if err != nil {
		return nil, err
	}
	return complaints, nil
}

func (s *ComplaintService) UpdateComplaintStatus(ctx context.Context, complaintID uuid.UUID, status domain.ComplaintStatus, answer string) error {
	err := s.repo.UpdateComplaintStatus(ctx, complaintID, status, answer)
	if err == nil {
		_ = s.cache.Delete(ctx, cacheKey)
	}
	return err
}

func (s *ComplaintService) UpdateComplaint(ctx context.Context, complaintID uuid.UUID, complaint domain.Complaint) (domain.Complaint, error) {
	updatedComplaint, err := s.repo.UpdateComplaint(ctx, complaintID, complaint)
	if err == nil {
		_ = s.cache.Delete(ctx, cacheKey)
	}
	return updatedComplaint, err
}

func (s *ComplaintService) DeleteComplaintById(ctx context.Context, complaintID uuid.UUID) error {
	err := s.repo.DeleteComplaint(ctx, complaintID)
	if err == nil {
		_ = s.cache.Delete(ctx, cacheKey)
	}
	return err
}

func (s *ComplaintService) CanSubmitByBarcode(ctx context.Context, barcode int) (bool, error) {
	return s.repo.CheckComplaintLimit(ctx, barcode)
}

func (s *ComplaintService) CanUserDeleteComplaintById(ctx context.Context, complaintID uuid.UUID, barcode int) (bool, error) {
	return s.repo.IsOwnerOfComplaint(ctx, complaintID, barcode)
}
