package services

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/models"
)

type PimpinanCreateReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type PimpinanUpdateReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func HandlePimpinanUsers(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		getPimpinanList(w, r)
	case "POST":
		createPimpinan(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandlePimpinanUserByID(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]
	if idStr == "" && len(parts) > 1 {
		idStr = parts[len(parts)-2]
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		updatePimpinan(w, r, id)
	case "DELETE":
		deletePimpinan(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getPimpinanList(w http.ResponseWriter, r *http.Request) {
	roleParam := r.URL.Query().Get("role")
	roleFilter := "pimpinan"
	if roleParam != "" {
		roleFilter = strings.ToLower(roleParam)
	}

	query := `
		SELECT id, name, email, role, COALESCE(unit, '') 
		FROM users 
		WHERE LOWER(role) = $1
		ORDER BY id ASC
	`
	rows, err := database.DB.Query(query, roleFilter)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Unit); err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	if len(users) == 0 {
		json.NewEncoder(w).Encode([]models.User{})
	} else {
		json.NewEncoder(w).Encode(users)
	}
}

func createPimpinan(w http.ResponseWriter, r *http.Request) {
	var req PimpinanCreateReq
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
		INSERT INTO users (name, email, role) 
		VALUES ($1, $2, $3) RETURNING id
	`, req.Name, req.Email, strings.ToLower(req.Role)).Scan(&newID)
	
	if err != nil {
		http.Error(w, "Insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": newID, "status": "success"})
}

func updatePimpinan(w http.ResponseWriter, r *http.Request, id int) {
	var req PimpinanUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	res, err := database.DB.Exec(`
		UPDATE users 
		SET name = $1, email = $2, role = $3
		WHERE id = $4
	`, req.Name, req.Email, strings.ToLower(req.Role), id)
	
	if err != nil {
		http.Error(w, "Update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func deletePimpinan(w http.ResponseWriter, _ *http.Request, id int) {
	res, err := database.DB.Exec(`
		DELETE FROM users WHERE id = $1
	`, id)
	
	if err != nil {
		http.Error(w, "Delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}
