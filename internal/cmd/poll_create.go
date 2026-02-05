package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/pkg/browser"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/auth"
	"github.com/dedene/strawpoll-cli/internal/config"
	"github.com/dedene/strawpoll-cli/internal/output"
	"github.com/dedene/strawpoll-cli/internal/tui"
)

const pollBaseURL = "https://strawpoll.com/"

// PollCreateCmd creates a multiple-choice poll.
type PollCreateCmd struct {
	Title   string   `arg:"" optional:"" help:"Poll title"`
	Options []string `arg:"" optional:"" help:"Poll options (2-30)"`

	// Voting Rules group
	Dupcheck          string `help:"Duplication checking: ip, session, none" default:"ip" group:"voting"`
	IsMultipleChoice  bool   `help:"Allow selecting multiple options" group:"voting"`
	MultipleChoiceMin int    `help:"Minimum selections (requires --is-multiple-choice)" group:"voting"`
	MultipleChoiceMax int    `help:"Maximum selections (requires --is-multiple-choice)" group:"voting"`
	AllowOther        bool   `help:"Allow voters to add options" group:"voting"`
	RequireNames      bool   `help:"Require voter names" group:"voting"`

	// Privacy & Access group
	IsPrivate        bool   `help:"Hide from public listings" group:"privacy"`
	ResultsVis       string `help:"Results visibility: always, after_deadline, after_vote, hidden" default:"always" group:"privacy"`
	HideParticipants bool   `help:"Hide participant names" group:"privacy"`
	AllowVPN         bool   `help:"Allow VPN users" default:"true" group:"privacy"`
	EditVotePerms    string `help:"Who can edit votes: admin, admin_voter, voter, nobody" default:"admin_voter" group:"privacy"`

	// Display & Scheduling group
	Deadline      string `help:"Deadline (RFC3339 or duration like 24h)" group:"display"`
	Randomize     bool   `help:"Randomize option order" group:"display"`
	AllowComments bool   `help:"Allow comments on poll" group:"display"`
}

// Run creates a poll via the API.
// If title and options are provided as args, uses the flag path directly.
// Otherwise, launches an interactive wizard in TTY or errors in a pipe.
func (c *PollCreateCmd) Run(flags *RootFlags) error {
	// Flag-based path: title and options provided
	if c.Title != "" && len(c.Options) > 0 {
		return c.createFromFlags(flags)
	}

	// Interactive path: need TTY
	if !tui.IsInteractive() {
		return fmt.Errorf("missing title and options; provide as arguments or run interactively in a terminal")
	}

	result, err := tui.RunPollWizard()
	if err != nil {
		return err
	}

	// Map wizard result to command fields
	c.Title = result.Title
	c.Options = result.Options
	c.Dupcheck = result.Dupcheck
	c.IsMultipleChoice = result.IsMultipleChoice
	c.IsPrivate = result.IsPrivate
	c.ResultsVis = result.ResultsVis
	c.AllowComments = result.AllowComments

	return c.createFromFlags(flags)
}

// createFromFlags creates a poll using the values already set on the command struct.
func (c *PollCreateCmd) createFromFlags(flags *RootFlags) error {
	if len(c.Options) < 2 || len(c.Options) > 30 {
		return fmt.Errorf("poll requires 2-30 options, got %d", len(c.Options))
	}

	// Apply config defaults
	cfg, _ := config.ReadConfig()
	c.applyDefaults(cfg)

	apiKey, err := auth.GetAPIKey()
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	req := c.buildRequest()

	client := api.NewClient(apiKey)
	defer client.Close()

	poll, err := client.CreatePoll(context.Background(), req)
	if err != nil {
		return err
	}

	pollURL := pollBaseURL + poll.ID

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)
	headers := []string{"ID", "Title", "URL"}
	rows := [][]string{{poll.ID, poll.Title, pollURL}}

	if err := f.Output(poll, headers, rows); err != nil {
		return err
	}

	if flags.Copy {
		if err := clipboard.WriteAll(pollURL); err != nil {
			fmt.Fprintf(os.Stderr, "clipboard: %v\n", err)
		}
	}

	if flags.Open {
		if err := browser.OpenURL(pollURL); err != nil {
			fmt.Fprintf(os.Stderr, "browser: %v\n", err)
		}
	}

	return nil
}

func (c *PollCreateCmd) applyDefaults(cfg config.File) {
	if cfg.Dupcheck != "" {
		c.Dupcheck = cfg.Dupcheck
	}

	if cfg.ResultsVisibility != "" {
		c.ResultsVis = cfg.ResultsVisibility
	}

	if cfg.IsPrivate != nil {
		c.IsPrivate = *cfg.IsPrivate
	}

	if cfg.AllowComments != nil {
		c.AllowComments = *cfg.AllowComments
	}

	if cfg.AllowVPN != nil {
		c.AllowVPN = *cfg.AllowVPN
	}

	if cfg.HideParticipants != nil {
		c.HideParticipants = *cfg.HideParticipants
	}

	if cfg.EditVotePerms != "" {
		c.EditVotePerms = cfg.EditVotePerms
	}
}

func (c *PollCreateCmd) buildRequest() *api.CreatePollRequest {
	opts := make([]*api.PollOption, len(c.Options))
	for i, v := range c.Options {
		opts[i] = &api.PollOption{Value: v, Position: i}
	}

	pollCfg := &api.PollConfig{
		DuplicationChecking: c.Dupcheck,
		ResultsVisibility:   c.ResultsVis,
		EditVotePermissions: c.EditVotePerms,
		IsPrivate:           boolPtr(c.IsPrivate),
		AllowComments:       boolPtr(c.AllowComments),
		AllowVpn:            boolPtr(c.AllowVPN),
		HideParticipants:    boolPtr(c.HideParticipants),
		AllowOtherOption:    boolPtr(c.AllowOther),
		RequireNames:        boolPtr(c.RequireNames),
		Randomize:           boolPtr(c.Randomize),
	}

	if c.IsMultipleChoice {
		pollCfg.IsMultipleChoice = boolPtr(true)

		if c.MultipleChoiceMin > 0 {
			pollCfg.MultipleChoicesMin = intPtr(c.MultipleChoiceMin)
		}

		if c.MultipleChoiceMax > 0 {
			pollCfg.MultipleChoicesMax = intPtr(c.MultipleChoiceMax)
		}
	}

	if c.Deadline != "" {
		pollCfg.Deadline = parseDeadline(c.Deadline)
	}

	return &api.CreatePollRequest{
		Title:       c.Title,
		PollOptions: opts,
		PollConfig:  pollCfg,
	}
}

func parseDeadline(s string) string {
	// Try RFC3339 first
	if _, err := time.Parse(time.RFC3339, s); err == nil {
		return s
	}

	// Try as duration from now
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(d).UTC().Format(time.RFC3339)
	}

	return s
}

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }
