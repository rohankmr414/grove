package main

import (
	"fmt"
	"os"

	"github.com/rohankmr414/grove/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetBuildInfo(version, commit, date)
	if err := cli.Execute(os.Args[1:]); err != nil {
		switch cli.ExitCode(err) {
		case 0:
			fmt.Fprintln(os.Stdout, err)
			return
		case 2:
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		default:
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	}
}
