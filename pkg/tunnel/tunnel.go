package tunnel

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
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

	pinggyToken := os.Getenv("pinggytoken")
	webhookURL := os.Getenv("discordWebhook")

	if pinggyToken == "" || webhookURL == "" {
		return fmt.Errorf("[     TUNNEL     ] pinggytoken or discordWebhook is missing in .env")
	}
	fmt.Println("[     TUNNEL     ] .env variables extracted.")

	target := fmt.Sprintf("%s@free.pinggy.io", strings.TrimSpace(pinggyToken))
	forwardRule := fmt.Sprintf("0:localhost%s", port)

	// Build the SSH command directly without an external bash script
	tunnelCmd = exec.Command(
		"ssh",
		"-tt",
		//"-T",
		"-p", "443",
		"-R", forwardRule,
		"-o", "StrictHostKeyChecking=no",
		"-o", "ServerAliveInterval=30",
		"-o", "PubkeyAuthentication=no",
		target,
	)

	// Pipe it.
	stdout, err := tunnelCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("[     TUNNEL     ] Failed to bind stdout pipe: %w", err)
	}

	stdin, err := tunnelCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("[     TUNNEL     ] Failed to bind stdin pipe: %w", err)
	}
	fmt.Println("[     TUNNEL     ] I/O pipes established.")

	//scanner := bufio.NewScanner(stdout)

	if err := tunnelCmd.Start(); err != nil {
		return fmt.Errorf("[     TUNNEL     ] Failed to start SSH tunnel: %w", err)
	}
	fmt.Println("[     TUNNEL     ] SSH tunnel started successfully.")

	// scanner.Scan()
	// stdin.Write([]byte("\r\n"))

	// go func() {
	// 	time.Sleep(10 * time.Second) // Wait for the Pinggy prompt to appear
	// 	fmt.Println("[     TUNNEL     ] Simulating 'Enter' keypress...")
	// 	stdin.Write([]byte("\r\n")) // Press Enter!
	// }()

	mu.Lock()
	tunnelActive = true
	mu.Unlock()

	// Asynchronous worker to parse output and alert Discord
	// Asynchronous worker to parse output and alert Discord
	go func() {
		// defer func() {
		// 	tunnelCmd.Wait()
		// 	mu.Lock()
		// 	tunnelActive = false
		// 	mu.Unlock()
		// 	fmt.Println("\n[     TUNNEL     ] Pinggy tunnel disconnected.")
		// }()

		// --- STAGE 1: The Byte-by-Byte "Prompt Hunter" ---
		// We read 1 byte at a time so we don't get stuck waiting for a newline
		buf := make([]byte, 1)
		var promptBuffer string

		fmt.Println("[     TUNNEL     ] Hunting for password prompt...")
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				fmt.Printf("Reading terminated: err, %v\n", err)
				break
			}
			if n == 0 {
				fmt.Println("Reading terminated: n == 0")
				break
			}

			char := string(buf[0])
			promptBuffer += char

			// Optional: Print the raw characters to your terminal so you can watch it live
			fmt.Print(promptBuffer)

			// The moment we see "password: " at the end of our buffer, we strike.
			// if strings.HasSuffix(strings.ToLower(promptBuffer), "password: ") {
			// 	fmt.Println("\n[     TUNNEL     ] Prompt detected! Firing Enter key...")

			// 	// Send the virtual keystroke to bypass the prompt
			// 	stdin.Write([]byte("\n"))

			// 	// Break out of the byte-reader loop so we can move to Stage 2
			// 	break
			// }
		}
		fmt.Println("Attempting to hit enter.")
		stdin.Write([]byte("\n"))
		fmt.Println("Hit enter.")

		// --- STAGE 2: The Line-by-Line "URL Scraper" ---
		// Now that the prompt is cleared, Pinggy dumps the text banner with newlines.
		// We can safely hand the stdout pipe directly into a standard scanner.
		scanner := bufio.NewScanner(stdout)
		urlRegex := regexp.MustCompile(`https://[a-zA-Z0-9.-]+\.pinggy\.[a-zA-Z]+`)
		foundURL := false

		fmt.Println("[     TUNNEL     ] Scanner activated. Digesting Pinggy banner...")
		for scanner.Scan() {
			line := scanner.Text()

			// Optional: Print the banner lines to the terminal
			fmt.Println(line)

			// Check for URL only if we haven't found and broadcasted it yet
			if !foundURL {
				if match := urlRegex.FindString(line); match != "" {
					foundURL = true
					fmt.Printf("[     TUNNEL     ] Public URL acquired: %s\n", match)

					if webhookURL != "" {
						go sendDiscordNotification(webhookURL, match)
					}
				}
			}
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
