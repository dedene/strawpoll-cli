package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/tui"
)

// PollResetCmd resets poll results.
type PollResetCmd struct {
	ID    string `arg:"" required:"" help:"Poll ID or URL"`
	Force bool   `help:"Skip confirmation prompt" short:"f"`
}

// Run resets poll results, prompting for confirmation unless --force.
func (c *PollResetCmd) Run(_ *RootFlags) error {
	id := api.ParsePollID(c.ID)

	if !c.Force {
		confirmed, err := tui.Confirm(fmt.Sprintf("Reset results for poll %s? This cannot be undone.", id))
		if err != nil {
			return err
		}

		if !confirmed {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}
	}

	client, err := newClientFromAuth()
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.ResetPollResults(context.Background(), id); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Poll %s results reset.\n", id)

	return nil
}
