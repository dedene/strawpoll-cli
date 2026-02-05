package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/tui"
)

// MeetingDeleteCmd deletes a meeting poll.
type MeetingDeleteCmd struct {
	ID    string `arg:"" required:"" help:"Meeting poll ID or URL"`
	Force bool   `help:"Skip confirmation prompt" short:"f"`
}

// Run deletes a meeting poll, prompting for confirmation unless --force.
func (c *MeetingDeleteCmd) Run(_ *RootFlags) error {
	id := api.ParsePollID(c.ID)

	if !c.Force {
		confirmed, err := tui.Confirm(fmt.Sprintf("Delete meeting poll %s? This cannot be undone.", id))
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

	if err := client.DeletePoll(context.Background(), id); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Meeting poll %s deleted.\n", id)

	return nil
}
