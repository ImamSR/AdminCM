package services

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/models"
)

func setCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
}

func HandleClasses(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		getClasses(w, r)
	case "POST":
		createClass(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleClassStats(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := database.DB.Query(`
		SELECT LOWER(COALESCE(unit, '')), COUNT(*) 
		FROM classes 
		GROUP BY LOWER(COALESCE(unit, ''))
	`)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stats []models.ClassStats
	for rows.Next() {
		var s models.ClassStats
		if err := rows.Scan(&s.Unit, &s.Count); err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		stats = append(stats, s)
	}

	if len(stats) == 0 {
		json.NewEncoder(w).Encode([]models.ClassStats{})
	} else {
		json.NewEncoder(w).Encode(stats)
	}
}

func HandleClassByID(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		updateClass(w, r, id)
	case "DELETE":
		deleteClass(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getClasses(w http.ResponseWriter, r *http.Request) {
	unit := r.URL.Query().Get("unit")
	
	query := `
		SELECT c.id, c.name, COALESCE(c.jenjang, ''), COALESCE(c.unit, ''), c.grade, COALESCE(c.class_name, ''), COALESCE(c.gender, '')
		FROM classes c
	`
	args := []interface{}{}
	if unit != "" {
		query += ` WHERE LOWER(c.unit) = LOWER($1)`
		args = append(args, unit)
	}
	query += ` ORDER BY c.id ASC`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var classes []models.ClassDetail
	for rows.Next() {
		var c models.ClassDetail
		if err := rows.Scan(&c.ID, &c.Name, &c.Level, &c.Unit, &c.Grade, &c.ClassName, &c.Gender); err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		var studentCount, teacherCount int
		database.DB.QueryRow(`
			SELECT COUNT(*) 
			FROM student_classes sc 
			JOIN academic_terms at ON sc.academic_term_id = at.id
			JOIN users u ON sc.user_id = u.id 
			WHERE sc.class_id = $1 AND COALESCE(u.is_active, TRUE) = TRUE 
			  AND at.id = (SELECT id FROM academic_terms WHERE is_active = TRUE ORDER BY id DESC LIMIT 1)
		`, c.ID).Scan(&studentCount)
		
		database.DB.QueryRow(`
			SELECT COUNT(*) 
			FROM teacher_classes tc 
			JOIN academic_terms at ON tc.academic_term_id = at.id
			JOIN users u ON tc.user_id = u.id 
			WHERE tc.class_id = $1 AND COALESCE(u.is_active, TRUE) = TRUE 
			  AND at.id = (SELECT id FROM academic_terms WHERE is_active = TRUE ORDER BY id DESC LIMIT 1)
		`, c.ID).Scan(&teacherCount)
		
		var mainTeacher sql.NullString
		err := database.DB.QueryRow(`
			SELECT STRING_AGG(u.name, ', ')
			FROM teacher_classes tc 
			JOIN academic_terms at ON tc.academic_term_id = at.id
			JOIN users u ON tc.user_id = u.id 
			WHERE tc.class_id = $1 AND tc.is_homeroom = TRUE AND COALESCE(u.is_active, TRUE) = TRUE 
			  AND at.id = (SELECT id FROM academic_terms WHERE is_active = TRUE ORDER BY id DESC LIMIT 1)
		`, c.ID).Scan(&mainTeacher)
		
		if err == nil && mainTeacher.Valid {
			c.Teacher = mainTeacher.String
		} else {
			c.Teacher = ""
		}

		c.StudentCount = studentCount
		c.TeacherCount = teacherCount

		classes = append(classes, c)
	}

	if len(classes) == 0 {
		json.NewEncoder(w).Encode([]models.ClassDetail{})
	} else {
		json.NewEncoder(w).Encode(classes)
	}
}

func createClass(w http.ResponseWriter, r *http.Request) {
	var req models.ClassCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Unit == "" {
		http.Error(w, "Name and Unit are required", http.StatusBadRequest)
		return
	}

	var newID int
	err := database.DB.QueryRow(`
		INSERT INTO classes (name, jenjang, unit, grade, class_name, gender) 
		VALUES ($1, '', $2, $3, '', $4) RETURNING id
	`, req.Name, req.Unit, req.Grade, req.Gender).Scan(&newID)
	
	if err != nil {
		http.Error(w, "Insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var termID int
	errTerm := database.DB.QueryRow("SELECT id FROM academic_terms WHERE is_active = TRUE LIMIT 1").Scan(&termID)
	if errTerm != nil {
		_ = database.DB.QueryRow("INSERT INTO academic_terms (term_name, year, is_active) VALUES ('Semester 1', '2026/2027', TRUE) RETURNING id").Scan(&termID)
	}

	if len(req.TeacherIDs) > 0 {
		for _, tID := range req.TeacherIDs {
			var isValid bool
			database.DB.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM users u JOIN classes c ON c.id = $2 
					WHERE u.id = $1 AND LOWER(c.unit) = ANY(string_to_array(LOWER(u.unit), ','))
				)
			`, tID, newID).Scan(&isValid)
			if !isValid {
				log.Printf("Blocked invalid cross-unit homeroom binding. Teacher %d to Class %d", tID, newID)
				continue
			}

			_, errIns := database.DB.Exec("INSERT INTO teacher_classes (user_id, class_id, is_homeroom, academic_term_id) VALUES ($1, $2, TRUE, $3) ON CONFLICT (user_id, class_id, academic_term_id) DO UPDATE SET is_homeroom = TRUE", tID, newID, termID)
			if errIns != nil {
				http.Error(w, "Insert teacher mapping error: "+errIns.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": newID, "status": "success"})
}

func updateClass(w http.ResponseWriter, r *http.Request, id int) {
	var req models.ClassUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`
		UPDATE classes 
		SET name = $1, grade = $2, gender = $3
		WHERE id = $4
	`, req.Name, req.Grade, req.Gender, id)
	
	if err != nil {
		http.Error(w, "Update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var termID int
	errTerm := database.DB.QueryRow("SELECT id FROM academic_terms WHERE is_active = TRUE LIMIT 1").Scan(&termID)
	if errTerm != nil {
		_ = database.DB.QueryRow("INSERT INTO academic_terms (term_name, year, is_active) VALUES ('Semester 1', '2026/2027', TRUE) RETURNING id").Scan(&termID)
	}

	database.DB.Exec(`DELETE FROM teacher_classes WHERE class_id = $1 AND is_homeroom = TRUE AND academic_term_id = $2`, id, termID)
	
	if len(req.TeacherIDs) > 0 {
		for _, tID := range req.TeacherIDs {
			var isValid bool
			database.DB.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM users u JOIN classes c ON c.id = $2 
					WHERE u.id = $1 AND LOWER(c.unit) = ANY(string_to_array(LOWER(u.unit), ','))
				)
			`, tID, id).Scan(&isValid)
			if !isValid {
				log.Printf("Blocked invalid cross-unit homeroom binding. Teacher %d to Class %d", tID, id)
				continue
			}

			_, errIns := database.DB.Exec("INSERT INTO teacher_classes (user_id, class_id, is_homeroom, academic_term_id) VALUES ($1, $2, TRUE, $3) ON CONFLICT (user_id, class_id, academic_term_id) DO UPDATE SET is_homeroom = TRUE", tID, id, termID)
			if errIns != nil {
				http.Error(w, "Update teacher mapping error: "+errIns.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func deleteClass(w http.ResponseWriter, _ *http.Request, id int) {
	_, err := database.DB.Exec(`DELETE FROM classes WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Delete error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}
