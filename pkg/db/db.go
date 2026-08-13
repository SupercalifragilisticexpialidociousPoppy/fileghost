package db

import (
	"database/sql"
	"encoding/json"
	"fileghost_server/pkg/auth"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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

type FileRecord struct {
	ID         string `json:"id"`
	SizeBytes  int64  `json:"size_bytes"`
	UploadedAt string `json:"uploaded_at"`
}

type MyFilesResponse struct {
	Files         []FileRecord `json:"files"`
	TotalFiles    int          `json:"totalFiles"`
	FilesReceived int          `json:"boobs"`
}

// 16 Gigabytes in Bytes (soft-cap)
const MaxStorageBytes int64 = 16 * 1024 * 1024 * 1024

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
			current_token TEXT UNIQUE DEFAULT NULL,
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

// ProcessUpload streams the encrypted file to disk and logs it to SQLite
func ProcessUpload(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n[      STOR      ] Upload request initiated.")

	// 1. Extract Token from Header
	currentToken := r.Header.Get("X-Session-Token")
	if currentToken == "" {
		fmt.Println("[      STOR      ] Process aborted:")
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		fmt.Println("[      STOR      ]     User token is missing.")
		return
	}
	fmt.Println("[      STOR      ]     User token found.")

	// 2. Verify User and Get Owner ID
	var ownerID int
	err := database.QueryRow("SELECT id FROM users WHERE current_token = ?", currentToken).Scan(&ownerID)
	if err != nil {
		fmt.Println("[   AUTH_CRYPT   ] Process aborted:")
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			fmt.Println("[   AUTH_CRYPT   ]     User token is invalid.")
			return
		}
		fmt.Println("[   AUTH_CRYPT   ]     User token is valid.")
		http.Error(w, "Database error", http.StatusInternalServerError)
		fmt.Println("[    DB__STOR    ]     Database error occurred.")
		return
	}
	fmt.Println("[   AUTH_CRYPT   ] User token is valid.")

	// 3. Soft storage cap.
	var usedBytes int64
	database.QueryRow("SELECT COALESCE(SUM(size_bytes), 0) FROM files").Scan(&usedBytes)
	availableBytes := MaxStorageBytes - usedBytes

	// Pre-flight check: If the HTTP request tells us the file size, check it immediately
	if r.ContentLength > availableBytes {
		fmt.Println("[      STOR      ] Process aborted:")
		http.Error(w, "Upload exceeds server capacity", http.StatusInsufficientStorage)
		fmt.Printf("[      STOR      ]     File exceeds available storage capacity: %v bytes\n", availableBytes)
		return
	}

	// 3. Roll the Token
	newToken, err := auth.GenerateOneTimeToken()
	if err != nil {
		fmt.Println("[   AUTH_CRYPT   ]  New token couldn't be generated. Reusing old token.")
		newToken = currentToken
	} else {
		fmt.Println("[   AUTH_CRYPT   ] New token generated.")
	}

	_, err = database.Exec("UPDATE users SET current_token = ? WHERE id = ?", newToken, ownerID)
	if err != nil {
		fmt.Println("[    DB__STOR    ] Database couldn't be updated with new token. Returning response with old token.")
		newToken = currentToken
	} else {
		fmt.Println("[      STOR      ] Database updated with new token.")
	}

	// Set the new token in the response headers immediately
	w.Header().Set("X-New-Token", newToken)

	// 4. Set up the file on disk
	fileID, _ := generateFileID()
	os.MkdirAll("storage", 0755)
	savePath := filepath.Join("storage", fileID+".enc")

	outFile, err := os.Create(savePath)
	if err != nil {
		fmt.Println("[      STOR      ] Process aborted:")
		http.Error(w, "Failed to create file on server", http.StatusInternalServerError)
		fmt.Println("[      STOR      ]     File couldn't be created on the server.") // Please log in again to generate a new token and try again.") //the new token has been saved on db, can't be returned because this failed.
		return
	}
	defer outFile.Close()

	// 5. Stream the request body directly to the file (RAM efficient!)
	writtenBytes, err := io.Copy(outFile, r.Body)
	if err != nil {
		fmt.Println("[      STOR      ] Process aborted:")
		os.Remove(savePath) //stream failed, delete incomplete file
		http.Error(w, "Failed to stream file data", http.StatusInternalServerError)
		fmt.Println("[      STOR      ]     File couldn't be streamed to the server completely.") // Please log in again to generate a new token and then try again.") // the new token has been saved on db, can't be returned because this failed.
		return
	}

	// 6. Insert metadata into SQLite
	_, err = database.Exec(`
		INSERT INTO files (id, owner_id, file_path, size_bytes) 
		VALUES (?, ?, ?, ?)`,
		fileID, ownerID, savePath, writtenBytes,
	)

	if err != nil {
		fmt.Println("[      STOR      ] Process aborted:")
		// If DB fails, delete the orphaned file to save space
		os.Remove(savePath)
		http.Error(w, "Failed to save file metadata", http.StatusInternalServerError)
		fmt.Println("[      STOR      ]     File metadata couldn't be saved.")
		return
	}

	fmt.Printf("[   STORAGE      ] File %s saved securely. Size: %d bytes\n", fileID, writtenBytes)

	// 7. Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(StandardResponse{
		Status:  "success",
		Message: fmt.Sprintf("File uploaded. ID: %s", fileID),
	})
}

// ProcessDownload fetches the file from disk and streams it back to the client
func ProcessDownload(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n[      STOR      ] Download request initiated.")

	// 1. Extract Token and File ID
	currentToken := r.Header.Get("X-Session-Token")
	fileID := r.URL.Query().Get("id")

	if currentToken == "" || fileID == "" {
		fmt.Println("[      STOR      ] Process aborted:")
		http.Error(w, "Missing token or file ID", http.StatusBadRequest)
		fmt.Println("[      STOR      ]     Token or file_ID is missing from payload.")
		return
	}
	fmt.Println("[      STOR      ] Extracted token and file_ID from request.")

	// 2. Verify User & Token
	var ownerID int
	err := database.QueryRow("SELECT id FROM users WHERE current_token = ?", currentToken).Scan(&ownerID)
	if err != nil {
		fmt.Println("[      STOR      ] Process aborted:")
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			fmt.Println("[      STOR      ]     Token invalid.")
			return
		}
		fmt.Println("[      STOR      ]     Token valid.")
		http.Error(w, "Database error", http.StatusInternalServerError)
		fmt.Println("[      STOR      ]     Database decided to give up.")
		return
	}
	fmt.Println("[      STOR      ] Token verified.")

	// 3. Verify File Ownership and Get Local Path
	var filePath string
	var fileSize int64
	err = database.QueryRow("SELECT file_path, size_bytes FROM files WHERE id = ? AND owner_id = ?", fileID, ownerID).Scan(&filePath, &fileSize)
	if err != nil {
		fmt.Println("[      STOR      ] Process aborted:")
		if err == sql.ErrNoRows {
			http.Error(w, "File not found or access denied", http.StatusNotFound)
			fmt.Println("[      STOR      ]     File not found in database, or you're not authorized to view it.")
			return
		}
		fmt.Println("[      STOR      ]     File found in database. User can access it.")
		http.Error(w, "Database error", http.StatusInternalServerError)
		fmt.Println("[      STOR      ]     Database decided to give up.")
		return
	}
	fmt.Println("[      STOR      ] File located in database. User can access it.")

	// 4. Roll the Token
	newToken, err := auth.GenerateOneTimeToken()
	if err != nil {
		//fmt.Println("[      STOR      ] Process aborted:")
		//http.Error(w, "Failed to generate new token", http.StatusInternalServerError)
		fmt.Println("[      STOR      ] New token generation failed.")
		newToken = currentToken
	} else {
		fmt.Println("[      STOR      ] New token generated.")
	}

	_, err = database.Exec("UPDATE users SET current_token = ? WHERE id = ?", newToken, ownerID)
	if err != nil {
		// fmt.Println("[      STOR      ] Process aborted:")
		// http.Error(w, "Failed to update session", http.StatusInternalServerError)
		fmt.Println("[      STOR      ] Couldn't update new token in the database.")
		newToken = currentToken
	} else {
		fmt.Println("[      STOR      ] Database updated with new token.")
	}

	// Set the new token in the response header
	w.Header().Set("X-New-Token", newToken)
	fmt.Println("[      STOR      ] Added to token to response header.")

	// 5. Open File from Disk
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("[      STOR      ] Process aborted:")
		http.Error(w, "File missing on server storage", http.StatusInternalServerError)
		fmt.Println("[      STOR      ]     File missing or corrupted in the server.")
		return
	}
	fmt.Println("[      STOR      ] File opened.")
	defer file.Close()

	// 6. Set Download Headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.enc\"", fileID))

	// 7. Stream File to Response Body (RAM Efficient!)
	written, err := io.Copy(w, file)
	if err != nil {
		fmt.Println("[      STOR      ] Process aborted:")
		fmt.Printf("[   STORAGE      ] Streaming interrupted: %v\n", err)
		fmt.Println("[      STOR      ]     File streaming interrupted.")
		return
	}

	fmt.Printf("[      STOR      ] File %s successfully streamed (%d bytes).\n", fileID, written)
}

// Helper to generate random file IDs
func generateFileID() (string, error) {
	return auth.GenerateOneTimeToken() // Reusing the secure hex generator for a 64-char file ID
}

// GetStorageStats queries the database for total used space
func GetStorageStats(database *sql.DB) StorageResponse {
	var usedBytes int64

	// COALESCE ensures we get 0 instead of NULL if the table is completely empty
	err := database.QueryRow("SELECT COALESCE(SUM(size_bytes), 0) FROM files").Scan(&usedBytes)
	if err != nil {
		fmt.Println("[      DB        ] Warning: Could not calculate used storage.")
		usedBytes = 0
	}

	available := MaxStorageBytes - usedBytes
	if available < 0 {
		available = 0
	}

	return StorageResponse{
		TotalBytes:     MaxStorageBytes,
		AvailableBytes: available,
	}
}

// ProcessMyFiles returns a JSON array of all files owned by the authenticated user
func ProcessMyFiles(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n[      STOR      ] File list request initiated.")

	// 1. Extract Token
	currentToken := r.Header.Get("X-Session-Token")
	if currentToken == "" {
		fmt.Println("[      STOR      ] Aborting Process:")
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		fmt.Println("[      STOR      ]     Token not found.")
		return
	}
	fmt.Println("[      STOR      ] Token extracted.")

	// 2. Verify User and Get Owner ID
	var ownerID int
	err := database.QueryRow("SELECT id FROM users WHERE current_token = ?", currentToken).Scan(&ownerID)
	if err != nil {
		fmt.Println("[      STOR      ] Aborting Process:")
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			fmt.Println("[      STOR      ]     Token invalid.")
			return
		}
		fmt.Println("[      STOR      ]     Token valid.")
		http.Error(w, "Database error", http.StatusInternalServerError)
		fmt.Println("[      STOR      ]     Internal database error.")
		return
	}
	fmt.Println("[      STOR      ] User ID verified.")

	// 3. Roll the Token (Keeping your architecture consistent)
	fmt.Println("[      STOR      ] Generating new token...")
	newToken, err := auth.GenerateOneTimeToken()
	if err == nil {
		_, err := database.Exec("UPDATE users SET current_token = ? WHERE id = ?", newToken, ownerID)
		if err != nil {
			fmt.Println("[      AUTH      ] Couldn't update database with the new token. Returning previous token.")
			w.Header().Set("X-New-Token", currentToken)
		} else {
			w.Header().Set("X-New-Token", newToken)
		}
	} else {
		fmt.Println("[      STOR      ] Error in token generation. Returning with previous token.")
		w.Header().Set("X-New-Token", currentToken)
	}

	// 4. Query all files for this specific user
	rows, err := database.Query("SELECT id, size_bytes, uploaded_at FROM files WHERE owner_id = ? ORDER BY uploaded_at DESC", ownerID)
	if err != nil {
		fmt.Println("[      STOR      ] Aborting Process:")
		http.Error(w, "Failed to retrieve files", http.StatusInternalServerError)
		fmt.Println("[      STOR      ]     Database error fetching files.")
		return
	}
	defer rows.Close()

	var filesFromDB int = 0
	// 5. Build the variable-sized dynamic slice
	var files []FileRecord
	for rows.Next() {
		var f FileRecord
		// We scan the exact columns requested in the SELECT statement
		if err := rows.Scan(&f.ID, &f.SizeBytes, &f.UploadedAt); err == nil {
			files = append(files, f)
		} else {
			filesFromDB++
		}
	}
	boobs := int(len(files))
	filesFromDB += boobs

	// 6. Ship it! Go automatically marshals the slice into a JSON array.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(MyFilesResponse{
		Files:         files,
		TotalFiles:    filesFromDB,
		FilesReceived: boobs,
	})

	fmt.Printf("[      STOR      ] Returned %d file records to client.\n", len(files))
}
