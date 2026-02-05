package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/pkg/browser"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/config"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// RankingCreateCmd creates a ranking poll.
type RankingCreateCmd struct {
	Title   string   `arg:"" required:"" help:"Poll title"`
	Options []string `arg:"" required:"" help:"Ranking options (2-30)"`

	// Voting Rules group
	Dupcheck string `help:"Duplication checking: ip, session, none" default:"ip" group:"voting"`

	// Privacy & Access group
	IsPrivate  bool   `help:"Hide from public listings" group:"privacy"`
	ResultsVis string `help:"Results visibility: always, after_deadline, after_vote, hidden" default:"always" group:"privacy"`

	// Display & Scheduling group
	Deadline      string `help:"Deadline (RFC3339 or duration like 24h)" group:"display"`
	AllowComments bool   `help:"Allow comments on poll" group:"display"`
	Description   string `help:"Poll description" group:"display"`
}

// Run creates a ranking poll via the API.
func (c *RankingCreateCmd) Run(flags *RootFlags) error {
	if len(c.Options) < 2 || len(c.Options) > 30 {
		return fmt.Errorf("ranking poll requires 2-30 options, got %d", len(c.Options))
	}

	cfg, _ := config.ReadConfig()
	c.applyDefaults(cfg)

	client, err := newClientFromAuth()
	if err != nil {
		return err
	}
	defer client.Close()

	req := c.buildRequest()

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

func (c *RankingCreateCmd) applyDefaults(cfg config.File) {
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
}

func (c *RankingCreateCmd) buildRequest() *api.CreatePollRequest {
	opts := make([]*api.PollOption, len(c.Options))
	for i, v := range c.Options {
		opts[i] = &api.PollOption{Type: api.OptionTypeText, Value: v, Position: i}
	}

	pollCfg := &api.PollConfig{
		DuplicationChecking: c.Dupcheck,
		ResultsVisibility:   c.ResultsVis,
		IsPrivate:           boolPtr(c.IsPrivate),
		AllowComments:       boolPtr(c.AllowComments),
	}

	if c.Deadline != "" {
		pollCfg.Deadline = rankingParseDeadline(c.Deadline)
	}

	req := &api.CreatePollRequest{
		Title:       c.Title,
		Type:        api.PollTypeRanking,
		PollOptions: opts,
		PollConfig:  pollCfg,
	}

	if c.Description != "" {
		req.PollMeta = &api.PollMeta{Description: c.Description}
	}

	return req
}

func rankingParseDeadline(s string) string {
	if _, err := time.Parse(time.RFC3339, s); err == nil {
		return s
	}

	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(d).UTC().Format(time.RFC3339)
	}

	return s
}
