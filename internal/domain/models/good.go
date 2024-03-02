package models

import "time"

type Good struct {
	ID          int       `json:"id" redis:"id"`
	ProjectId   int       `json:"project_id" redis:"project_id" validate:"required"`
	Name        string    `json:"name" redis:"name" validate:"required"`
	Description string    `json:"description" redis:"description"`
	Priority    *int      `json:"priority" redis:"priority"`
	Removed     bool      `json:"removed" redis:"removed"`
	CreatedAt   time.Time `json:"created_at" redis:"created_at"`
}
