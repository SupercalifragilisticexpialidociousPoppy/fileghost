package api

import (
	"net/http"
)

// NewRouter sets up all the HTTP routes and binds them to the Server handlers
func NewRouter(srv *Server) *http.ServeMux {
	mux := http.NewServeMux()

	// Utility Routes
	mux.HandleFunc("/ping", srv.HandlePing)
	mux.HandleFunc("/storage", srv.HandleCheckStorage)

	// Auth Routes
	mux.HandleFunc("/register", srv.HandleRegister)
	mux.HandleFunc("/login", srv.HandleLogin)
	mux.HandleFunc("/logout", srv.HandleLogout)
	mux.HandleFunc("/changepassword", srv.HandleChangePassword)

	// File Storage Routes
	mux.HandleFunc("/upload", srv.HandleUpload)
	mux.HandleFunc("/download", srv.HandleDownload)

	return mux
}
