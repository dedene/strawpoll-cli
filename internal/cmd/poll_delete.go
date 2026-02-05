package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/auth"
	"github.com/dedene/strawpoll-cli/internal/tui"
)

// PollDeleteCmd deletes a poll.
type PollDeleteCmd struct {
	ID    string `arg:"" required:"" help:"Poll ID or URL"`
	Force bool   `help:"Skip confirmation prompt" short:"f"`
}

// Run deletes a poll, prompting for confirmation unless --force.
func (c *PollDeleteCmd) Run(_ *RootFlags) error {
	id := api.ParsePollID(c.ID)

	if !c.Force {
		confirmed, err := tui.Confirm(fmt.Sprintf("Delete poll %s? This cannot be undone.", id))
		if err != nil {
			return err
		}

		if !confirmed {
			fmt.Fprintln(os.Stderr, "Aborted.")

			return nil
		}
	}

	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	client := api.NewClient(apiKey)
	defer client.Close()

	if err := client.DeletePoll(context.Background(), id); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Poll %s deleted.\n", id)

	return nil
}
