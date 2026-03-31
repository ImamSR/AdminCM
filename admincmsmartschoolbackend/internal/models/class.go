package models

import "time"

type Class struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Level       string    `json:"level"`
	Unit        string    `json:"unit"`
	Grade       int       `json:"grade"`
	ClassName   *string   `json:"class_name"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type ClassDetail struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Level        string `json:"level"`
	Unit         string `json:"unit"`
	Grade        int    `json:"grade"`
	ClassName    string `json:"class_name"`
	Teacher      string `json:"teacher"`
	TeacherCount int    `json:"teacherCount"`
	StudentCount int    `json:"studentCount"`
}

type ClassCreateReq struct {
	Name       string `json:"name"`
	Unit       string `json:"unit"`
	TeacherIDs []int  `json:"teacher_ids"`
}

type ClassUpdateReq struct {
	Name       string `json:"name"`
	TeacherIDs []int  `json:"teacher_ids"`
}

type ClassStats struct {
	Unit  string `json:"unit"`
	Count int    `json:"count"`
}
