package services

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/middleware"
	"admincmsmartschoolbackend/internal/models"
)

func HandleTeachers(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		getTeachers(w, r)
	case "POST":
		createTeacher(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleTeacherByID(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		updateTeacher(w, r, id)
	case "DELETE":
		deleteTeacher(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTeachers(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	unitStr := r.URL.Query().Get("unit")
	if claims.Role != "superadmin" {
		unitStr = claims.Unit
	}

	query := `
		SELECT 
			u.id, 
			u.name, 
			u.email, 
			COALESCE(u.unit, ''), 
			COALESCE(t.nip, ''), 
			COALESCE(t.qualification, ''), 
			COALESCE(t.status, ''),
			COALESCE(u.is_active, TRUE)
		FROM users u
		LEFT JOIN teachers t ON u.id = t.user_id
		WHERE u.role IN ('guru', 'wali_kelas') AND COALESCE(u.is_active, TRUE) = TRUE
	`
	
	var rows *sql.Rows
	var err error

	if unitStr != "" {
		query += " AND LOWER($1) = ANY(string_to_array(LOWER(u.unit), ',')) ORDER BY u.name ASC"
		rows, err = database.DB.Query(query, unitStr)
	} else {
		query += " ORDER BY u.name ASC"
		rows, err = database.DB.Query(query)
	}
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var teachers []models.TeacherDetail
	for rows.Next() {
		var t models.TeacherDetail
		if err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Unit, &t.NIP, &t.Qualification, &t.Status, &t.IsActive); err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		teachers = append(teachers, t)
	}

	if len(teachers) == 0 {
		json.NewEncoder(w).Encode([]models.TeacherDetail{})
	} else {
		json.NewEncoder(w).Encode(teachers)
	}
}

func createTeacher(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.TeacherCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if claims.Role != "superadmin" {
		req.Unit = claims.Unit
	}

	if req.Name == "" || req.Email == "" {
		http.Error(w, "Name and Email are required", http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var newUserID int
	err = tx.QueryRow(`
		INSERT INTO users (email, name, role, unit, is_active) 
		VALUES ($1, $2, $3, $4, TRUE) RETURNING id
	`, req.Email, req.Name, "guru", req.Unit).Scan(&newUserID)
	
	if err != nil {
		tx.Rollback()
		http.Error(w, "User Insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
		INSERT INTO teachers (user_id, qualification, status, nip) 
		VALUES ($1, $2, $3, $4)
	`, newUserID, req.Qualification, req.Status, req.NIP)
	
	if err != nil {
		tx.Rollback()
		http.Error(w, "Teacher Insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Commit error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": newUserID, "status": "success"})
}

func updateTeacher(w http.ResponseWriter, r *http.Request, id int) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.TeacherUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if claims.Role != "superadmin" {
		req.Unit = claims.Unit
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var res sql.Result
	if claims.Role != "superadmin" {
		res, err = tx.Exec(`
			UPDATE users 
			SET name = $1, email = $2, unit = $3 
			WHERE id = $4 AND role IN ('guru', 'wali_kelas') AND LOWER($5) = ANY(string_to_array(LOWER(unit), ','))
		`, req.Name, req.Email, req.Unit, id, claims.Unit)
	} else {
		res, err = tx.Exec(`
			UPDATE users 
			SET name = $1, email = $2, unit = $3 
			WHERE id = $4 AND role IN ('guru', 'wali_kelas')
		`, req.Name, req.Email, req.Unit, id)
	}
	
	if err != nil {
		tx.Rollback()
		http.Error(w, "User update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		http.Error(w, "Teacher not found or unauthorized", http.StatusNotFound)
		return
	}

	res, err = tx.Exec(`
		UPDATE teachers 
		SET qualification = $1, status = $2, nip = $3 
		WHERE user_id = $4
	`, req.Qualification, req.Status, req.NIP, id)
	
	if err != nil {
		tx.Rollback()
		http.Error(w, "Teacher update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tRows, _ := res.RowsAffected()
	if tRows == 0 {
		_, err = tx.Exec(`
			INSERT INTO teachers (user_id, qualification, status, nip) 
			VALUES ($1, $2, $3, $4)
		`, id, req.Qualification, req.Status, req.NIP)
		
		if err != nil {
			tx.Rollback()
			http.Error(w, "Teacher Insert missing error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Commit error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func deleteTeacher(w http.ResponseWriter, r *http.Request, id int) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var res sql.Result
	var err error

	if claims.Role != "superadmin" {
		res, err = database.DB.Exec(`
			UPDATE users 
			SET is_active = FALSE 
			WHERE id = $1 AND role IN ('guru', 'wali_kelas') AND LOWER($2) = ANY(string_to_array(LOWER(unit), ','))
		`, id, claims.Unit)
	} else {
		res, err = database.DB.Exec(`
			UPDATE users 
			SET is_active = FALSE 
			WHERE id = $1 AND role IN ('guru', 'wali_kelas')
		`, id)
	}
	
	if err != nil {
		http.Error(w, "Delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Teacher not found or unauthorized", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}
