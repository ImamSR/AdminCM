package services

import (
	"encoding/json"
	"net/http"

	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/models"
)

func HandleUnits(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == "GET" {
		getUnits(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getUnits(w http.ResponseWriter, _ *http.Request) {
	rows, err := database.DB.Query(`SELECT id, title, level, accent_bar, created_at FROM units ORDER BY created_at ASC`)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var units []models.Unit
	for rows.Next() {
		var u models.Unit
		if err := rows.Scan(&u.ID, &u.Title, &u.Level, &u.AccentBar, &u.CreatedAt); err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		units = append(units, u)
	}

	if len(units) == 0 {
		json.NewEncoder(w).Encode([]models.Unit{})
	} else {
		json.NewEncoder(w).Encode(units)
	}
}
