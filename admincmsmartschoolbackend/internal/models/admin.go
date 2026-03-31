package models

import "time"

type AdminDetail struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Unit      string    `json:"unit"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type AdminCreateReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Unit  string `json:"unit"`
}

type AdminUpdateReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Unit  string `json:"unit"`
}
