package services

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"admincmsmartschoolbackend/internal/database"
)

type AcademicTerm struct {
	ID        int    `json:"id"`
	Year      string `json:"year"`
	Semester  string `json:"semester"`
	IsActive  bool   `json:"is_active"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func HandleAcademicTerms(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		getAcademicTerms(w, r)
	case "POST":
		createAcademicTerm(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleAcademicTermByID(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid term ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		updateAcademicTerm(w, r, id)
	case "DELETE":
		deleteAcademicTerm(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleActiveAcademicTerm(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid term ID", http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE academic_terms SET is_active = FALSE")
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to reset active terms", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE academic_terms SET is_active = TRUE WHERE id = $1", id)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to set active term", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Transaction commit error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getAcademicTerms(w http.ResponseWriter, _ *http.Request) {
	query := `
		SELECT id, year, semester, is_active, 
		       COALESCE(TO_CHAR(start_date, 'YYYY-MM-DD'), ''), 
		       COALESCE(TO_CHAR(end_date, 'YYYY-MM-DD'), '')
		FROM academic_terms ORDER BY start_date DESC, id DESC
	`
	rows, err := database.DB.Query(query)
	if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "Query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var terms []AcademicTerm
	for rows.Next() {
		var t AcademicTerm
		if err := rows.Scan(&t.ID, &t.Year, &t.Semester, &t.IsActive, &t.StartDate, &t.EndDate); err != nil {
			log.Println("Scan error:", err)
			continue
		}
		terms = append(terms, t)
	}

	json.NewEncoder(w).Encode(terms)
}

func createAcademicTerm(w http.ResponseWriter, r *http.Request) {
	var req AcademicTerm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var start, end interface{}
	if req.StartDate != "" {
		start = req.StartDate
	}
	if req.EndDate != "" {
		end = req.EndDate
	}

	if req.IsActive {
		database.DB.Exec("UPDATE academic_terms SET is_active = FALSE")
	}

	var newID int
	err := database.DB.QueryRow(`
		INSERT INTO academic_terms (year, semester, start_date, end_date, is_active)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, req.Year, req.Semester, start, end, req.IsActive).Scan(&newID)

	if err != nil {
		log.Println("Insert term error:", err)
		http.Error(w, "Failed to create term", http.StatusInternalServerError)
		return
	}

	req.ID = newID
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func updateAcademicTerm(w http.ResponseWriter, r *http.Request, id int) {
	var req AcademicTerm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var start, end interface{}
	if req.StartDate != "" {
		start = req.StartDate
	}
	if req.EndDate != "" {
		end = req.EndDate
	}

	_, err := database.DB.Exec(`
		UPDATE academic_terms 
		SET year = $1, semester = $2, start_date = $3, end_date = $4
		WHERE id = $5
	`, req.Year, req.Semester, start, end, id)

	if err != nil {
		log.Println("Update term error:", err)
		http.Error(w, "Failed to update term", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func deleteAcademicTerm(w http.ResponseWriter, _ *http.Request, id int) {
	_, err := database.DB.Exec("DELETE FROM academic_terms WHERE id = $1 AND is_active = FALSE", id)
	if err != nil {
		log.Println("Delete term error:", err)
		http.Error(w, "Failed to delete term, or term is currently active", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
