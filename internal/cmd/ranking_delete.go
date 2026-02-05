package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/tui"
)

// RankingDeleteCmd deletes a ranking poll.
type RankingDeleteCmd struct {
	ID    string `arg:"" required:"" help:"Poll ID or URL"`
	Force bool   `help:"Skip confirmation prompt" short:"f"`
}

// Run deletes a ranking poll, prompting for confirmation unless --force.
func (c *RankingDeleteCmd) Run(_ *RootFlags) error {
	id := api.ParsePollID(c.ID)

	if !c.Force {
		confirmed, err := tui.Confirm(fmt.Sprintf("Delete ranking poll %s? This cannot be undone.", id))
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

	fmt.Fprintf(os.Stderr, "Ranking poll %s deleted.\n", id)

	return nil
}
