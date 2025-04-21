package domain

import "time"

type Admin struct {
	ID           int       `json:"id"`
	Barcode      int       `json:"barcode"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
