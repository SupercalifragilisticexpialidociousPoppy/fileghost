package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// Cryptographic Parameters
const (
	SaltSize     = 16        // 128-bit salt
	NonceSize    = 12        // 96-bit standard AES-GCM nonce
	KeySize      = 32        // 256-bit AES key
	ArgonTime    = 3         // 3 iterations
	ArgonMem     = 64 * 1024 // 64 MB memory
	ArgonThreads = 4         // 4 parallel threads
)

// DeriveKey generates a 256-bit AES key from a password and salt using Argon2id
func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMem, ArgonThreads, KeySize)
}

// Encrypt takes raw data and a password, producing [Salt][Nonce][Ciphertext+Tag]
func Encrypt(plaintext []byte, password string) ([]byte, error) {
	// 0. Confirmation of command.
	fmt.Println("[      ENC       ] Received payload.")
	fmt.Println("[      ENC       ] Attempting encryption...")

	// 1. Generate random 16-byte salt
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("[      ENC      ]     Failed to generate salt: %w", err)
	}
	fmt.Println("[      ENC      ]     Generated salt.")

	// 2. Derive key using Argon2id
	key := DeriveKey(password, salt)

	// 3. Generate random 12-byte nonce
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("[      ENC      ]     Failed to generate nonce: %w", err)
	}
	fmt.Println("[      ENC      ]     Generated nonce.")

	// 4. Initialize AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("[      ENC      ]     Failed to create AES cipher: %w", err)
	}
	fmt.Println("[      ENC      ]     Created AES cipher.")

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("[      ENC      ]     Failed to create GCM mode: %w", err)
	}
	fmt.Println("[      ENC      ]     Created GCM Code.")

	// 5. Encrypt data (gcm.Seal appends ciphertext to the salt + nonce header)
	headerSize := SaltSize + NonceSize
	out := make([]byte, headerSize, headerSize+len(plaintext)+gcm.Overhead())
	copy(out[0:SaltSize], salt)
	copy(out[SaltSize:headerSize], nonce)

	ciphertext := gcm.Seal(out, nonce, plaintext, nil)
	fmt.Println("[      ENC      ] Returning payload...")
	return ciphertext, nil
}

// Decrypt extracts salt and nonce from header, derives the key, and authenticates/decrypts
func Decrypt(encryptedData []byte, password string) ([]byte, error) {
	fmt.Println("[  CLI -> ENC  ] Received payload.")

	headerSize := SaltSize + NonceSize
	if len(encryptedData) < headerSize+16 { // 16 bytes is GCM min overhead
		return nil, errors.New("[      ENC      ] Encrypted file is too short or corrupted.")
	}
	fmt.Println("[      ENC      ] Attempting decryption...")

	// 1. Extract salt and nonce from header
	salt := encryptedData[0:SaltSize]
	nonce := encryptedData[SaltSize:headerSize]
	ciphertext := encryptedData[headerSize:]
	fmt.Println("[      ENC      ]     Extracted salt and nonce from the ciphertext.")

	// 2. Re-derive key using extracted salt
	key := DeriveKey(password, salt)

	// 3. Initialize AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("[      ENC      ]     Failed to create AES cipher: %w", err)
	}
	fmt.Println("[      ENC      ]     Created AES cipher.")

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("[      ENC      ]     Failed to create GCM mode: %w", err)
	}
	fmt.Println("[      ENC      ]     Created GCM mode.")

	// 4. Decrypt and verify integrity tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("[      ENC      ]     Decryption failed: Message Authentication Code inconsistent. You've either used an incorrect password or the payload is corrupted.")
	}
	fmt.Println("[      ENC      ]     Decryption complete.")

	fmt.Println("[      ENC      ] Returning payload...")
	return plaintext, nil
}
