package services

import (
	"encoding/json"
	"net/http"
	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/models"
)

func GetAdmins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `
		SELECT id, name, email, role, COALESCE(unit, '') 
		FROM users 
		WHERE LOWER(role) IN ('admin', 'super admin', 'operator', 'superadmin')
	`
	rows, err := database.DB.Query(query)
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
