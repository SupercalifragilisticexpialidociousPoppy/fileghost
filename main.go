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
	eyecon1 := `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢐⠀⠀⠀⠀⠀⠀⡷
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⡇⢠⠀⠀⡀⢠⣿⠀⠀⡀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣘⡁⠀⠀⠀⠀⢀⠀⡀⠀⣷⠈⣧⡀⢀⡿⡿⠀⣰⠁⡀⢀⠀⠀⠀⠀⢠⠄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀	         [                 GHOSTORAG_v0.5                 ]
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡁⢱⠀⠀⠀⢀⠀⣀⢱⣆⣿⡀⢸⣷⣀⣱⣿⣀⡿⢰⢇⡎⠀⠀⠀⠀⢸⠀⠀⠀⠀⠀⢀⡇
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣆⠀⠈⠀⡘⣷⡄⡄⠂⢮⣱⣬⣷⣿⣷⡞⣿⣿⣻⣿⣿⡏⣿⢸⡝⢒⡖⠶⡄⣾⡠⣀⠀⠀⡐⣼⠁
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⢯⡂⠄⢂⣴⢿⣿⡿⡇⣩⣿⢻⣿⣿⣿⣷⣽⣿⣾⣿⣿⣦⣿⣿⢡⣎⣿⢁⣼⣟⠂⣌⣹⢻⣲⣇⢀
⠀⠀⠀⠀⠀⠶⡀⠀⠀⠀⠀⠀⠐⣩⢿⣿⣙⣲⣌⢿⣿⣷⣹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣾⣿⣿⣿⣷⣶⣾⣿⣿⣿⣣⣿⠶⠠⡀
⠀⠀⠀⠀⠀⠀⠹⣔⠄⢐⣶⡞⢫⣑⣦⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣟⣻⣿⣿⣿⣿⣿⣿⣿⣿⣿⡍⡀⠀⠈⠢⢀
⢀⣀⣀⣀⣀⣐⣄⣹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⣷⣦⡀⠀⠀⠁
⠈⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⣿⠟⠉⠁⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⠞⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⢈⡍⠻⣦⣀			 [ USER:`
	fmt.Printf("%v\n", eyecon1)
	if loggedInUser != "" {
		fmt.Printf("⠀⠀⠘⣻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠉⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⠿⠛⠻⠟⠀⠈⠉⠉⢉⣭⣿⣿⣿⣿⣿⣿⣿⣿⠈⠃⠀⠀⠏⣳⡀⠀⠀⠀⠀⠀		        	%s\n", loggedInUser)
	} else {
		fmt.Print("⠀⠀⠘⣻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠉⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⠿⠛⠻⠟⠀⠈⠉⠉⢉⣭⣿⣿⣿⣿⣿⣿⣿⣿⠈⠃⠀⠀⠏⣳⡀⠀⠀⠀⠀⠀		        	<Not logged in yet>\n")
	}
	eyecon2 := `⠀⠀⠀⣩⠼⣟⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠃⠀⠀⠀⠀⠀⠀⠈⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⠀⠀⠀⠀⢴⣿⣾⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⢠⠀⠀⠈⠦⡀								  ]
⠀⠀⠀⠀⡸⢹⡉⢿⣿⣿⡿⡿⠏⠏⠉⢿⡀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⡏⠀⢀⠀⠀⠈⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀⠀⠀⠀⠀⠀⠀⠀⣱⡀
⠀⠀⠀⠀⠀⢙⢧⢻⣥⣚⣽⣱⢢⣆⡠⠈⠙⠂⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣏⣤⣾⣿⣿⣦⣄⡘⢿⣿⣿⣿⣿⣿⣿⣿⠃⠀⠀⠀⠀⠀⠀⠀⠀⣀⣜⡀	 	 [ SERVER URL:
`
	fmt.Printf("%v", eyecon2)
	fmt.Printf("⠀⠀⠀⠀⠀⠈⢧⢺⡷⣿⣟⣿⣯⣯⣖⢇⠀⠀⠉⠢⡀⠀⠀⠀⠀⠀⢹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠇⠀⠀⠀⠀⠀⠀⠠⢂⣎⣰⣿⢷				%s\n", serverURL)
	eyecon4 := `⠀⠀⠀⠀⠀⠀⠈⡝⣿⣿⣿⣿⣿⣾⡿⣯⣉⣤⣀⡀⠀⠑⠀⢀⡀⠀⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠃⠀⠀⢀⡀⠀⣠⠰⢍⣡⡟⣴⣣⣯⡆							  	  ]
⠀⠀⠀⠀⠀⠀⠀⣵⣿⣿⣿⣿⣿⣿⣿⡿⣿⡾⣦⡜⡞⣒⡐⠢⣄⣉⡂⠄⠀⠉⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⢋⠤⠠⠶⠒⠉⠉⠈⠁⠀⠰⠄⠀⠹⢏⠟⠁
⠀⠀⠀⠀⠀⢠⡾⠋⠀⠉⠩⠽⢷⡟⣿⣿⣿⣿⣏⣻⣽⣡⣛⡑⢦⡉⢯⣩⡿⢶⣥⣦⣝⢻⣟⠿⣿⡿⢿⡿⢟⠟⠍⠁⠀⠀⠋⠀⠄⠃
⠀⠀⠀⠀⠔⠋⠀⠀⠀⠀⠄⠀⢊⣾⣿⣽⡿⠿⡿⢟⡿⣷⣿⣿⣾⣿⣲⡴⠏⣲⠄⠤⠍⠣⠌⣻⠈⢠⠀⢀⣀⠄
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⢂⣾⠛⠉⠀⠀⠀⠀⠀⠀⠯⠁⣾⡟⣩⣶⡿⠟⠋⠙⠚⠓⠃⢻⡟⣰⠏⠈⠀⠁
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢽⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣿⡾⠛⠁⠀⠀⠀⠀⠀⠀⠀⠘⣿⣏
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠹⡇					⠁⡇⠉⠘⠟⠦⢏⡿        ⢀⡀⠀⣠⣏⣻⣽⣡⣛⡑⢦⡉⢯⣩⡿⢶⣥⣦⣝⠘⠟⠦⢏⡿       ⢏⡿
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠰⠃							⡟⣿⣳⡀⣤⡏⠟⢉⠰⢍⣡⡟⣴⣣⣯⡆⡟⣿⣳⡀⣤⡏⠟⢉       ⡟⣿⣳⡀⣤⡏⠟⢉
`
	fmt.Printf("%v", eyecon4)
}

func main() {
	printBanner()
	fmt.Print("\n[ RUN:		/help		TO VIEW COMMANDS ]\n")

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

		fmt.Printf("[    GHOSTORAG_   ] Take care of yourself.\n\n")
		os.Exit(0)
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		// Dynamic prompt showing who is logged in
		prompt := "guest"
		if loggedInUser != "" {
			prompt = loggedInUser
		}
		fmt.Printf("\nghostorag_@%s > ", prompt)

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
		case "/neofetch":
			printBanner()
		case "/ping":
			handlePing()
		case "/storage":
			handleStorage()
		case "/crypt":
			handleCrypt(scanner)
		case "/token":
			showtoken()
		case "/myfiles":
			handleMyFiles()
		case "/upload":
			handleUpload(scanner)
		case "/download":
			handleDownload(scanner)
		case "/server":
			handleServer(scanner)
		case "/exit", "exit", "quit":
			if sessionToken == "" {
				fmt.Println("[      CLI      ] Already logged out, safe to exit.")
			} else {
				fmt.Println("[      CLI      ] Logging out...")
				handleLogout()
			}
			fmt.Printf("[   GHOSTORAG_   ] Take care of yourself.\n\n")
			os.Exit(0)
		default:
			fmt.Printf("[      CLI      ] Unknown command: '%s'. Type /help for available commands.\n", command)
		}
	}
}

func printHelp() {
	fmt.Println("\n[ COMMANDS: ]")
	fmt.Println("  /register      		- Create a new account.")
	fmt.Println("  /login         		- Authenticate and begin session.")
	fmt.Println("  /logout        		- Log out and destroy token.")
	fmt.Println("  /token         		- View current token.")
	fmt.Println("  /changepassword		- Take a guess.")
	fmt.Println()
	fmt.Println("  /ping          		- Check if the server is online.")
	fmt.Println("  /storage       		- Check server capacity.")
	fmt.Println("  /myfiles       		- View your uploaded files.")
	fmt.Println("  /upload        		- Upload an encrypted file to the server.")
	fmt.Println("  /download      		- Download your file from the server.")
	fmt.Println()
	fmt.Println("  /crypt         		- Encrypt or decrpyt a file.")
	fmt.Println("  /server				- Set server URL.")
	fmt.Println("  /neofetch			- Funny banner.")
	fmt.Println("  /exit          		- Close the application safely.")
}
