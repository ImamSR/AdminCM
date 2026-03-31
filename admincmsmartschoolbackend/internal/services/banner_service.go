package services

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/models"
)

func HandleBanners(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		getBanners(w, r)
	case "POST":
		createBanner(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleBannerByID(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid banner ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		updateBanner(w, r, id)
	case "DELETE":
		deleteBanner(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getBanners(w http.ResponseWriter, r *http.Request) {
	unit := r.URL.Query().Get("unit")
	if unit == "" {
		http.Error(w, "Missing unit parameter", http.StatusBadRequest)
		return
	}
	unit = strings.ToUpper(unit)

	query := `
		SELECT id, unit, COALESCE(grade, 0), COALESCE(title, ''), COALESCE(description, ''), image_url, COALESCE(link_action, ''), COALESCE(is_active, TRUE), created_at
		FROM banners
		WHERE UPPER(unit) = $1
		ORDER BY created_at DESC
	`
	rows, err := database.DB.Query(query, unit)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var bs []models.Banner
	for rows.Next() {
		var b models.Banner
		if err := rows.Scan(&b.ID, &b.Unit, &b.Grade, &b.Title, &b.Description, &b.ImageURL, &b.LinkAction, &b.IsActive, &b.CreatedAt); err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		bs = append(bs, b)
	}

	if len(bs) == 0 {
		json.NewEncoder(w).Encode([]models.Banner{})
	} else {
		json.NewEncoder(w).Encode(bs)
	}
}

func createBanner(w http.ResponseWriter, r *http.Request) {
	var req models.Banner
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if req.Unit == "" {
		http.Error(w, "Unit is required", http.StatusBadRequest)
		return
	}

	var newID int
	err := database.DB.QueryRow(`
		INSERT INTO banners (unit, grade, title, description, image_url, link_action, is_active) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`, strings.ToUpper(req.Unit), req.Grade, req.Title, req.Description, req.ImageURL, req.LinkAction, req.IsActive).Scan(&newID)
	
	if err != nil {
		http.Error(w, "Insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": newID, "status": "success"})
}

func updateBanner(w http.ResponseWriter, r *http.Request, id int) {
	var req models.Banner
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	res, err := database.DB.Exec(`
		UPDATE banners 
		SET unit = $1, grade = $2, title = $3, description = $4, image_url = $5, link_action = $6, is_active = $7
		WHERE id = $8
	`, strings.ToUpper(req.Unit), req.Grade, req.Title, req.Description, req.ImageURL, req.LinkAction, req.IsActive, id)
	
	if err != nil {
		http.Error(w, "Update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Banner not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func deleteBanner(w http.ResponseWriter, _ *http.Request, id int) {
	res, err := database.DB.Exec(`DELETE FROM banners WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Banner not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}
