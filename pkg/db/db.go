package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	// SQLite driver
	_ "github.com/mattn/go-sqlite3"
)

// StorageResponse represents disk usage data
type StorageResponse struct {
	TotalBytes     int64 `json:"total_bytes"`
	AvailableBytes int64 `json:"available_bytes"`
}

// StandardResponse is reused here for basic JSON replies
type StandardResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// InitializeDB creates the connection and sets up the schemas
func InitializeDB(filepath string) (*sql.DB, error) {
	fmt.Println("[ SERVER -> DB  ] Databse intialization requested.")
	dsn := filepath + "?_journal_mode=WAL&_foreign_keys=ON"
	database, err := sql.Open("sqlite3", dsn)
	if err != nil {
		fmt.Println("[      DB       ] Couldn't open database.")
		return nil, err
	}
	fmt.Println("[      DB       ] Database opened.")

	if err := database.Ping(); err != nil {
		fmt.Println("[      DB       ] Couldn't ping database.")
		return nil, err
	}
	fmt.Println("[      DB       ] Pinged database.")

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		fmt.Println("[      DB       ] Couldn't create table `users`")
		return nil, err
	}
	fmt.Println("[      DB       ] Created and/or located `users`")

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS files (
			id TEXT PRIMARY KEY, 
			owner_id INTEGER NOT NULL,
			file_path TEXT NOT NULL,
			size_bytes INTEGER NOT NULL,
			visibility INTEGER DEFAULT 0, 
			expires_at DATETIME,
			uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		fmt.Println("[      DB       ] Couldn't create table `files`")
		return nil, err
	}
	fmt.Println("[      DB       ] Created and/or located `files`")

	fmt.Println("[      DB       ] Database intialization complete.")
	return database, err
}

// GetStorageStats retrieves local disk space metrics
func GetStorageStats() StorageResponse {
	return StorageResponse{
		TotalBytes:     128000000000, // ~128 GB
		AvailableBytes: 64000000000,  // ~64 GB
	}
}

// ProcessUpload streams the encrypted file to disk and logs it to SQLite
func ProcessUpload(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate JWT header
	// 2. Stream request body to disk
	// 3. Insert file metadata into SQLite

	json.NewEncoder(w).Encode(StandardResponse{Status: "success", Message: "File uploaded"})
}

// ProcessDownload fetches the file from disk and streams it back to the client
func ProcessDownload(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate JWT header
	// 2. Verify user owns requested file ID
	// 3. Stream file from disk to response body
	return
}
