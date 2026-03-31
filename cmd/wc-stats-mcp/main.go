package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ipolyzos/wc-stats-mcp/internal/mcp"
	"github.com/ipolyzos/wc-stats-mcp/internal/tools"
	"github.com/ipolyzos/wc-stats-mcp/internal/version"
)

func main() {
	for _, a := range os.Args[1:] {
		switch a {
		case "--version", "-version":
			fmt.Println(version.String())
			return
		case "-h", "--help":
			fmt.Fprintln(os.Stderr, "wc-stats-mcp — stdio MCP server.")
			fmt.Fprintln(os.Stderr, "Usage: wc-stats-mcp [--version] [--help]")
			return
		}
	}

	reg := tools.NewRegistry()
	reg.Register("server.ping", tools.PingHandler)

	srv := &mcp.Server{
		Reg: reg,
		Log: log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds),
	}
	if err := srv.ServeStdio(context.Background(), os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "server exit: %v\n", err)
		os.Exit(1)
	}
}
