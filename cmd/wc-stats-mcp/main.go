package main

import (
	"fmt"
	"os"

	"wc-stats-mcp/internal/version"
)

func main() {
	for _, a := range os.Args[1:] {
		if a == "--version" || a == "-version" {
			fmt.Println(version.String())
			return
		}
		if a == "-h" || a == "--help" {
			fmt.Fprintln(os.Stderr, "wc-stats-mcp — stdio MCP server (WIP).")
			fmt.Fprintln(os.Stderr, "Usage: wc-stats-mcp [--version] [--help]")
			return
		}
	}
	fmt.Fprintln(os.Stderr, "wc-stats-mcp: MCP server not wired yet (WIP).")
}
