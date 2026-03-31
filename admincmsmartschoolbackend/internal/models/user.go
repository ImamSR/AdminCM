package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Unit      string    `json:"unit"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}
