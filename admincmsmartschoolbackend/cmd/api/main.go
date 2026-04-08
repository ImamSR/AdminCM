package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"admincmsmartschoolbackend/internal/database"
	"admincmsmartschoolbackend/internal/middleware"
	"admincmsmartschoolbackend/internal/services"
)

func main() {
	database.Connect()

	http.HandleFunc("/api/v1/auth/google", services.HandleGoogleAuth)
	
	http.HandleFunc("/api/v1/dashboard/stats", middleware.RequireAuth(services.HandleDashboardStats))

	http.HandleFunc("/api/v1/admins", middleware.RequireAuth(services.HandleAdmins))
	http.HandleFunc("/api/v1/admins/", middleware.RequireAuth(services.HandleAdminByID))
	http.HandleFunc("/api/v1/banners", middleware.RequireAuth(services.HandleBanners))
	http.HandleFunc("/api/v1/banners/", middleware.RequireAuth(services.HandleBannerByID))
	
	http.HandleFunc("/api/v1/academic-terms", middleware.RequireAuth(services.HandleAcademicTerms))
	http.HandleFunc("/api/v1/academic-terms/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/active") {
			middleware.RequireAuth(services.HandleActiveAcademicTerm)(w, r)
			return
		}
		middleware.RequireAuth(services.HandleAcademicTermByID)(w, r)
	})

	http.HandleFunc("/api/v1/subjects", middleware.RequireAuth(services.HandleSubjects))
	http.HandleFunc("/api/v1/subjects/", middleware.RequireAuth(services.HandleSubjectByID))
	http.HandleFunc("/api/v1/units", middleware.RequireAuth(services.HandleUnits))
	http.HandleFunc("/api/v1/classes/stats", middleware.RequireAuth(services.HandleClassStats))
	http.HandleFunc("/api/v1/classes", middleware.RequireAuth(services.HandleClasses))
	http.HandleFunc("/api/v1/classes/", middleware.RequireAuth(services.HandleClassByID))
	http.HandleFunc("/api/v1/teachers", middleware.RequireAuth(services.HandleTeachers))
	http.HandleFunc("/api/v1/teachers/", middleware.RequireAuth(services.HandleTeacherByID))
	http.HandleFunc("/api/v1/students/stats", middleware.RequireAuth(services.HandleStudentStats))
	http.HandleFunc("/api/v1/students/bulk", middleware.RequireAuth(services.HandleBulkStudents))
	http.HandleFunc("/api/v1/students/transfer", middleware.RequireAuth(services.HandleStudentTransfer))
	http.HandleFunc("/api/v1/students", middleware.RequireAuth(services.HandleStudents))
	http.HandleFunc("/api/v1/students/", middleware.RequireAuth(services.HandleStudentByID))

	http.HandleFunc("/api/v1/class-teachers/", middleware.RequireAuth(services.HandleClassTeachers))
	http.HandleFunc("/api/v1/class-unbind-teacher/", middleware.RequireAuth(services.HandleClassTeacherUnbind))
	http.HandleFunc("/api/v1/class-students/", middleware.RequireAuth(services.HandleClassStudentsBinding))
	http.HandleFunc("/api/v1/class-unbind-student/", middleware.RequireAuth(services.HandleClassStudentUnbind))

	http.HandleFunc("/api/v1/subject-teachers/", middleware.RequireAuth(services.HandleSubjectTeachers))
	http.HandleFunc("/api/v1/subject-unbind-teacher/", middleware.RequireAuth(services.HandleSubjectTeacherUnbind))
	http.HandleFunc("/api/v1/subject-students/", middleware.RequireAuth(services.HandleSubjectStudentsBinding))
	http.HandleFunc("/api/v1/subject-unbind-student/", middleware.RequireAuth(services.HandleSubjectStudentUnbind))

	port := os.Getenv("PORT")

	fmt.Printf("Backend HTTP Server starting on :%s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
