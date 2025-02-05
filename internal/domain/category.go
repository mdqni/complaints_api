package domain

type Category struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"` // Detailed description of the category
	Answer      string `json:"answer"`
}
