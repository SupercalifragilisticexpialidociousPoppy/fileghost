package main

import (
	"flag"
	"fmt"
	"os"
)

// RunCLI handles terminal arguments and coordinates file reading/writing
func RunCLI() {
	mode := flag.String("mode", "", "Operation mode: 'encrypt' or 'decrypt'")
	inPath := flag.String("in", "", "Input file path")
	outPath := flag.String("out", "", "Output file path")
	password := flag.String("pass", "", "Password for key derivation")
	flag.Parse()

	if *mode == "" || *inPath == "" || *outPath == "" || *password == "" {
		fmt.Println("[      CLI      ] Usage:")
		fmt.Println("[      CLI      ]     go run . -mode=encrypt -in=absolute-path/sample.zip -out=absolute-path/sample.enc -pass=secret123")
		fmt.Println("[      CLI      ]     go run . -mode=decrypt -in=absolute-path/sample.enc -out=absolute-path/restored.zip -pass=secret123")
		os.Exit(1)
	}

	data, err := os.ReadFile(*inPath)
	if err != nil {
		fmt.Printf("[      CLI      ] Error reading input file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[      CLI      ] File read successfully from %v.\n", *inPath)

	var result []byte
	switch *mode {
	case "encrypt":
		result, err = Encrypt(data, *password)
		if err != nil {
			fmt.Printf("[  ENC -> CLI   ] Encryption error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[  ENC -> CLI   ] Payload receieved.")
	case "decrypt":
		result, err = Decrypt(data, *password)
		if err != nil {
			fmt.Printf("[  ENC -> CLI   ] Decryption error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[  ENC -> CLI   ] Payload received.")
	default:
		fmt.Println("[      CLI      ] Invalid mode. Use 'encrypt' or 'decrypt'.")
		os.Exit(1)
	}

	if err := os.WriteFile(*outPath, result, 0644); err != nil {
		fmt.Printf("[      CLI      ] Error writing output file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[      CLI      ] File written successfully at %v.\n", *outPath)
}
