package models

import "time"

type Unit struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Level     string    `json:"level"`
	AccentBar string    `json:"accentBar"`
	CreatedAt time.Time `json:"created_at"`
}
