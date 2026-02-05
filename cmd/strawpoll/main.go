package main

import (
	"os"
	_ "time/tzdata" // Embed timezone database for portable time.LoadLocation

	"github.com/dedene/strawpoll-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		os.Exit(cmd.ExitCode(err))
	}
}
