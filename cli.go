package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Matches the server's AuthRequest
type AuthPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Matches the server's StandardResponse
type ServerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
}

type LogoutRequest struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type StorageResponse struct {
	TotalBytes     int64 `json:"total_bytes"`
	AvailableBytes int64 `json:"available_bytes"`
}

const serverURL = "http://localhost:2050"

func handleLogin(scanner *bufio.Scanner) {
	fmt.Println("\n[      CLI      ] Login sequence initiated.")

	fmt.Print("[      CLI      ] Username: ")
	if !scanner.Scan() {
		return
	}
	username := strings.TrimSpace(scanner.Text())

	fmt.Print("[      CLI      ] Password: ")
	if !scanner.Scan() {
		return
	}
	password := strings.TrimSpace(scanner.Text())

	fmt.Printf("[      CLI      ] Attempting login for %s...\n", username)

	// 1. Prepare the JSON payload
	payload := AuthPayload{Username: username, Password: password}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("[      CLI      ] Error formatting JSON request.")
		return
	}
	fmt.Println("[      CLI      ] JSON-ified request constructed.")

	// 2. Send the HTTP POST request
	resp, err := http.Post(serverURL+"/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[      CLI      ] Network error: Could not reach server at %s\n", serverURL)
		return
	}
	fmt.Printf("[      CLI      ] Posted to server at %s\n", serverURL)
	defer resp.Body.Close()

	// 3. Decode the server's response
	var srvResp ServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&srvResp); err != nil {
		fmt.Println("[      CLI      ] Failed to read server response.")
		return
	}
	fmt.Println("[      CLI      ] Server response received and read.")

	// 4. Handle the result
	if resp.StatusCode == http.StatusOK {
		// Save the state in our RAM variables!
		sessionToken = srvResp.Token
		loggedInUser = username
		fmt.Println("[      CLI      ] Login successful. Token secured in memory.")
	} else {
		fmt.Printf("[      CLI      ] Login failed: %s\n", srvResp.Message)
	}
}

func handleRegister(scanner *bufio.Scanner) {
	fmt.Println("\n[      CLI      ] Registration sequence initiated.")

	fmt.Print("[      CLI      ] Username: ")
	if !scanner.Scan() {
		return
	}
	username := strings.TrimSpace(scanner.Text())

	fmt.Print("[      CLI      ] Password: ")
	if !scanner.Scan() {
		return
	}
	password := strings.TrimSpace(scanner.Text())

	fmt.Printf("[      CLI      ] Attempting to register %s...\n", username)

	payload := AuthPayload{Username: username, Password: password}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("[      CLI      ] Error formatting request.")
		return
	}
	fmt.Println("[      CLI      ] JSON request constructed.")

	resp, err := http.Post(serverURL+"/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[      CLI      ] Network error: Could not reach server at %s\n", serverURL)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("[      CLI      ] Posted to server at %s\n", serverURL)

	var srvResp ServerResponse
	json.NewDecoder(resp.Body).Decode(&srvResp)

	if resp.StatusCode == http.StatusCreated {
		fmt.Println("[      CLI      ] Registration successful! You can now /login.")
	} else {
		fmt.Printf("[      CLI      ] Registration failed: %s\n", srvResp.Message)
	}
}

func handlePing() {
	fmt.Println()
	sendTime := time.Now()

	resp, err := http.Get(serverURL + "/ping")
	if err != nil {
		fmt.Printf("[      CLI      ] Network error: Could not reach server at %s\n", serverURL)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("%s", string(body)) // Prints the server's "[ SERVER ] Server pinged." string

	duration := time.Since(sendTime)
	fmt.Printf("[      CLI      ] Round trip time: %v\n", duration)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func handleStorage() {
	resp, err := http.Get(serverURL + "/storage")
	if err != nil {
		fmt.Println("[      CLI      ] Failed to connect to server.")
		return
	}
	defer resp.Body.Close()
	fmt.Println("[      CLI      ] Response recieved from server.")

	// if resp.StatusCode != http.StatusOK {
	// 	rawBody, _ := io.ReadAll(resp.Body)
	// 	fmt.Printf("[      CLI      ] Server returned status %d: %s\n", resp.StatusCode, string(rawBody))
	// 	return
	// }

	var srvResp StorageResponse
	if err := json.NewDecoder(resp.Body).Decode(&srvResp); err != nil {
		fmt.Printf("[      CLI      ] Failed to parse storage data. %v\n", err)
		return
	}
	fmt.Println("[      CLI      ] Response parsed.")

	fmt.Printf("[    STORAGE    ] Total Capacity: %s\n", formatBytes(srvResp.TotalBytes))
	fmt.Printf("[    STORAGE    ] Available:      %s\n", formatBytes(srvResp.AvailableBytes))
}

func handleLogout() {
	if sessionToken == "" {
		fmt.Println("[      CLI      ] You are not currently logged in.")
		return
	}

	payload := LogoutRequest{Username: loggedInUser, Token: sessionToken}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(serverURL+"/logout", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("[      CLI      ] Network error during logout.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		sessionToken = ""
		loggedInUser = ""
		fmt.Println("[      CLI      ] Logged out successfully. Session token destroyed.")
	} else {
		fmt.Println("[      CLI      ] Server rejected logout request. Token may already be invalid.")
		sessionToken = ""
		loggedInUser = ""
		fmt.Println("[      CLI      ] To be safe: Log in again and then log out, this would ensure server flushes all tokens.")
	}
}

func handleChangePassword(scanner *bufio.Scanner) {
	if sessionToken == "" {
		fmt.Println("[      CLI      ] You must be logged in to change your password.")
		return
	}

	fmt.Print("[      CLI      ] Old Password: ")
	if !scanner.Scan() {
		return
	}
	oldPass := strings.TrimSpace(scanner.Text())

	fmt.Print("[      CLI      ] New Password: ")
	if !scanner.Scan() {
		return
	}
	newPass := strings.TrimSpace(scanner.Text())

	payload := ChangePasswordRequest{OldPassword: oldPass, NewPassword: newPass}
	jsonData, _ := json.Marshal(payload)

	// Create a custom HTTP client to attach our Auth header
	req, _ := http.NewRequest("POST", serverURL+"/changepassword", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-Token", sessionToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("[      CLI      ] Network error.")
		return
	}
	defer resp.Body.Close()

	var srvResp ServerResponse
	json.NewDecoder(resp.Body).Decode(&srvResp)

	if resp.StatusCode == http.StatusOK {
		// Capture the newly rolled token from the server's headers!
		sessionToken = resp.Header.Get("X-New-Token")
		fmt.Println("[      CLI      ] Password updated successfully. New session token secured.")
	} else {
		fmt.Printf("[      CLI      ] Failed to update password: %s\n", srvResp.Message)
	}
}

func handleCrypt(scanner *bufio.Scanner) {
	fmt.Println("\n[      CLI      ] Local Cryptography Engine Initiated.")

	fmt.Print("[      CLI      ] Mode (encrypt/decrypt): ")
	if !scanner.Scan() {
		return
	}
	mode := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if mode != "encrypt" && mode != "decrypt" {
		fmt.Println("[      CLI      ] Invalid mode. Aborting.")
		return
	}

	fmt.Print("[      CLI      ] Input file path: ")
	if !scanner.Scan() {
		return
	}
	inPath := strings.TrimSpace(scanner.Text())

	fmt.Print("[      CLI      ] Output file path: ")
	if !scanner.Scan() {
		return
	}
	outPath := strings.TrimSpace(scanner.Text())

	fmt.Print("[      CLI      ] Encryption Password: ")
	if !scanner.Scan() {
		return
	}
	password := strings.TrimSpace(scanner.Text())

	data, err := os.ReadFile(inPath)
	if err != nil {
		fmt.Printf("[      CLI      ] Error reading input file: %v\n", err)
		return
	}

	var result []byte
	switch mode {
	case "encrypt":
		result, err = Encrypt(data, password) // Ensure Encrypt() is available in this package
		if err != nil {
			fmt.Printf("[  ENC -> CLI   ] Encryption error: %v\n", err)
			return
		}
	case "decrypt":
		result, err = Decrypt(data, password) // Ensure Decrypt() is available in this package
		if err != nil {
			fmt.Printf("[  ENC -> CLI   ] Decryption error:\n%v", err)
			return
		}
	}

	if err := os.WriteFile(outPath, result, 0644); err != nil {
		fmt.Printf("[      CLI      ] Error writing output file: %v\n", err)
		return
	}
	fmt.Printf("[      CLI      ] Operation successful. File written to %s.\n", outPath)
}

func showtoken() {
	if sessionToken == "" {
		fmt.Println("\n[      CLI      ] You must be logged in to have a token assigned.")
		return
	}
	fmt.Printf("\n[      CLI      ] Current token is %s\n", sessionToken)
}
