package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// --- Argon2id Cryptography Helpers ---

const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MB (Safe for a Pi 4)
	argonThreads = 4         // Cores to utilize
	argonKeyLen  = 32
	saltLen      = 16
)

// generateSalt creates a cryptographically secure random salt
func generateSalt() ([]byte, error) {
	b := make([]byte, saltLen)
	_, err := rand.Read(b)
	return b, err
}

// hashPassword combines the salt and hash into a single storable string
func hashPassword(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	// Convert raw bytes to standard text for SQLite storage
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: "encoded_salt.encoded_hash"
	fmt.Println("[   AUTH_CRYPT   ] Password hashed.")
	return b64Salt + "." + b64Hash
}

// verifyPassword safely compares a plaintext password attempt against the database string
func verifyPassword(password string, storedData string) (bool, error) {
	parts := strings.Split(storedData, ".")
	if len(parts) != 2 {
		fmt.Println("[   AUTH_CRYPT   ] Password entry in database corrupted.")
		return false, errors.New("invalid hash format in database")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	// Hash the incoming attempt using the exact same salt
	attemptHash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	// subtle.ConstantTimeCompare prevents attackers from measuring how long the comparison takes
	if subtle.ConstantTimeCompare(attemptHash, storedHash) == 1 {
		fmt.Println("[   AUTH_CRYPT   ] Password matches.")
		return true, nil
	}
	fmt.Println("[   AUTH_CRYPT   ] Password doesn't match.")
	return false, nil
}

// GenerateOneTimeToken creates a 64-character random hex string
func GenerateOneTimeToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
