package main

import (
	"central-cyclone/config"
	"fmt"
	"os"
)

func main() {
	importFlag := ""
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "--config" && i+1 < len(args) {
			importFlag = args[i+1]
			i++
		}
	}

	if importFlag != "" {
		// Load config file
		settings, err := config.LoadFromFile(importFlag)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Loaded settings: %+v\n", *settings)

	}
}
