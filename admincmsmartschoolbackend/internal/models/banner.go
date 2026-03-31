package models

import "time"



type Banner struct {
	ID          int       `json:"id"`
	Unit        string    `json:"unit"`
	Grade       int       `json:"grade"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	LinkAction  string    `json:"link_action"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
