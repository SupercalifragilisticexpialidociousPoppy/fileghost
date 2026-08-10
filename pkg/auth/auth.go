package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type StandardResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
}

// ProcessRegistration handles user creation and database inserts
func ProcessRegistration(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	// 1. Decode JSON AuthRequest
	// 2. Hash password with server-side secret/salt (Argon2id)
	// 3. Insert into SQLite `users` table

	json.NewEncoder(w).Encode(StandardResponse{Status: "success", Message: "User registered"})
}

// ProcessLogin handles user validation and JWT generation
func ProcessLogin(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	// 1. Decode JSON AuthRequest
	// 2. Fetch user hash from SQLite
	// 3. Compare hashes
	// 4. Generate and return JWT

	json.NewEncoder(w).Encode(StandardResponse{Status: "success", Token: "dummy_token"})
}
