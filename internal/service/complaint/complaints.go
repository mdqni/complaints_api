package serviceComplaint

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/repository"
	"complaint_server/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
)

type ComplaintService struct {
	repo repository.ComplaintRepository
}

func NewComplaintsService(repo repository.ComplaintRepository) *ComplaintService {
	return &ComplaintService{repo: repo}
}

func (s *ComplaintService) CreateComplaint(ctx context.Context, barcode int, categoryID uuid.UUID, message string) (uuid.UUID, string, error) {
	canSubmit, err := s.repo.CheckComplaintLimit(ctx, barcode)
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("failed to check complaint limit: %w", err)
	}
	if !canSubmit {
		return uuid.UUID{}, "", storage.ErrLimitOneComplaintInOneHour
	}

	complaintID, answer, err := s.repo.SaveComplaint(ctx, barcode, categoryID, message)
	if err != nil {
		return uuid.UUID{}, "", fmt.Errorf("failed to save complaint: %w", err)
	}

	return complaintID, answer, nil
}

func (s *ComplaintService) GetComplaintByUUID(ctx context.Context, complaintID uuid.UUID) (domain.Complaint, error) {
	complaint, err := s.repo.GetComplaintByUUID(ctx, complaintID)
	return complaint, err
}

func (s *ComplaintService) GetAllComplaints(ctx context.Context) ([]domain.Complaint, error) {
	complaints, err := s.repo.GetComplaints(ctx)
	if errors.Is(err, storage.ErrComplaintNotFound) {
		return nil, storage.ErrComplaintNotFound
	}
	if err != nil {
		return []domain.Complaint{}, err
	}

	return complaints, nil
}

func (s *ComplaintService) GetComplaintsByCategoryId(ctx context.Context, categoryId uuid.UUID) ([]domain.Complaint, error) {
	return s.repo.GetComplaintsByCategoryId(ctx, categoryId)
}

func (s *ComplaintService) UpdateComplaintStatus(ctx context.Context, complaintID uuid.UUID, status domain.ComplaintStatus, answer string) error {
	return s.repo.UpdateComplaintStatus(ctx, complaintID, status, answer)
}

func (s *ComplaintService) UpdateComplaint(ctx context.Context, complaintID uuid.UUID, complaint domain.Complaint) (domain.Complaint, error) {
	return s.repo.UpdateComplaint(ctx, complaintID, complaint)
}

func (s *ComplaintService) DeleteComplaintById(ctx context.Context, complaintID uuid.UUID) error {
	return s.repo.DeleteComplaint(ctx, complaintID)
}

func (s *ComplaintService) CanSubmitByBarcode(ctx context.Context, barcode int) (bool, error) {
	return s.repo.CheckComplaintLimit(ctx, barcode)
}

func (s *ComplaintService) CanUserDeleteComplaintById(ctx context.Context, complaintID uuid.UUID, barcode int) (bool, error) {
	can, err := s.repo.IsOwnerOfComplaint(ctx, complaintID, barcode)
	return can, err
}

func (s *ComplaintService) GetComplaintsByBarcode(ctx context.Context, barcode int) ([]domain.Complaint, error) {
	complaints, err := s.repo.GetComplaintsByBarcode(ctx, barcode)
	if errors.Is(err, storage.ErrComplaintNotFound) {
		fmt.Println("scan error:", err)
		return nil, storage.ErrComplaintNotFound
	}
	if err != nil {
		fmt.Println("scan error:", err)
		return []domain.Complaint{}, err
	}
	return complaints, err
}
