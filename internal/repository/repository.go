package repository

import (
	"complaint_server/internal/domain"
	"context"
	"github.com/google/uuid"
)

type CategoryRepository interface {
	Create(ctx context.Context, category domain.Category) (uuid.UUID, error)
	GetAll(ctx context.Context) ([]domain.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Category, error)
	Update(ctx context.Context, id uuid.UUID, category domain.Category) (uuid.UUID, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ComplaintRepository interface {
	SaveComplaint(ctx context.Context, barcode int, categoryID uuid.UUID, message string) (uuid.UUID, string, error)
	IsOwnerOfComplaint(ctx context.Context, id uuid.UUID, barcode int) (bool, error)
	GetComplaintByUUID(ctx context.Context, id uuid.UUID) (domain.Complaint, error)
	GetComplaints(ctx context.Context) ([]domain.Complaint, error)
	GetComplaintsByCategoryId(ctx context.Context, categoryId uuid.UUID) ([]domain.Complaint, error)
	GetComplaintsByBarcode(ctx context.Context, barcode int) ([]domain.Complaint, error)
	CheckComplaintLimit(ctx context.Context, barcode int) (bool, error)
	UpdateComplaintStatus(ctx context.Context, id uuid.UUID, status domain.ComplaintStatus, answer string) error
	DeleteComplaint(ctx context.Context, id uuid.UUID) error
	UpdateComplaint(ctx context.Context, id uuid.UUID, complaint domain.Complaint) (domain.Complaint, error)
}
