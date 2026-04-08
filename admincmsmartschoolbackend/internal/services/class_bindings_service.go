package services

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/models"
)

type BindingReq struct {
	UserIDs  []int `json:"user_ids"`
	ClassIDs []int `json:"class_ids"`
}

func HandleClassTeachers(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	classIDStr := parts[len(parts)-1]

	classID, err := strconv.Atoi(classIDStr)
	if err != nil || classID == 0 {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		getClassTeachers(w, classID)
	case "POST":
		bindTeacherToClass(w, r, classID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleClassTeacherUnbind(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	// /api/v1/class-unbind-teacher/{class_id}/{user_id}
	classIDStr := parts[len(parts)-2]
	userIDStr := parts[len(parts)-1]

	classID, err1 := strconv.Atoi(classIDStr)
	userID, err2 := strconv.Atoi(userIDStr)

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid ID parameters", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`DELETE FROM teacher_classes WHERE class_id = $1 AND user_id = $2 AND is_homeroom = FALSE`, classID, userID)
	if err != nil {
		log.Println("Unbind teacher error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getClassTeachers(w http.ResponseWriter, classID int) {
	query := `
		SELECT u.id, u.name, u.email, COALESCE(u.unit, ''), COALESCE(t.nip, ''), COALESCE(t.qualification, ''), COALESCE(t.status, ''), COALESCE(u.is_active, TRUE)
		FROM users u
		JOIN teacher_classes tc ON u.id = tc.user_id
		LEFT JOIN teachers t ON u.id = t.user_id
		WHERE tc.class_id = $1 AND tc.is_homeroom = FALSE AND COALESCE(u.is_active, TRUE) = TRUE
		ORDER BY u.name ASC
	`
	rows, err := database.DB.Query(query, classID)
	if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "Query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var teachers []models.TeacherDetail
	for rows.Next() {
		var t models.TeacherDetail
		if err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Unit, &t.NIP, &t.Qualification, &t.Status, &t.IsActive); err != nil {
			log.Println("Scan error:", err)
			continue
		}
		teachers = append(teachers, t)
	}

	if len(teachers) == 0 {
		json.NewEncoder(w).Encode([]models.TeacherDetail{})
	} else {
		json.NewEncoder(w).Encode(teachers)
	}
}

func bindTeacherToClass(w http.ResponseWriter, r *http.Request, classID int) {
	var req BindingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if len(req.UserIDs) > 0 {
		for _, uID := range req.UserIDs {
			_, err := database.DB.Exec(`
				INSERT INTO teacher_classes (user_id, class_id, is_homeroom)
				VALUES ($1, $2, FALSE)
				ON CONFLICT (user_id, class_id) DO UPDATE SET is_homeroom = FALSE
			`, uID, classID)
			
			if err != nil {
				log.Println("Bind teacher error:", err)
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}


func HandleClassStudentsBinding(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	classIDStr := parts[len(parts)-1]

	classID, err := strconv.Atoi(classIDStr)
	if err != nil || classID == 0 {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		bindStudentToClass(w, r, classID)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleClassStudentUnbind(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	classIDStr := parts[len(parts)-2]
	userIDStr := parts[len(parts)-1]

	classID, err1 := strconv.Atoi(classIDStr)
	userID, err2 := strconv.Atoi(userIDStr)

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid ID parameters", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`DELETE FROM student_classes WHERE class_id = $1 AND user_id = $2`, classID, userID)
	if err != nil {
		log.Println("Unbind student error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func bindStudentToClass(w http.ResponseWriter, r *http.Request, classID int) {
	var req BindingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var termID int
	database.DB.QueryRow("SELECT id FROM academic_terms ORDER BY id DESC LIMIT 1").Scan(&termID)

	if len(req.UserIDs) > 0 {
		for _, uID := range req.UserIDs {
			_, err := database.DB.Exec(`
				INSERT INTO student_classes (user_id, class_id, academic_term_id)
				VALUES ($1, $2, $3)
				ON CONFLICT (user_id, academic_term_id) 
				DO UPDATE SET class_id = EXCLUDED.class_id
			`, uID, classID, termID)

			if err != nil {
				log.Println("Bind student error:", err)
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
