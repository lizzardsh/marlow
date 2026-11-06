package main

import (
	"fmt"
	"os"

	"marlow/internal/cli"
	"marlow/pkg/banner"
	_ "marlow/pkg/modules"
	"marlow/pkg/terminal"
)

func main() {
	printBanner()
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, terminal.Paint(terminal.Error, "[-] ERROR: "+err.Error()))
		os.Exit(2)
	}
}

func printBanner() {
	fmt.Print(banner.Render())
}
