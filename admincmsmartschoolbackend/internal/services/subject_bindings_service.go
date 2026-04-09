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

func HandleSubjectTeachers(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	subjectIDStr := parts[len(parts)-1]

	subjectID, err := strconv.Atoi(subjectIDStr)
	if err != nil || subjectID == 0 {
		http.Error(w, "Invalid subject ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		getSubjectTeachers(w, subjectID)
	case "POST":
		bindTeacherToSubject(w, r, subjectID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleSubjectTeacherUnbind(w http.ResponseWriter, r *http.Request) {
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
	subjectIDStr := parts[len(parts)-2]
	userIDStr := parts[len(parts)-1]

	subjectID, err1 := strconv.Atoi(subjectIDStr)
	userID, err2 := strconv.Atoi(userIDStr)

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid ID parameters", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`DELETE FROM teacher_subjects WHERE subject_id = $1 AND user_id = $2`, subjectID, userID)
	if err != nil {
		log.Println("Unbind teacher from subject error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getSubjectTeachers(w http.ResponseWriter, subjectID int) {
	query := `
		SELECT DISTINCT u.id, u.name, u.email, COALESCE(u.unit, ''), COALESCE(t.nip, ''), COALESCE(t.qualification, ''), COALESCE(t.status, ''), COALESCE(u.is_active, TRUE)
		FROM users u
		JOIN teacher_subjects ts ON u.id = ts.user_id
		LEFT JOIN teachers t ON u.id = t.user_id
		WHERE ts.subject_id = $1 AND COALESCE(u.is_active, TRUE) = TRUE
		ORDER BY u.name ASC
	`
	rows, err := database.DB.Query(query, subjectID)
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

func bindTeacherToSubject(w http.ResponseWriter, r *http.Request, subjectID int) {
	var req BindingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var unit string
	err := database.DB.QueryRow("SELECT unit FROM subjects WHERE id = $1", subjectID).Scan(&unit)
	if err != nil {
		http.Error(w, "Subject not found", http.StatusBadRequest)
		return
	}

	if len(req.UserIDs) > 0 {
		for _, uID := range req.UserIDs {
			_, err := database.DB.Exec(`
				INSERT INTO teacher_subjects (user_id, subject_id, unit)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, uID, subjectID, unit)
			
			if err != nil {
				if !strings.Contains(err.Error(), "unique constraint") {
					log.Println("Bind teacher subject error:", err)
				}
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}


func HandleSubjectStudentsBinding(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	subjectIDStr := parts[len(parts)-1]

	subjectID, err := strconv.Atoi(subjectIDStr)
	if err != nil || subjectID == 0 {
		http.Error(w, "Invalid subject ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		getSubjectStudents(w, subjectID)
	case "POST":
		bindStudentToSubject(w, r, subjectID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleSubjectStudentUnbind(w http.ResponseWriter, r *http.Request) {
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
	subjectIDStr := parts[len(parts)-2]
	userIDStr := parts[len(parts)-1]

	subjectID, err1 := strconv.Atoi(subjectIDStr)
	userID, err2 := strconv.Atoi(userIDStr)

	if err1 != nil || err2 != nil {
		http.Error(w, "Invalid ID parameters", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`DELETE FROM student_subjects WHERE subject_id = $1 AND user_id = $2`, subjectID, userID)
	if err != nil {
		log.Println("Unbind student error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getSubjectStudents(w http.ResponseWriter, subjectID int) {
	query := `
		SELECT DISTINCT u.id, u.name, u.email, COALESCE(sd.nisn, ''), COALESCE(u.unit, ''), c.id, COALESCE(c.name, '')
		FROM users u
		JOIN student_subjects ss ON u.id = ss.user_id
		JOIN academic_terms at ON ss.academic_term_id = at.id
		JOIN classes c ON ss.class_id = c.id
		LEFT JOIN student_details sd ON u.id = sd.user_id
		WHERE ss.subject_id = $1 AND COALESCE(u.is_active, TRUE) = TRUE
		  AND at.id = (SELECT id FROM academic_terms WHERE is_active = TRUE ORDER BY id DESC LIMIT 1)
		ORDER BY COALESCE(c.name, '') ASC, u.name ASC
	`
	rows, err := database.DB.Query(query, subjectID)
	if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "Query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []models.StudentDetail
	for rows.Next() {
		var s models.StudentDetail
		if err := rows.Scan(&s.ID, &s.Name, &s.Email, &s.NISN, &s.Unit, &s.ClassID, &s.ClassName); err != nil {
			log.Println("Scan error:", err)
			continue
		}
		students = append(students, s)
	}

	if len(students) == 0 {
		json.NewEncoder(w).Encode([]models.StudentDetail{})
	} else {
		json.NewEncoder(w).Encode(students)
	}
}

func bindStudentToSubject(w http.ResponseWriter, r *http.Request, subjectID int) {
	var req BindingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var termID int
	err := database.DB.QueryRow("SELECT id FROM academic_terms WHERE is_active = TRUE LIMIT 1").Scan(&termID)
	if err != nil {
		_ = database.DB.QueryRow("INSERT INTO academic_terms (term_name, year, is_active) VALUES ('Semester 1', '2026/2027', TRUE) RETURNING id").Scan(&termID)
	}

	if len(req.UserIDs) > 0 {
		for _, uID := range req.UserIDs {
			var classID int
			errClass := database.DB.QueryRow(`
				SELECT class_id FROM student_classes 
				WHERE user_id = $1 AND academic_term_id = $2
				ORDER BY id DESC LIMIT 1
			`, uID, termID).Scan(&classID)
			
			if errClass != nil {
				errClass = database.DB.QueryRow(`
				SELECT class_id FROM student_classes 
				WHERE user_id = $1
				ORDER BY id DESC LIMIT 1
				`, uID).Scan(&classID)
			}

			if errClass == nil {
				_, err := database.DB.Exec(`
					INSERT INTO student_subjects (user_id, subject_id, class_id, academic_term_id)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT DO NOTHING
				`, uID, subjectID, classID, termID)

				if err != nil && !strings.Contains(err.Error(), "unique constraint") {
					log.Println("Bind student error:", err)
				}
			} else {
				log.Println("Could not resolve class for student:", uID)
			}
		}
	}

	if len(req.ClassIDs) > 0 {
		for _, cID := range req.ClassIDs {
			_, err := database.DB.Exec(`
				INSERT INTO student_subjects (user_id, subject_id, class_id, academic_term_id)
				SELECT user_id, $1, class_id, academic_term_id
				FROM student_classes
				WHERE class_id = $2 AND academic_term_id = $3
				ON CONFLICT DO NOTHING
			`, subjectID, cID, termID)

			if err != nil {
				log.Println("Bulk bind relative class error:", err)
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
