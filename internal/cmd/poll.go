package cmd

// PollCmd groups poll subcommands.
type PollCmd struct {
	Create  PollCreateCmd  `cmd:"" help:"Create a multiple-choice poll"`
	Get     PollGetCmd     `cmd:"" help:"Get poll details"`
	Results PollResultsCmd `cmd:"" help:"View poll results"`
	Delete  PollDeleteCmd  `cmd:"" help:"Delete a poll"`
	Update  PollUpdateCmd  `cmd:"" help:"Update a poll"`
	Reset   PollResetCmd   `cmd:"" help:"Reset poll results"`
	List    PollListCmd    `cmd:"" help:"List your polls"`
}
