package services

import (
	"encoding/json"
	"net/http"
	"strings"

	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/models"
)

func HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	unit := r.URL.Query().Get("unit")
	if unit == "" {
		unit = "all"
	}
	unit = strings.ToUpper(unit)

	var stats models.DashboardStats

	studentQuery := `
		SELECT COUNT(DISTINCT s.id) 
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE u.role = 'siswa' AND COALESCE(u.is_active, TRUE) = TRUE AND ($1 = 'ALL' OR UPPER(u.unit) = $1)
	`
	database.DB.QueryRow(studentQuery, unit).Scan(&stats.TotalStudents)

	teacherQuery := `
		SELECT COUNT(id) 
		FROM users u
		WHERE u.role IN ('guru', 'wali_kelas', 'kepala_sekolah', 'wakil_kepala_sekolah') AND COALESCE(u.is_active, TRUE) = TRUE AND ($1 = 'ALL' OR UPPER(u.unit) = $1)
	`
	database.DB.QueryRow(teacherQuery, unit).Scan(&stats.TotalTeachers)
	classQuery := `
		SELECT COUNT(*) 
		FROM classes 
		WHERE $1 = 'ALL' OR UPPER(unit) = $1
	`
	database.DB.QueryRow(classQuery, unit).Scan(&stats.TotalClasses)
	bannerCountQuery := `
		SELECT COUNT(id)
		FROM banners
		WHERE $1 = 'ALL' OR UPPER(unit) = $1
	`
	database.DB.QueryRow(bannerCountQuery, unit).Scan(&stats.TotalBanners)

	recentBannersQuery := `
		SELECT id, 'Informasi' as type, COALESCE(title, 'Tanpa Judul') as title, COALESCE(description, 'Tidak ada deskripsi') as desc, unit, created_at
		FROM banners
		WHERE $1 = 'ALL' OR UPPER(unit) = $1
		ORDER BY created_at DESC
		LIMIT 5
	`
	rows, err := database.DB.Query(recentBannersQuery, unit)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var b models.BannerSummary
			if err := rows.Scan(&b.ID, &b.Type, &b.Title, &b.Desc, &b.Unit, &b.CreatedAt); err == nil {
				stats.RecentBanners = append(stats.RecentBanners, b)
			}
		}
	}

	if stats.RecentBanners == nil {
		stats.RecentBanners = []models.BannerSummary{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}
