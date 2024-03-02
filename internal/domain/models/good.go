package models

import "time"

type Good struct {
	ID          int       `json:"id"`
	ProjectId   int       `json:"project_id" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Priority    *int      `json:"priority"`
	Removed     bool      `json:"removed"`
	CreatedAt   time.Time `json:"created_at"`
}
