package main

import (
	"fmt"
	"os"

	"byte_crusher/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Byte Crusher - TUI Hex Editor for File Glitching")
		fmt.Println("Usage: ")
		fmt.Println("  go run main.go <file_path>")
		fmt.Println("Example: ")
		fmt.Println("  go run main.go ./test_image.jpg")
		os.Exit(1)
	}

	filePath := os.Args[1]

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Error: File %s does not exist.\n", filePath)
		os.Exit(1)
	}

	err := ui.RunTUI(filePath)
	if err != nil {
		fmt.Printf("Application crashed: %v\n", err)
	}
}
