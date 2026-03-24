package main

import (
	"fmt"
	"io"
	"os"
)

var version = "dev"

func main() {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version":
			fmt.Println("claude-telemetry", version)
			return
		}
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(data) == 0 {
		fmt.Println("⚠ statusline: no input")
		return
	}

	// TODO: parse, render
	fmt.Println("claude-telemetry v2 (stub)")
}
