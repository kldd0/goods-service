package models

import "time"

type Good struct {
	ID          int       `json:"id"`
	ProjectId   int       `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Removed     bool      `json:"removed"`
	CreatedAt   time.Time `json:"created_at"`
}
