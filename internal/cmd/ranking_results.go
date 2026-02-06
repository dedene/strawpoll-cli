package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/dedene/strawpoll-cli/internal/api"
	"github.com/dedene/strawpoll-cli/internal/output"
)

// RankingResultsCmd displays ranking poll results with Borda count scoring.
type RankingResultsCmd struct {
	ID      string `arg:"" required:"" help:"Poll ID or URL"`
	Verbose bool   `help:"Show per-option position breakdown" short:"v"`
}

// Run fetches ranking results and displays Borda count scores.
func (c *RankingResultsCmd) Run(flags *RootFlags) error {
	id := api.ParsePollID(c.ID)

	client, err := newClientFromAuth()
	if err != nil {
		return err
	}
	defer client.Close()

	results, err := client.GetPollResults(context.Background(), id)
	if err != nil {
		return err
	}

	f := output.NewFormatter(os.Stdout, flags.JSON, flags.Plain, flags.NoColor)

	// JSON mode: output enriched struct with computed scores
	if flags.JSON {
		enriched := buildRankingJSON(results)
		return f.Output(enriched, nil, nil)
	}

	// Summary table sorted by score descending
	headers, rows := rankingScoreTable(results)
	if err := f.Output(results, headers, rows); err != nil {
		return err
	}

	// Position breakdown if --verbose
	if c.Verbose && len(results.PollOptions) > 0 {
		fmt.Fprintln(os.Stdout)

		bHeaders, bRows := rankingBreakdownTable(results)
		if err := f.Output(results, bHeaders, bRows); err != nil {
			return err
		}
	}

	return nil
}

// bordaScores computes Borda count scores for each option.
// For n options, position 0 (first place) scores n points, position n-1 scores 1 point.
func bordaScores(results *api.PollResults) []int {
	n := len(results.PollOptions)
	scores := make([]int, n)

	for _, p := range results.PollParticipants {
		for i, v := range p.PollVotes {
			if v != nil && i < n {
				scores[i] += n - *v
			}
		}
	}

	return scores
}

// positionBreakdown computes how many times each option was placed at each position.
// breakdown[optionIdx][positionIdx] = count of participants who ranked that option at that position.
func positionBreakdown(results *api.PollResults) [][]int {
	n := len(results.PollOptions)
	breakdown := make([][]int, n)

	for i := range breakdown {
		breakdown[i] = make([]int, n)
	}

	for _, p := range results.PollParticipants {
		for i, v := range p.PollVotes {
			if v != nil && i < n && *v < n {
				breakdown[i][*v]++
			}
		}
	}

	return breakdown
}

// rankingScoreTable builds the summary table sorted by score descending.
func rankingScoreTable(results *api.PollResults) ([]string, [][]string) {
	n := len(results.PollOptions)
	scores := bordaScores(results)
	maxScore := n * len(results.PollParticipants)

	// Build sortable entries
	type entry struct {
		name  string
		score int
	}

	entries := make([]entry, n)
	for i, opt := range results.PollOptions {
		entries[i] = entry{name: opt.Value, score: scores[i]}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].score > entries[j].score
	})

	headers := []string{"Option", "Score", "Percentage"}
	rows := make([][]string, n)

	for i, e := range entries {
		pct := 0.0
		if maxScore > 0 {
			pct = float64(e.score) / float64(maxScore) * 100
		}

		rows[i] = []string{
			e.name,
			fmt.Sprintf("%d", e.score),
			fmt.Sprintf("%.1f%%", pct),
		}
	}

	return headers, rows
}

// rankingBreakdownTable builds the per-option position breakdown table.
func rankingBreakdownTable(results *api.PollResults) ([]string, [][]string) {
	n := len(results.PollOptions)
	breakdown := positionBreakdown(results)

	// Headers: Option, #1, #2, #3, ...
	headers := make([]string, 0, 1+n)
	headers = append(headers, "Option")

	for i := 0; i < n; i++ {
		headers = append(headers, fmt.Sprintf("#%d", i+1))
	}

	rows := make([][]string, n)

	for i, opt := range results.PollOptions {
		row := make([]string, 0, 1+n)
		row = append(row, opt.Value)

		for j := 0; j < n; j++ {
			row = append(row, fmt.Sprintf("%d", breakdown[i][j]))
		}

		rows[i] = row
	}

	return headers, rows
}

// rankingJSONResult is the enriched JSON output for ranking results.
type rankingJSONResult struct {
	ID               string              `json:"id"`
	VoteCount        int                 `json:"voteCount"`
	ParticipantCount int                 `json:"participantCount"`
	Options          []rankingJSONOption `json:"options"`
}

type rankingJSONOption struct {
	ID         string  `json:"id"`
	Value      string  `json:"value"`
	Score      int     `json:"score"`
	Percentage float64 `json:"percentage"`
	Positions  []int   `json:"positions"`
}

func buildRankingJSON(results *api.PollResults) *rankingJSONResult {
	n := len(results.PollOptions)
	scores := bordaScores(results)
	breakdown := positionBreakdown(results)
	maxScore := n * len(results.PollParticipants)

	opts := make([]rankingJSONOption, n)
	for i, opt := range results.PollOptions {
		pct := 0.0
		if maxScore > 0 {
			pct = float64(scores[i]) / float64(maxScore) * 100
		}

		opts[i] = rankingJSONOption{
			ID:         opt.ID,
			Value:      opt.Value,
			Score:      scores[i],
			Percentage: pct,
			Positions:  breakdown[i],
		}
	}

	return &rankingJSONResult{
		ID:               results.ID,
		VoteCount:        results.VoteCount,
		ParticipantCount: results.ParticipantCount,
		Options:          opts,
	}
}
