package tunnel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	tunnelCmd    *exec.Cmd
	tunnelActive bool
	mu           sync.Mutex
)

type discordPayload struct {
	Content string `json:"content"`
}

// StartPinggyTunnel boots the SSH tunnel and broadcasts the URL to Discord
func StartPinggyTunnel(port string) error {
	fmt.Println("[     TUNNEL     ] Tunnel process initiated...")

	mu.Lock()
	if tunnelActive {
		mu.Unlock()
		return nil
	}
	mu.Unlock()
	fmt.Println("[     TUNNEL     ] Mutex pass.")

	webhookURL := os.Getenv("discordWebhook")

	if webhookURL == "" {
		return fmt.Errorf("[     TUNNEL     ] pinggytoken or discordWebhook is missing in .env")
	}
	fmt.Println("[     TUNNEL     ] .env variables extracted.")

	forwardRule := fmt.Sprintf("0:localhost%s", port)
	homeDir, _ := os.UserHomeDir()
	keyPath := filepath.Join(homeDir, ".ssh", "pinggy_key")

	// Build the SSH command exactly as requested
	tunnelCmd = exec.Command(
		"ssh",
		"-T",
		"-p", "443",
		"-i", keyPath,
		"-R", forwardRule,
		"-o", "StrictHostKeyChecking=no",
		"-o", "ServerAliveInterval=30",
		"+text@free.pinggy.io",
	)

	// We only need stdout to read the banner
	stdout, err := tunnelCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("[     TUNNEL     ] Failed to bind stdout pipe: %w", err)
	}
	fmt.Println("[     TUNNEL     ] Stdout pipe established.")

	// tunnelCmd.Stdout = os.Stdout
	// tunnelCmd.Stderr = os.Stderr

	if err := tunnelCmd.Start(); err != nil {
		return fmt.Errorf("[     TUNNEL     ] Failed to start SSH tunnel: %w", err)
	}
	fmt.Println("[     TUNNEL     ] SSH tunnel started successfully.")

	mu.Lock()
	tunnelActive = true
	mu.Unlock()

	// Asynchronous worker to capture the raw banner and send it
	// go func() {
	// 	tunnelCmd.Wait()
	// 	mu.Lock()
	// 	tunnelActive = false
	// 	mu.Unlock()
	// 	fmt.Println("\n[     TUNNEL     ] Pinggy tunnel disconnected.")
	// }()

	go func() {
		defer func() {
			tunnelCmd.Wait()
			mu.Lock()
			tunnelActive = false
			mu.Unlock()
			fmt.Println("\n[     TUNNEL     ] Pinggy tunnel disconnected.")
		}()

		var outputBuilder strings.Builder
		var readMu sync.Mutex

		//1. Start a continuous, non-blocking reader to ingest the stdout stream
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stdout.Read(buf)
				if n > 0 {
					readMu.Lock()
					outputBuilder.Write(buf[:n])
					readMu.Unlock()
				}
				if err != nil {
					break // Pipe closed
				}
			}
		}()

		//2. Wait exactly 4 seconds for Pinggy to finish transmitting the banner
		time.Sleep(15 * time.Second)

		//3. Take a snapshot of everything we read
		readMu.Lock()
		rawOutput := strings.TrimSpace(outputBuilder.String())
		readMu.Unlock()

		// 4. Send the snapshot to Discord
		if rawOutput != "" {
			fmt.Println("[     TUNNEL     ] Output captured. Broadcasting to Discord...")
			go sendDiscordNotification(webhookURL, rawOutput)
		} else {
			fmt.Println("[     TUNNEL     ] Warning: Captured no output from Pinggy.")
		}
	}()

	return nil
}

// sendDiscordNotification dispatches the newly acquired URL to your Discord channel
func sendDiscordNotification(webhookURL, tunnelURL string) {
	msg := discordPayload{
		Content: fmt.Sprintf("**FileGhost Server is Online!**\nTunnel URL: `%s`\nTime: `%s`", tunnelURL, time.Now().Format(time.Kitchen)),
	}

	body, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("[     DISCORD    ] Error formatting webhook JSON: %v\n", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("[     DISCORD    ] Failed to fire webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		fmt.Println("[     DISCORD    ] Broadcast sent to Discord webhook.")
	} else {
		fmt.Printf("[     DISCORD    ] Webhook returned status: %d\n", resp.StatusCode)
	}
}

// StopPinggyTunnel terminates the SSH tunnel process cleanly
func StopPinggyTunnel() {
	mu.Lock()
	defer mu.Unlock()

	if tunnelCmd != nil && tunnelCmd.Process != nil && tunnelActive {
		fmt.Println("[     DISCORD    ] Shutting down Pinggy tunnel...")
		_ = tunnelCmd.Process.Kill()
		tunnelActive = false
	}
}

// IsActive returns the current tunnel state
func IsActive() bool {
	mu.Lock()
	defer mu.Unlock()
	return tunnelActive
}
