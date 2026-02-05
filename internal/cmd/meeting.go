package cmd

// MeetingCmd groups meeting poll subcommands.
type MeetingCmd struct {
	Create  MeetingCreateCmd  `cmd:"" help:"Create a meeting poll"`
	Get     MeetingGetCmd     `cmd:"" help:"Get meeting poll details"`
	Results MeetingResultsCmd `cmd:"" help:"View meeting availability"`
	Delete  MeetingDeleteCmd  `cmd:"" help:"Delete a meeting poll"`
	Update  MeetingUpdateCmd  `cmd:"" help:"Update a meeting poll"`
	List    MeetingListCmd    `cmd:"" help:"List meeting polls"`
}
