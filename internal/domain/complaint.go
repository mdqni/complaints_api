package domain

import (
	"database/sql"
	"time"
)

type Complaint struct {
	ID         int            `json:"id"`
	UserUUID   string         `json:"user_uuid"`
	CategoryID int            `json:"category_id"`
	Message    string         `json:"message"`
	Status     string         `json:"status"` // active, resolved, or rejected
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Answer     sql.NullString `json:"answer"`
}

type ComplaintStatus string

const (
	StatusApproved ComplaintStatus = "Approved"
	StatusPending  ComplaintStatus = "Pending"
	StatusRejected ComplaintStatus = "Rejected"
)
