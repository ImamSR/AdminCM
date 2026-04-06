package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"admincmsmartschoolbackend/internal/database"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

type Claims struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Unit  string `json:"unit"`
	jwt.RegisteredClaims
}

func getJWTSecret() string {
	return os.Getenv("JWT_SECRET")
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization Header Format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			secret := os.Getenv("JWT_SECRET")
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		if !strings.HasSuffix(claims.Email, "@cendekiamuda.sch.id") && !strings.HasSuffix(claims.Email, "@kibarcm.id") {
			http.Error(w, "Invalid email domain. Only @cendekiamuda.sch.id and @kibarcm.id are allowed", http.StatusForbidden)
			return
		}

		var actualRole string
		err = database.DB.QueryRow("SELECT role FROM admin_users WHERE id = $1", claims.ID).Scan(&actualRole)
		if err != nil {
			http.Error(w, "User not found or database error", http.StatusForbidden)
			return
		}

		if actualRole != claims.Role {
			http.Error(w, "Role mismatch detected. Please re-login.", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
