package domain

import "github.com/google/uuid"

type Category struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"` // Detailed description of the categories
	Answer      string    `json:"answer"`
}
