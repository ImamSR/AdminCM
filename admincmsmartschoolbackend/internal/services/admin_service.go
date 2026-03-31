package services

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/models"
)

func HandleAdmins(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		getAdminsList(w, r)
	case "POST":
		createAdmin(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleAdminByID(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid admin ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		updateAdmin(w, r, id)
	case "DELETE":
		deleteAdmin(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getAdminsList(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, name, email, role, COALESCE(unit, ''), COALESCE(is_active, TRUE)
		FROM admin_users
		WHERE COALESCE(is_active, TRUE) = TRUE
		ORDER BY id ASC
	`
	rows, err := database.DB.Query(query)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var admins []models.AdminDetail
	for rows.Next() {
		var a models.AdminDetail
		if err := rows.Scan(&a.ID, &a.Name, &a.Email, &a.Role, &a.Unit, &a.IsActive); err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		admins = append(admins, a)
	}

	if len(admins) == 0 {
		json.NewEncoder(w).Encode([]models.AdminDetail{})
	} else {
		json.NewEncoder(w).Encode(admins)
	}
}

func createAdmin(w http.ResponseWriter, r *http.Request) {
	var req models.AdminCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Name == "" || req.Role == "" {
		http.Error(w, "Name, Email, and Role are required", http.StatusBadRequest)
		return
	}

	var newID int
	err := database.DB.QueryRow(`
		INSERT INTO admin_users (name, email, role, unit, is_active) 
		VALUES ($1, $2, $3, NULLIF($4, ''), TRUE) RETURNING id
	`, req.Name, req.Email, req.Role, req.Unit).Scan(&newID)
	
	if err != nil {
		http.Error(w, "Insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": newID, "status": "success"})
}

func updateAdmin(w http.ResponseWriter, r *http.Request, id int) {
	var req models.AdminUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	res, err := database.DB.Exec(`
		UPDATE admin_users 
		SET name = $1, email = $2, role = $3, unit = NULLIF($4, '')
		WHERE id = $5
	`, req.Name, req.Email, req.Role, req.Unit, id)
	
	if err != nil {
		http.Error(w, "Update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Admin not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func deleteAdmin(w http.ResponseWriter, _ *http.Request, id int) {
	res, err := database.DB.Exec(`
		UPDATE admin_users 
		SET is_active = FALSE 
		WHERE id = $1
	`, id)
	
	if err != nil {
		http.Error(w, "Delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Admin not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}
