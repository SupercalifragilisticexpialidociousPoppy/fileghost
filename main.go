package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var sessionToken string
var loggedInUser string

func printBanner() {
	eyecon := `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢐⠀⠀⠀⠀⠀⠀⡷⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀	
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⡇⢠⠀⠀⡀⢠⣿⠀⠀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣘⡁⠀⠀⠀⠀⢀⠀⡀⠀⣷⠈⣧⡀⢀⡿⡿⠀⣰⠁⡀⢀⠀⠀⠀⠀⢠⠄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀	         [                FILEGHOST  v_0.1                ]
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡁⢱⠀⠀⠀⢀⠀⣀⢱⣆⣿⡀⢸⣷⣀⣱⣿⣀⡿⢰⢇⡎⠀⠀⠀⠀⢸⠀⠀⠀⠀⠀⢀⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀			   
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣆⠀⠈⠀⡘⣷⡄⡄⠂⢮⣱⣬⣷⣿⣷⡞⣿⣿⣻⣿⣿⡏⣿⢸⡝⢒⡖⠶⡄⣾⡠⣀⠀⠀⡐⣼⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⢯⡂⠄⢂⣴⢿⣿⡿⡇⣩⣿⢻⣿⣿⣿⣷⣽⣿⣾⣿⣿⣦⣿⣿⢡⣎⣿⢁⣼⣟⠂⣌⣹⢻⣲⣇⢀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠶⡀⠀⠀⠀⠀⠀⠐⣩⢿⣿⣙⣲⣌⢿⣿⣷⣹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣾⣿⣿⣿⣷⣶⣾⣿⣿⣿⣣⣿⠶⠠⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀	         
⠀⠀⠀⠀⠀⠀⠹⣔⠄⢐⣶⡞⢫⣑⣦⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣟⣻⣿⣿⣿⣿⣿⣿⣿⣿⣿⡍⡀⠀⠈⠢⢀⠀⠀⠀⠀⠀⠀⠀	         
⢀⣀⣀⣀⣀⣐⣄⣹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⣷⣦⡀⠀⠀⠁⠀⠀⠀⠀⠀⠀
⠈⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⣿⠟⠉⠁⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⠞⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⢈⡍⠻⣦⣀⠀⠀⠀⠀⠀⠀⠀           [ COMMANDS:                                      ]
⠀⠀⠘⣻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠉⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⠿⠛⠻⠟⠀⠈⠉⠉⢉⣭⣿⣿⣿⣿⣿⣿⣿⣿⠈⠃⠀⠀⠏⣳⡀⠀⠀⠀⠀⠀		 
⠀⠀⠀⣩⠼⣟⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠃⠀⠀⠀⠀⠀⠀⠈⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠀⠀⠀⠀⢴⣿⣾⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⢠⠀⠀⠈⠦⡀⠀⠀⠀	         [    /register                                   ]
⠀⠀⠀⠀⡸⢹⡉⢿⣿⣿⡿⡿⠏⠏⠉⢿⡀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⡏⠀⢀⠀⠀⠈⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀⠀⠀⠀⠀⠀⠀⠀⣱⡀⠀⠀		 [    /login                                      ]
⠀⠀⠀⠀⠀⢙⢧⢻⣥⣚⣽⣱⢢⣆⡠⠈⠙⠂⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣏⣤⣾⣿⣿⣦⣄⡘⢿⣿⣿⣿⣿⣿⣿⣿⠃⠀⠀⠀⠀⠀⠀⠀⠀⣀⣜⡀⠀		 [    /logout                                     ]
⠀⠀⠀⠀⠀⠈⢧⢺⡷⣿⣟⣿⣯⣯⣖⢇⠀⠀⠉⠢⡀⠀⠀⠀⠀⠀⢹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠇⠀⠀⠀⠀⠀⠀⠠⢂⣎⣰⣿⢷⠀		 [    /changepassword                             ]
⠀⠀⠀⠀⠀⠀⠈⡝⣿⣿⣿⣿⣿⣾⡿⣯⣉⣤⣀⡀⠀⠑⠀⢀⡀⠀⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠃⠀⠀⢀⡀⠀⣠⠰⢍⣡⡟⣴⣣⣯⡆
⠀⠀⠀⠀⠀⠀⠀⣵⣿⣿⣿⣿⣿⣿⣿⡿⣿⡾⣦⡜⡞⣒⡐⠢⣄⣉⡂⠄⠀⠉⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢋⠤⠠⠶⠒⠉⠉⠈⠁⠀⠰⠄⠀⠹⢏⠟⠁		 [    /ping                                       ]
⠀⠀⠀⠀⠀⢠⡾⠋⠀⠉⠩⠽⢷⡟⣿⣿⣿⣿⣏⣻⣽⣡⣛⡑⢦⡉⢯⣩⡿⢶⣥⣦⣝⢻⣟⠿⣿⡿⢿⡿⢟⠟⠍⠁⠀⠀⠋⠀⠄⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀		 [    /storage                                    ]
⠀⠀⠀⠀⠔⠋⠀⠀⠀⠀⠄⠀⢊⣾⣿⣽⡿⠿⡿⢟⡿⣷⣿⣿⣾⣿⣲⡴⠏⣲⠄⠤⠍⠣⠌⣻⠈⢠⠀⢀⣀⠄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀		 [    /myfiles                                    ]
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⢂⣾⠛⠉⠀⠀⠀⠀⠀⠀⠯⠁⣾⡟⣩⣶⡿⠟⠋⠙⠚⠓⠃⢻⡟⣰⠏⠈⠀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀		 [    /crypt                                      ]
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢽⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣿⡾⠛⠁⠀⠀⠀⠀⠀⠀⠀⠘⣿⣏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀		 [    /upload                                     ]
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠹⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀		 [    /download                                   ]
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠰⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
`

	fmt.Printf("%v\n", eyecon)
}

func main() {
	printBanner()

	// --- OS Signal Trap ---
	sigChan := make(chan os.Signal, 1)
	// Catch Ctrl+C (SIGINT) and standard kill commands (SIGTERM)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan // This goroutine sleeps here until a signal is received
		fmt.Println("\n[      CLI      ] Keyboard interrupt detected.")

		// If a user is logged in, securely terminate the session on the server
		if sessionToken != "" {
			fmt.Println("[      CLI      ] Logging out...")
			handleLogout()
		}

		fmt.Printf("[    FILEGHOST   ] Take care of yourself.\n\n")
		os.Exit(0)
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		// Dynamic prompt showing who is logged in
		prompt := "guest"
		if loggedInUser != "" {
			prompt = loggedInUser
		}
		fmt.Printf("\nfileghost@%s > ", prompt)

		// Wait for user input
		if !scanner.Scan() {
			break
		}

		// Clean the input
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Parse the command and arguments
		parts := strings.Split(input, " ")
		command := parts[0]

		// 4. Command Router
		switch command {
		case "/help":
			printHelp()
		case "/login":
			handleLogin(scanner)
		case "/register":
			handleRegister(scanner)
		case "/changepassword":
			handleChangePassword(scanner)
		case "/logout":
			handleLogout()
		case "/ping":
			handlePing()
		case "/storage":
			handleStorage()
		case "/crypt":
			handleCrypt(scanner)
		case "/token":
			showtoken()
		case "/exit", "exit", "quit":
			if sessionToken == "" {
				fmt.Println("[      CLI      ] Already logged out, safe to exit.")
			} else {
				fmt.Println("[      CLI      ] Logging out...")
				handleLogout()
			}
			fmt.Printf("[   FILEGHOST   ] Take care of yourself.\n\n")
			os.Exit(0)
		default:
			fmt.Printf("[      CLI      ] Unknown command: '%s'. Type /help for available commands.\n", command)
		}
	}
}

func printHelp() {
	fmt.Println("\n[ COMMANDS: ]")
	fmt.Println("  /register      		- Create a new account")
	fmt.Println("  /login         		- Authenticate and begin session")
	fmt.Println("  /logout        		- Log out and destroy token.")
	fmt.Println("  /token         		- View current token.")
	fmt.Println("  /changepassword		- Change your account password.")
	fmt.Println()
	fmt.Println("  /ping          		- Check if server is online.")
	fmt.Println("  /storage       		- Check server capacity.")
	fmt.Println("  /myfiles       		- View your uploaded files.")
	fmt.Println("  /upload        		- Upload an encrypted file to the server.")
	fmt.Println("  /download      		- Download your file from the server.")
	fmt.Println("  /crypt         		- Encrypt or decrpyt a file.")
	fmt.Println()
	fmt.Println("  /exit          		- Close the application safely.")
}
