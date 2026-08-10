package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// --- Structs ---

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type StandardResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
}

type LogoutRequest struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

// --- Route Handlers ---

func ProcessRegistration(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n[      AUTH      ] Registration request initiated.")

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		fmt.Println("[      AUTH      ]     Request payload is invalid.")
		return
	}
	fmt.Println("[      AUTH      ] Request payload is valid.")

	// 1. Generate Salt and Hash
	salt, err := generateSalt()
	if err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("[   AUTH_CRYPT   ]     Salt generation failed.")
		return
	}
	fmt.Println("[   AUTH_CRYPT   ] Salt generated.")
	secureHashString := hashPassword(req.Password, salt)

	// 2. Insert secure string into database
	_, err = database.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", req.Username, secureHashString)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(StandardResponse{Status: "error", Message: "Username already exists"})
			fmt.Println("[      AUTH      ]     Username already exists.")
			return
		}
		fmt.Println("[      AUTH      ]     Username is unique.")

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(StandardResponse{Status: "error", Message: "Database error occurred"})
		fmt.Println("[      AUTH      ]     Database error occurred.")
		return
	}
	fmt.Println("[      AUTH      ] Username is unique.")
	fmt.Println("[      AUTH      ] Account added to database. Registration successful.")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(StandardResponse{Status: "success", Message: "User registered securely."})
}

func ProcessLogin(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n[      AUTH      ] Login request initiated.")

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		fmt.Println("[      AUTH      ]     Request payload invalid.")
		return
	}
	fmt.Println("[      AUTH      ] Request payload valid.")

	var storedHashString string
	err := database.QueryRow("SELECT password_hash FROM users WHERE username = ?", req.Username).Scan(&storedHashString)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(StandardResponse{Status: "error", Message: "Username doesn't exist."})
			fmt.Println("[      AUTH      ]     Username doesn't exist.")
			return
		}
		fmt.Println("[      AUTH      ]     Username exists.")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(StandardResponse{Status: "error", Message: "Database error occurred"})
		fmt.Println("[      AUTH      ]     Database error occurred.")
		return
	}
	fmt.Println("[      AUTH      ] Username exists.")
	fmt.Println("[      AUTH      ] Account located in database.")

	// 3. Verify incoming password against stored data
	match, err := verifyPassword(req.Password, storedHashString)
	if err != nil || !match {
		fmt.Println("[      AUTH      ] Aborting process:")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(StandardResponse{Status: "error", Message: "We found a username, but the password doesn't match. Ensure both of them are correct."})
		fmt.Println("[      AUTH      ]     Password doesn't match.")
		return
	}

	fmt.Println("[      AUTH      ] Password matches.")

	newToken, err := GenerateOneTimeToken()
	if err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(StandardResponse{Status: "error", Message: "Failed to generate token"})
		fmt.Println("[   AUTH_CRYPT   ]     Token generation failure.")
		return
	}
	fmt.Println("[   AUTH_CRYPT   ] New token generated.")

	_, err = database.Exec("UPDATE users SET current_token = ? WHERE username = ?", newToken, req.Username)
	if err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(StandardResponse{Status: "error", Message: "Database error occurred"})
		fmt.Println("[    AUTH__DB    ]     Token couldn't be updated.")
		return
	}
	fmt.Println("[    AUTH__DB    ] Token updated.")

	fmt.Println("[      AUTH      ] Login successful.")
	json.NewEncoder(w).Encode(StandardResponse{
		Status: "success",
		Token:  newToken,
	})
}

func ProcessLogout(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n[      AUTH      ] Logout request initiated.")
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		fmt.Println("[      AUTH      ]     Request payload is invalid.")
		return
	}
	fmt.Println("[      AUTH      ] Request payload valid.")

	// Wipe the token where the username AND current token match
	result, err := database.Exec("UPDATE users SET current_token = NULL WHERE username = ? AND current_token = ?", req.Username, req.Token)
	//result, err := database.Exec("UPDATE users SET current_token = NULL WHERE username = ?", req.Username)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(StandardResponse{Status: "error", Message: "Database error occurred"})
		fmt.Println("[      AUTH      ]     Database failed to wipe current token.")
		return
	}
	fmt.Println("[      AUTH      ] Database has wiped current token.")

	// Check if any rows were actually updated (prevents fake token logouts)
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		fmt.Println("[      AUTH      ] Aborting process:")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(StandardResponse{Status: "error", Message: "Databse failed to update."})
		fmt.Println("[      AUTH      ]     Database failed secondary check for updates.")
		return
	}

	fmt.Println("[      AUTH      ] Logout successful.")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StandardResponse{Status: "success", Message: "Session terminated securely."})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func ProcessChangePassword(database *sql.DB, w http.ResponseWriter, r *http.Request) {
	fmt.Println("\n[      AUTH      ] Change password request initiated.")

	// 1. Extract Token
	currentToken := r.Header.Get("X-Session-Token")
	if currentToken == "" {
		fmt.Println("[      AUTH      ] Aborting process:")
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		fmt.Println("[      AUTH      ]     Token not found.")
		return
	}
	fmt.Println("[      AUTH      ] Token extracted.")

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		fmt.Println("[      AUTH      ]     Payload invalid.")
		return
	}
	fmt.Println("[      AUTH      ] Payload valid.")

	// 2. Fetch User ID and Stored Hash by Token
	var userID int
	var storedHash string
	err := database.QueryRow("SELECT id, password_hash FROM users WHERE current_token = ?", currentToken).Scan(&userID, &storedHash)
	if err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		if err == sql.ErrNoRows {
			fmt.Println("[      AUTH      ]     Token invalid.")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		fmt.Println("[      AUTH      ]     Token valid.")
		http.Error(w, "Database error", http.StatusInternalServerError)
		fmt.Println("[      AUTH      ]     Database couldn't extract old password.")
		return
	}
	fmt.Println("[      AUTH      ] Old password extracted.")

	// 3. Verify Old Password
	match, err := verifyPassword(req.OldPassword, storedHash)
	if err != nil || !match {
		fmt.Println("[      AUTH      ] Aborting process:")
		fmt.Println("[      AUTH      ]     Old password doesn't match.")
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(StandardResponse{Status: "error", Message: "Old password does not match"})
		return
	}
	fmt.Println("[      AUTH      ] Old password matches.")

	// 4. Hash New Password
	salt, err := generateSalt()
	if err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("[      AUTH      ]     Salt generation failed.")
		return
	}
	fmt.Println("[      AUTH      ] Salt generated.")
	newHashString := hashPassword(req.NewPassword, salt)
	fmt.Println("[      AUTH      ] New password hashed.")

	// 5. Roll Token
	newToken, err := GenerateOneTimeToken()
	if err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		http.Error(w, "Failed to generate new token", http.StatusInternalServerError)
		fmt.Println("[      AUTH      ]     Token generation failed.")
		return
	}
	fmt.Println("[      AUTH      ] New token generated.")

	// 6. Update Password and Token in DB
	_, err = database.Exec("UPDATE users SET password_hash = ?, current_token = ? WHERE id = ?", newHashString, newToken, userID)
	if err != nil {
		fmt.Println("[      AUTH      ] Aborting process:")
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		fmt.Println("[      AUTH      ]     Couldn't update database with the new password.")
		return
	}
	fmt.Println("[      AUTH      ] Database updated.")

	// 7. Send Response
	fmt.Println("[      AUTH      ] Password changed.")
	w.Header().Set("X-New-Token", newToken)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StandardResponse{
		Status:  "success",
		Message: "Password updated successfully.",
	})
}
