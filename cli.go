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

var serverURL = "http://localhost:2050/"

func handleLogin(scanner *bufio.Scanner) {
	fmt.Println("[      CLI      ] Login sequence initiated.")

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
	fmt.Println("[      CLI      ] Registration sequence initiated.")

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

	fmt.Printf("[    STORAGE    ] Total Capacity: %d B\n", srvResp.TotalBytes)
	fmt.Printf("[    STORAGE    ] Available:      %d B\n", srvResp.AvailableBytes)
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
	fmt.Println("[      CLI      ] Local Cryptography Engine Initiated.")

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
		fmt.Println("[      CLI      ] You must be logged in to have a token assigned.")
		return
	}
	fmt.Printf("[      CLI      ] Current token is %s\n", sessionToken)
}

func handleUpload(scanner *bufio.Scanner) {
	if sessionToken == "" {
		fmt.Println("[      CLI      ] You must be logged in to upload files.")
		return
	}
	fmt.Println("[      CLI      ] Token found.")

	fmt.Println("[      CLI      ] Ensure you're uploading an encrypted file. The storage server in itself doesn't have any encryption.")
	fmt.Print("[      CLI      ] File path to upload: ")
	if !scanner.Scan() {
		return
	}
	filePath := strings.TrimSpace(scanner.Text())

	// 1. Open the file directly (No os.ReadFile to avoid RAM spikes!)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[      CLI      ] Failed to open file: %v\n", err)
		return
	}
	fmt.Println("[      CLI      ] Located input path.")
	defer file.Close() // Ensure the file handle is released when done

	// 2. Get file stats to populate the Content-Length header
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("[      CLI      ] Failed to read file stats: %v\n", err)
		return
	}
	fmt.Println("[      CLI      ] Read input stats.")

	if fileInfo.IsDir() {
		fmt.Println("[      CLI      ] Cannot upload a directory. Please specify a file.")
		return
	}
	fmt.Println("[      CLI      ] Verified input isn't a directory.")

	fmt.Printf("[      CLI      ] Initiating stream for %s (%s)...\n", fileInfo.Name(), formatBytes(fileInfo.Size()))

	// 3. Create the HTTP Request
	// Passing the 'file' object directly into the request body makes Go stream it efficiently.
	req, err := http.NewRequest("POST", serverURL+"/upload", file)
	if err != nil {
		fmt.Println("[      CLI      ] Failed to build request.")
		return
	}
	fmt.Println("[      CLI      ] Request constructed.")

	// 4. Set required headers
	req.Header.Set("X-Session-Token", sessionToken)
	req.ContentLength = fileInfo.Size() // Triggers the server's pre-flight capacity check
	req.Header.Set("Content-Type", "application/octet-stream")

	// 5. Execute the stream
	fmt.Println("[      CLI      ] Sending request to the server...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("[      CLI      ] Network error during upload.")
		return
	}
	fmt.Println("[      CLI      ] Server response received.")
	defer resp.Body.Close()

	// 6. Parse the Server Response
	if resp.StatusCode != http.StatusCreated {
		rawBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("[      CLI      ] Upload failed. Server returned status %d: %s\n", resp.StatusCode, string(bytes.TrimSpace(rawBody)))

		// Rescue the token! Even if the upload failed (e.g., storage full), the server rolled the token.
		if newToken := resp.Header.Get("X-New-Token"); newToken != "" {
			sessionToken = newToken
			fmt.Println("[      CLI      ] Session token was synced and rolled despite upload failure.")
		} else {
			fmt.Println("[      CLI      ] No new token received. User is advised to log in again.")
		}
		return
	}

	// Success! Capture the new token and the file ID
	fmt.Println("[      CLI      ] Session token was synced and rolled.")
	sessionToken = resp.Header.Get("X-New-Token")

	var srvResp ServerResponse
	json.NewDecoder(resp.Body).Decode(&srvResp)

	fmt.Printf("[      CLI      ] Success! %s\n", srvResp.Message)
	fmt.Println("[      CLI      ] Session token updated securely.")
}

func handleMyFiles() {
	if sessionToken == "" {
		fmt.Println("[      CLI      ] You must be logged in to view your files.")
		return
	}
	fmt.Println("[      CLI      ] Session token exists.")

	// 1. Prepare the GET request
	req, err := http.NewRequest("GET", serverURL+"/myfiles", nil)
	if err != nil {
		fmt.Println("[      CLI      ] Error constructing request.")
		return
	}
	req.Header.Set("X-Session-Token", sessionToken)
	fmt.Println("[      CLI      ] Request constructed.")

	// 2. Fire the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("[      CLI      ] Network error while fetching files.")
		return
	}
	fmt.Println("[      CLI      ] Response received from server.")
	defer resp.Body.Close()

	// 3. Handle Token Rolling
	if newToken := resp.Header.Get("X-New-Token"); newToken != "" {
		sessionToken = newToken
		fmt.Println("[      CLI      ] New token received and updated.")
	} else {
		fmt.Println("[      CLI      ] Couldn't update token. Old token is still active. Consider logging in again.")
	}

	if resp.StatusCode != http.StatusOK {
		rawBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("[      CLI      ] Server rejected request: %s\n", string(bytes.TrimSpace(rawBody)))
		return
	}
	fmt.Println("[      CLI      ] Server response OK.")

	// 4. Decode the dynamic JSON array
	var srvResp MyFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&srvResp); err != nil {
		fmt.Println("[      CLI      ] Failed to parse file list.")
		return
	}
	fmt.Println("[      CLI      ] Parsed server response.")

	// 5. Present the data
	if len(srvResp.Files) == 0 {
		fmt.Println("\n[   MY FILES   ] You have no files stored on the server.")
		return
	}

	fmt.Printf("\n[   %s's FILES   ]\n", loggedInUser)
	fmt.Printf("%-66s | %-12s | %s\n", "FILE ID", "SIZE", "UPLOAD DATE")
	fmt.Println(strings.Repeat("-", 105))

	for _, file := range srvResp.Files {
		// Reusing your formatBytes function for readability
		fmt.Printf("%-66s | %-12s | %s\n", file.ID, formatBytes(file.SizeBytes), file.UploadedAt)
	}

	fmt.Printf("Files found in database:   %d\n", srvResp.TotalFiles)
	fmt.Printf("Files fetched in response: %d\n", srvResp.FilesReceived)
	fmt.Println()
}

func handleDownload(scanner *bufio.Scanner) {
	if sessionToken == "" {
		fmt.Println("[      CLI      ] You must be logged in to download files.")
		return
	}

	fmt.Print("[      CLI      ] File ID to download: ")
	if !scanner.Scan() {
		return
	}
	fileID := strings.TrimSpace(scanner.Text())

	fmt.Print("[      CLI      ] Save as (e.g., downloaded.pdf): ")
	if !scanner.Scan() {
		return
	}
	outputPath := strings.TrimSpace(scanner.Text())

	// 1. Prepare the GET request with the file ID in the query string
	url := fmt.Sprintf("%s/download?id=%s", serverURL, fileID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("[      CLI      ] Error constructing request.")
		return
	}
	req.Header.Set("X-Session-Token", sessionToken)
	fmt.Println("[      CLI      ] Contructed request.")

	fmt.Printf("[      CLI      ] Contacting server for file %s...\n", fileID)

	// 2. Fire the Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("[      CLI      ] Network error during download.")
		return
	}
	fmt.Println("[      CLI      ] Server has responded.")
	defer resp.Body.Close()

	// 3. Roll the Token
	// We extract this immediately. Even if the download fails (e.g., file not found),
	// if the server reached the auth stage, it rolled the token.
	if newToken := resp.Header.Get("X-New-Token"); newToken != "" {
		sessionToken = newToken
		fmt.Println("[      CLI      ] Token updated.")
	} else {
		fmt.Println("[      CLI      ] No token received. User is advised to log in again.")
	}

	// 4. Handle Server Errors
	if resp.StatusCode != http.StatusOK {
		rawBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("[      CLI      ] Download failed. Server returned status %d: %s\n", resp.StatusCode, string(bytes.TrimSpace(rawBody)))
		return
	}
	fmt.Println("[      CLI      ] Server response OK.")

	// 5. Create the local destination file
	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("[      CLI      ] Failed to create local file: %v\n", err)
		return
	}
	fmt.Println("[      CLI      ] Outfile created successfully.")
	defer outFile.Close()

	fmt.Println("[      CLI      ] Downloading stream initiated...")

	// 6. Stream the network response directly to the disk (RAM Efficient)
	written, err := io.Copy(outFile, resp.Body)
	if err != nil {
		fmt.Printf("[      CLI      ] Error during file stream: %v\n", err)
		return
	}
	fmt.Println("[      CLI      ] Download stream concluded.")

	fmt.Printf("[      CLI      ] Download complete! Saved %s to %s.\n", formatBytes(written), outputPath)
}

func handleServer(scanner *bufio.Scanner) {
	fmt.Println("[      CLI      ] Enter server URL. Empty input will default to http://localhost:2050")
	fmt.Print("[      CLI      ] Server URL: ")
	if !scanner.Scan() {
		serverURL = "http://localhost:2050/"
	} else {
		serverURL = strings.TrimSpace(scanner.Text())
	}

	fmt.Printf("[      CLI      ] Server URL set as %s\n", serverURL)
}
