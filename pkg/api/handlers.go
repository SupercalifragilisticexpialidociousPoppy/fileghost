package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"fileghost_server/pkg/auth"
	"fileghost_server/pkg/db"
)

// Server holds dependencies like the database connection for the handlers
type Server struct {
	DB *sql.DB
}

// --- Utility Handlers ---

func (s *Server) HandlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("[     SERVER    ] Server pinged.\n"))
	fmt.Println("[     SERVER    ] Server pinged.")
}

func (s *Server) HandleCheckStorage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delegate logic to the db package
	resp := db.GetStorageStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// --- Auth Handlers ---

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delegate business logic to the auth package
	auth.ProcessRegistration(s.DB, w, r)
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delegate business logic to the auth package
	auth.ProcessLogin(s.DB, w, r)
}

func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delegate business logic to the auth package
	auth.ProcessLogout(s.DB, w, r)
}

func (s *Server) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	auth.ProcessChangePassword(s.DB, w, r)
}

// --- File Storage Handlers ---

func (s *Server) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delegate logic to the db package
	db.ProcessUpload(s.DB, w, r)
}

func (s *Server) HandleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delegate logic to the db package
	db.ProcessDownload(s.DB, w, r)
}
