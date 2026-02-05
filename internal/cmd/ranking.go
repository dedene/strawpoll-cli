package cmd

// RankingCmd groups ranking poll subcommands.
type RankingCmd struct {
	Create  RankingCreateCmd  `cmd:"" help:"Create a ranking poll"`
	Get     RankingGetCmd     `cmd:"" help:"Get ranking poll details"`
	Results RankingResultsCmd `cmd:"" help:"View ranking results"`
	Delete  RankingDeleteCmd  `cmd:"" help:"Delete a ranking poll"`
	Update  RankingUpdateCmd  `cmd:"" help:"Update a ranking poll"`
	List    RankingListCmd    `cmd:"" help:"List ranking polls"`
}
