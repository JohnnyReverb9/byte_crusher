package main

import (
	"fmt"
	"os"

	"byte_crusher/ui"
)

func main() {
	var filePath string
	if len(os.Args) >= 2 {
		filePath = os.Args[1]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("Error: File %s does not exist.\n", filePath)
			os.Exit(1)
		}
	}

	err := ui.RunTUI(filePath)
	if err != nil {
		fmt.Printf("Application crashed: %v\n", err)
	}
}
