/*
Copyright © 2026 HexaGo Contributors
*/
package main

import (
	"fmt"
	"os"

	"github.com/padiazg/hexago/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
