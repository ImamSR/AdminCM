package models

import "time"

type DashboardStats struct {
	TotalStudents int             `json:"total_students"`
	TotalTeachers int             `json:"total_teachers"`
	TotalClasses  int             `json:"total_classes"`
	TotalBanners  int             `json:"total_banners"`
	RecentBanners []BannerSummary `json:"recent_banners"`
}

type BannerSummary struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	Unit      string    `json:"unit"`
	CreatedAt time.Time `json:"created_at"`
}
