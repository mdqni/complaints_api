package domain

import (
	"database/sql"
	"time"
)

type Complaint struct {
	ID        int            `json:"id"`
	UserUUID  string         `json:"user_uuid"`
	Category  Category       `json:"category"`
	Message   string         `json:"message"`
	Status    string         `json:"status"` // "approved" or "rejected", default is "rending"
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Answer    sql.NullString `json:"answer" swaggertype:"string"`
}

type ComplaintStatus string

const (
	StatusApproved ComplaintStatus = "Approved"
	StatusPending  ComplaintStatus = "Pending"
	StatusRejected ComplaintStatus = "Rejected"
)
