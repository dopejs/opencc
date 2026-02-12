package main

import (
	"fmt"
	"os"

	"github.com/dopejs/opencc/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if err.Error() == "cancelled" {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
