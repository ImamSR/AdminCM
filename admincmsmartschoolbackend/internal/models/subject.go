package models

import "time"

type Subject struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Unit      string    `json:"unit"`
	Grade     *int      `json:"grade"`
	CreatedAt time.Time `json:"created_at"`
}

type SubjectCreateReq struct {
	Name  string `json:"name"`
	Code  string `json:"code"`
	Unit  string `json:"unit"`
	Grade *int   `json:"grade"`
}

type SubjectUpdateReq struct {
	Name  string `json:"name"`
	Code  string `json:"code"`
	Unit  string `json:"unit"`
	Grade *int   `json:"grade"`
}
