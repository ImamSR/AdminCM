package services

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"admincmsmartschoolbackend/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/idtoken"
)

type GoogleLoginRequest struct {
	Credential string `json:"credential"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
		Unit  string `json:"unit"`
	} `json:"user"`
}

func HandleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	payload, err := idtoken.Validate(context.Background(), req.Credential, clientID)
	if err != nil {
		http.Error(w, "Invalid Google Token", http.StatusUnauthorized)
		return
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		http.Error(w, "Unable to extract email from token", http.StatusBadRequest)
		return
	}

	name, _ := payload.Claims["name"].(string)

	var user struct {
		ID       int
		Name     string
		Email    string
		Role     string
		Unit     string
		IsActive bool
	}

	err = database.DB.QueryRow(`
		SELECT id, name, email, role, COALESCE(unit, ''), COALESCE(is_active, TRUE)
		FROM admin_users 
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.Unit, &user.IsActive)

	if err != nil {
		http.Error(w, "Error Forbidden", http.StatusForbidden)
		return
	}

	if !user.IsActive {
		http.Error(w, "Account is disabled", http.StatusForbidden)
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		http.Error(w, "Authentication is unavailable: JWT_SECRET not configured on server", http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"unit":  user.Unit,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		Token: tokenString,
	}
	resp.User.ID = user.ID
	resp.User.Name = name 
	resp.User.Email = user.Email
	resp.User.Role = user.Role
	resp.User.Unit = user.Unit

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
