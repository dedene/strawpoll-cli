package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

// RootFlags are global flags available to all commands.
type RootFlags struct {
	JSON    bool `help:"Output JSON to stdout" short:"j"`
	Plain   bool `help:"Output plain TSV (for scripting)"`
	NoColor bool `help:"Disable colors" env:"NO_COLOR"`
	Copy    bool `help:"Copy poll URL to clipboard"`
	Open    bool `help:"Open poll URL in browser"`
}

// CLI is the top-level Kong CLI struct.
type CLI struct {
	RootFlags `embed:""`

	Version    kong.VersionFlag `help:"Print version and exit"`
	VersionCmd VersionCmd       `cmd:"" name:"version" help:"Show version information"`
	Auth       AuthCmd          `cmd:"" help:"Manage API key"`
	Config     ConfigCmd        `cmd:"" help:"Manage configuration"`
	Poll       PollCmd          `cmd:"" help:"Poll commands"`
	Meeting    MeetingCmd       `cmd:"" help:"Meeting poll commands"`
	Ranking    RankingCmd       `cmd:"" help:"Ranking poll commands"`
	Completion CompletionCmd    `cmd:"" help:"Generate shell completions"`
}


type exitPanic struct{ code int }

// Execute runs the CLI with the given arguments.
func Execute(args []string) (err error) {
	parser, err := newParser()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				if ep.code == 0 {
					err = nil

					return
				}

				err = &ExitError{Code: ep.code, Err: errors.New("exited")}

				return
			}

			panic(r)
		}
	}()

	if len(args) == 0 {
		args = []string{"--help"}
	}

	kctx, err := parser.Parse(args)
	if err != nil {
		parsedErr := wrapParseError(err)
		_, _ = fmt.Fprintln(os.Stderr, parsedErr)

		return parsedErr
	}

	err = kctx.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		return err
	}

	return nil
}

func wrapParseError(err error) error {
	if err == nil {
		return nil
	}

	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return &ExitError{Code: CodeUsage, Err: parseErr}
	}

	return err
}

func newParser() (*kong.Kong, error) {
	vars := kong.Vars{
		"version": VersionString(),
	}

	cli := &CLI{}
	parser, err := kong.New(
		cli,
		kong.Name("strawpoll"),
		kong.Description("StrawPoll CLI - Create and manage polls from the command line"),
		kong.Vars(vars),
		kong.Writers(os.Stdout, os.Stderr),
		kong.Exit(func(code int) { panic(exitPanic{code: code}) }),
		kong.Bind(&cli.RootFlags),
		kong.Help(helpPrinter),
		kong.ConfigureHelp(helpOptions()),
		kong.ExplicitGroups([]kong.Group{
			{Key: "voting", Title: "Voting Rules"},
			{Key: "privacy", Title: "Privacy & Access"},
			{Key: "display", Title: "Display & Scheduling"},
		}),
	)
	if err != nil {
		return nil, err
	}

	return parser, nil
}
