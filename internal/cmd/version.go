package cmd

import (
	"fmt"
	"os"
	"strings"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func VersionString() string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "dev"
	}

	if strings.TrimSpace(commit) == "" && strings.TrimSpace(date) == "" {
		return v
	}

	if strings.TrimSpace(commit) == "" {
		return fmt.Sprintf("%s (%s)", v, strings.TrimSpace(date))
	}

	if strings.TrimSpace(date) == "" {
		return fmt.Sprintf("%s (%s)", v, strings.TrimSpace(commit))
	}

	return fmt.Sprintf("%s (%s %s)", v, strings.TrimSpace(commit), strings.TrimSpace(date))
}

type VersionCmd struct{}

func (c *VersionCmd) Run() error {
	fmt.Fprintln(os.Stdout, "strawpoll "+VersionString())

	return nil
}
