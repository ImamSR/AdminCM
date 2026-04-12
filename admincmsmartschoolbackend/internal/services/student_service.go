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

func HandleStudentStats(w http.ResponseWriter, r *http.Request) {
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
		SELECT LOWER(COALESCE(c.unit, '')), COUNT(DISTINCT u.id)
		FROM users u
		JOIN student_classes sc ON u.id = sc.user_id
		JOIN academic_terms at ON sc.academic_term_id = at.id
		JOIN classes c ON sc.class_id = c.id
		WHERE u.role = 'siswa' AND COALESCE(u.is_active, TRUE) = TRUE 
		  AND at.id = (SELECT id FROM academic_terms WHERE is_active = TRUE ORDER BY id DESC LIMIT 1)
		GROUP BY LOWER(COALESCE(c.unit, ''))
	`)
	if err != nil {
		log.Println("Stats error:", err)
		http.Error(w, "Query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stats []models.StudentStats
	for rows.Next() {
		var s models.StudentStats
		if err := rows.Scan(&s.Unit, &s.Count); err != nil {
			log.Println("Scan error:", err)
			continue
		}
		stats = append(stats, s)
	}

	if len(stats) == 0 {
		json.NewEncoder(w).Encode([]models.StudentStats{})
	} else {
		json.NewEncoder(w).Encode(stats)
	}
}

func HandleStudents(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		getStudents(w, r)
	case "POST":
		createStudent(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HandleStudentByID(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		updateStudent(w, r, id)
	case "DELETE":
		deleteStudent(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getStudents(w http.ResponseWriter, r *http.Request) {
	classIDStr := r.URL.Query().Get("class_id")
	unitStr := r.URL.Query().Get("unit")

	if classIDStr == "" && unitStr == "" {
		http.Error(w, "class_id or unit is required", http.StatusBadRequest)
		return
	}

	var query string
	var rows *sql.Rows
	var err error

	if classIDStr != "" {
		query = `
			SELECT DISTINCT u.id, u.name, u.email, COALESCE(sd.nisn, ''), COALESCE(u.unit, ''), c.id, COALESCE(c.name, ''), COALESCE(u.is_active, TRUE)
			FROM users u
			JOIN student_classes sc ON u.id = sc.user_id
			JOIN academic_terms at ON sc.academic_term_id = at.id
			JOIN classes c ON sc.class_id = c.id
			LEFT JOIN student_details sd ON u.id = sd.user_id
			WHERE u.role = 'siswa' AND c.id = $1 AND COALESCE(u.is_active, TRUE) = TRUE 
			  AND at.id = (SELECT id FROM academic_terms WHERE is_active = TRUE ORDER BY id DESC LIMIT 1)
			ORDER BY u.name ASC
		`
		rows, err = database.DB.Query(query, classIDStr)
	} else {
		query = `
			SELECT DISTINCT u.id, u.name, u.email, COALESCE(sd.nisn, ''), COALESCE(u.unit, ''), 0, '', COALESCE(u.is_active, TRUE)
			FROM users u
			LEFT JOIN student_details sd ON u.id = sd.user_id
			WHERE u.role = 'siswa' AND LOWER(u.unit) = LOWER($1) AND COALESCE(u.is_active, TRUE) = TRUE
			ORDER BY u.name ASC
		`
		rows, err = database.DB.Query(query, unitStr)
	}
	if err != nil {
		log.Println("Query error:", err)
		http.Error(w, "Query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []models.StudentDetail
	for rows.Next() {
		var s models.StudentDetail
		if err := rows.Scan(&s.ID, &s.Name, &s.Email, &s.NISN, &s.Unit, &s.ClassID, &s.ClassName, &s.IsActive); err != nil {
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

func createStudent(w http.ResponseWriter, r *http.Request) {
	var req models.StudentCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if req.ClassID == 0 {
		http.Error(w, "Class ID is required", http.StatusBadRequest)
		return
	}

	var termID int
	err := database.DB.QueryRow("SELECT id FROM academic_terms WHERE is_active = TRUE LIMIT 1").Scan(&termID)
	if err == sql.ErrNoRows {
		err = database.DB.QueryRow("INSERT INTO academic_terms (term_name, year, is_active) VALUES ('Semester 1', '2026/2027', TRUE) RETURNING id").Scan(&termID)
	}
	if err != nil {
		log.Println("Term DB Error:", err)
		http.Error(w, "Sistem belum memiliki tahun ajaran aktif.", http.StatusInternalServerError)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction start error", http.StatusInternalServerError)
		return
	}

	var newID int
	var existingActive bool
	var existingRole string
	var existingUnit string

	err = tx.QueryRow(`SELECT id, COALESCE(is_active, TRUE), role, COALESCE(unit, '') FROM users WHERE email = $1`, req.Email).Scan(&newID, &existingActive, &existingRole, &existingUnit)
	switch err {
	case nil:
		if existingActive {
			tx.Rollback()
			if existingRole == "siswa" {
				if !strings.EqualFold(existingUnit, req.Unit) {
					http.Error(w, "Gagal: Email sudah terdaftar sebagai siswa aktif di unit " + strings.ToUpper(existingUnit) + ".", http.StatusBadRequest)
				} else {
					http.Error(w, "Gagal: Siswa dengan email ini sudah terdaftar dan berstatus aktif di kelas/angkatan lain.", http.StatusBadRequest)
				}
			} else {
				http.Error(w, "Gagal: Email ini sudah digunakan oleh pengguna aktif (" + existingRole + ").", http.StatusBadRequest)
			}
			return
		}
		if existingRole != "siswa" {
			tx.Rollback()
			http.Error(w, "Email ini sudah digunakan oleh akun non-aktif dengan hak akses selain siswa.", http.StatusBadRequest)
			return
		}

		_, err = tx.Exec(`UPDATE users SET is_active = TRUE, name = $1, unit = $2 WHERE id = $3`, req.Name, req.Unit, newID)
		if err != nil {
			tx.Rollback()
			log.Println("Update user for reactivation:", err)
			http.Error(w, "Error reactivating user record", http.StatusInternalServerError)
			return
		}
	case sql.ErrNoRows:
		err = tx.QueryRow(`
			INSERT INTO users (name, email, role, unit, is_active)
			VALUES ($1, $2, 'siswa', $3, TRUE) RETURNING id
		`, req.Name, req.Email, req.Unit).Scan(&newID)

		if err != nil {
			tx.Rollback()
			log.Println("Insert user:", err)
			http.Error(w, "Error creating user record", http.StatusInternalServerError)
			return
		}
	default:
		tx.Rollback()
		log.Println("Check user query:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var nullableNISN interface{}
	if req.NISN != "" {
		nullableNISN = req.NISN
	}

	_, err = tx.Exec(`
		INSERT INTO student_details (user_id, nisn)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET nisn = EXCLUDED.nisn
	`, newID, nullableNISN)

	if err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "NISN sudah digunakan oleh siswa lain.", http.StatusBadRequest)
		} else {
			log.Println("Insert student detail:", err)
			http.Error(w, "Error creating student metadata", http.StatusInternalServerError)
		}
		return
	}

	_, err = tx.Exec(`
		INSERT INTO students (user_id, unit) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET unit = EXCLUDED.unit
	`, newID, req.Unit)

	if err != nil {
		tx.Rollback()
		log.Println("Insert root student mapping:", err)
		http.Error(w, "Error establishing root student mapping", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
		INSERT INTO student_classes (user_id, class_id, academic_term_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, academic_term_id) DO UPDATE SET class_id = EXCLUDED.class_id
	`, newID, req.ClassID, termID)

	if err != nil {
		tx.Rollback()
		log.Println("Insert class junction:", err)
		http.Error(w, "Error joining student to class", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Transaction commit error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": newID, "status": "success"})
}

func updateStudent(w http.ResponseWriter, r *http.Request, id int) {
	var req models.StudentUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction start error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`UPDATE users SET name = $1, email = $2, unit = $3 WHERE id = $4 AND role = 'siswa'`, req.Name, req.Email, req.Unit, id)
	if err != nil {
		tx.Rollback()
		log.Println("Update user:", err)
		http.Error(w, "Server error updating identity", http.StatusInternalServerError)
		return
	}

	var nullableNISN interface{}
	if req.NISN != "" {
		nullableNISN = req.NISN
	}

	_, err = tx.Exec(`
		INSERT INTO student_details (user_id, nisn) 
		VALUES ($1, $2) 
		ON CONFLICT (user_id) DO UPDATE SET nisn = EXCLUDED.nisn
	`, id, nullableNISN)
	if err != nil {
		tx.Rollback()
		log.Println("Upsert details:", err)
		http.Error(w, "Server error updating NISN", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Transaction commit error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func deleteStudent(w http.ResponseWriter, _ *http.Request, id int) {
	_, err := database.DB.Exec(`UPDATE users SET is_active = FALSE WHERE id = $1 AND role = 'siswa'`, id)
	if err != nil {
		log.Println("Delete user:", err)
		http.Error(w, "Delete error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func HandleBulkStudents(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.StudentBulkCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if req.ClassID == 0 {
		http.Error(w, "Class ID is required", http.StatusBadRequest)
		return
	}

	var termID int
	err := database.DB.QueryRow("SELECT id FROM academic_terms WHERE is_active = TRUE LIMIT 1").Scan(&termID)
	if err == sql.ErrNoRows {
		err = database.DB.QueryRow("INSERT INTO academic_terms (term_name, year, is_active) VALUES ('Semester 1', '2026/2027', TRUE) RETURNING id").Scan(&termID)
	}
	if err != nil {
		log.Println("Term DB Error:", err)
		http.Error(w, "Sistem belum memiliki tahun ajaran aktif.", http.StatusInternalServerError)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction start error", http.StatusInternalServerError)
		return
	}

	addedCount := 0
	for _, student := range req.Students {
		if student.Name == "" || student.Email == "" {
			continue
		}

		var newID int
		var existingActive bool
		var existingRole string
		err = tx.QueryRow(`SELECT id, COALESCE(is_active, TRUE), role FROM users WHERE email = $1`, student.Email).Scan(&newID, &existingActive, &existingRole)
		
		if err == nil {
			if existingRole != "siswa" {
				continue
			}
			if !existingActive {
				_, err = tx.Exec(`UPDATE users SET is_active = TRUE, name = $1, unit = $2 WHERE id = $3`, student.Name, req.Unit, newID)
				if err != nil {
					log.Println("Update bulk user err:", err)
					continue
				}
			}
		} else if err == sql.ErrNoRows {
			err = tx.QueryRow(`
				INSERT INTO users (name, email, role, unit, is_active)
				VALUES ($1, $2, 'siswa', $3, TRUE) 
				RETURNING id
			`, student.Name, student.Email, req.Unit).Scan(&newID)
			
			if err != nil {
				log.Println("Insert bulk user err:", err)
				continue
			}
		} else {
			log.Println("Check bulk user exist err:", err)
			continue
		}

		var nullableNISN interface{}
		if student.NISN != "" {
			nullableNISN = student.NISN
		}

		_, err = tx.Exec(`
			INSERT INTO student_details (user_id, nisn)
			VALUES ($1, $2)
			ON CONFLICT (user_id) DO UPDATE SET nisn = EXCLUDED.nisn
		`, newID, nullableNISN)
		if err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "unique constraint") {
				errMsg := "Gagal impor: NISN " + student.NISN + " sudah digunakan oleh siswa lain."
				http.Error(w, errMsg, http.StatusBadRequest)
			} else {
				http.Error(w, "Sistem gagal menyimpan rincian siswa: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		_, err = tx.Exec(`
			INSERT INTO students (user_id, unit) 
			VALUES ($1, $2) 
			ON CONFLICT (user_id) DO UPDATE SET unit = EXCLUDED.unit
		`, newID, req.Unit)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Sistem gagal menyimpan status siswa: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(`
			INSERT INTO student_classes (user_id, class_id, academic_term_id)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING
		`, newID, req.ClassID, termID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Sistem gagal mendaftarkan kelas siswa: "+err.Error(), http.StatusInternalServerError)
			return
		}

		addedCount++
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Transaction commit error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"added":  addedCount,
	})
}

func HandleStudentTransfer(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.StudentTransferReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if req.OldClassID == 0 || req.NewClassID == 0 || len(req.StudentIDs) == 0 {
		http.Error(w, "Missing required criteria", http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction start error", http.StatusInternalServerError)
		return
	}

	transferredCount := 0
	for _, studentId := range req.StudentIDs {
		res, err := tx.Exec(`
			UPDATE student_classes 
			SET class_id = $1 
			WHERE class_id = $2 AND user_id = $3
		`, req.NewClassID, req.OldClassID, studentId)

		if err != nil {
			log.Printf("Transfer err for student %d: %v", studentId, err)
			continue
		}

		rows, _ := res.RowsAffected()
		if rows > 0 {
			transferredCount++
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Transaction commit error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"transferred": transferredCount,
	})
}
