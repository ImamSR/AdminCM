package services

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/models"
)

func HandleSubjects(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == "GET" {
		getSubjects(w, r)
	} else if r.Method == "POST" {
		createSubject(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleSubjectByID(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] == "" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	id := parts[4]

	if r.Method == "PUT" {
		updateSubject(w, r, id)
	} else if r.Method == "DELETE" {
		deleteSubject(w, r, id)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getSubjects(w http.ResponseWriter, r *http.Request) {
	unit := r.URL.Query().Get("unit")
	
	var rows *sql.Rows
	var err error

	if unit != "" {
		rows, err = database.DB.Query(`SELECT id, name, code, unit, grade, created_at FROM subjects WHERE unit = $1 ORDER BY name ASC`, unit)
	} else {
		rows, err = database.DB.Query(`SELECT id, name, code, unit, grade, created_at FROM subjects ORDER BY name ASC`)
	}

	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subjects []models.Subject
	for rows.Next() {
		var s models.Subject
		if err := rows.Scan(&s.ID, &s.Name, &s.Code, &s.Unit, &s.Grade, &s.CreatedAt); err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		subjects = append(subjects, s)
	}

	if len(subjects) == 0 {
		json.NewEncoder(w).Encode([]models.Subject{})
	} else {
		json.NewEncoder(w).Encode(subjects)
	}
}

func createSubject(w http.ResponseWriter, r *http.Request) {
	var req models.SubjectCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var newID int
	err := database.DB.QueryRow(
		`INSERT INTO subjects (name, code, unit, grade, created_at) VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP) RETURNING id`,
		req.Name, req.Code, req.Unit, req.Grade,
	).Scan(&newID)

	if err != nil {
		http.Error(w, "Insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Subject created successfully",
		"id":      newID,
	})
}

func updateSubject(w http.ResponseWriter, r *http.Request, id string) {
	var req models.SubjectUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	res, err := database.DB.Exec(
		`UPDATE subjects SET name = $1, code = $2, unit = $3, grade = $4 WHERE id = $5`,
		req.Name, req.Code, req.Unit, req.Grade, id,
	)

	if err != nil {
		http.Error(w, "Update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Subject not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Subject updated successfully",
	})
}

func deleteSubject(w http.ResponseWriter, r *http.Request, id string) {
	res, err := database.DB.Exec(`DELETE FROM subjects WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Subject not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Subject deleted successfully",
	})
}
