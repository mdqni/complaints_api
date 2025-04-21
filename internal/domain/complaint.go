package domain

import (
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type Complaint struct {
	ID        uuid.UUID      `json:"id" example:"71cb6c7d-5c1c-4a5f-bc08-204b6d435c25"`
	Barcode   int            `json:"barcode" example:"242590"`
	Category  Category       `json:"category"`
	Message   string         `json:"message" example:"Too noisy in the dorm"`
	Status    string         `json:"status" example:"approved"`
	CreatedAt time.Time      `json:"created_at" example:"2025-04-21T12:00:00Z"`
	UpdatedAt time.Time      `json:"updated_at" example:"2025-04-21T14:00:00Z"`
	Answer    sql.NullString `json:"answer" swaggertype:"string" example:"Complaint resolved"`
}

type ComplaintStatus string

const (
	StatusApproved ComplaintStatus = "Approved"
	StatusPending  ComplaintStatus = "Pending"
	StatusRejected ComplaintStatus = "Rejected"
)
