package api

import "time"

// Results visibility constants (actual API values).
const (
	ResultsVisibilityAlways        = "always"
	ResultsVisibilityAfterDeadline = "after_deadline"
	ResultsVisibilityAfterVote     = "after_vote"
	ResultsVisibilityHidden        = "hidden" // NOT "never" -- doc bug
)

// Poll type constants (actual API values).
const (
	PollTypeMultipleChoice = "multiple_choice"
	PollTypeMeeting        = "meeting"
	PollTypeRanking        = "ranking" // NOT "ranked_choice" -- doc bug
)

// Option type constants.
const (
	OptionTypeText      = "text"
	OptionTypeDate      = "date"
	OptionTypeTimeRange = "time_range"
)

// Vote type constants.
const (
	VoteTypeDefault         = "default"
	VoteTypeParticipantGrid = "participant_grid"
)

// Duplicate check constants.
const (
	DupcheckIP      = "ip"
	DupcheckSession = "session"
	DupcheckNone    = "none"
)

// Poll represents a StrawPoll poll.
type Poll struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Type        string       `json:"type"`
	PollOptions []*PollOption `json:"poll_options"`
	PollConfig  *PollConfig  `json:"poll_config"`
	PollMeta    *PollMeta    `json:"poll_meta"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	ResetAt     *time.Time   `json:"reset_at"`
	Version     int          `json:"version"`
}

// PollOption represents a single option in a poll.
type PollOption struct {
	ID          string `json:"id,omitempty"`
	Type        string `json:"type,omitempty"`
	Value       string `json:"value"`
	Position    int    `json:"position,omitempty"`
	VoteCount   int    `json:"vote_count,omitempty"`
	MaxVotes    int    `json:"max_votes,omitempty"`
	Description string `json:"description,omitempty"`
	IsWriteIn   bool   `json:"is_write_in,omitempty"`
	Date        string `json:"date,omitempty"`       // YYYY-MM-DD for type="date" (meeting all-day)
	StartTime   *int64 `json:"start_time,omitempty"` // Unix timestamp for type="time_range"
	EndTime     *int64 `json:"end_time,omitempty"`   // Unix timestamp for type="time_range"
}

// PollConfig holds all poll configuration fields.
type PollConfig struct {
	IsPrivate            *bool   `json:"is_private,omitempty"`
	VoteType             string  `json:"vote_type,omitempty"`
	AllowComments        *bool   `json:"allow_comments,omitempty"`
	AllowIndecision      *bool   `json:"allow_indecision,omitempty"`
	AllowOtherOption     *bool   `json:"allow_other_option,omitempty"`
	AllowVpn             *bool   `json:"allow_vpn,omitempty"`
	Deadline             string  `json:"deadline,omitempty"`
	DuplicationChecking  string  `json:"duplication_checking,omitempty"`
	EditVotePermissions  string  `json:"edit_vote_permissions,omitempty"`
	ForceAppearance      string  `json:"force_appearance,omitempty"`
	HideParticipants     *bool   `json:"hide_participants,omitempty"`
	IsMultipleChoice     *bool   `json:"is_multiple_choice,omitempty"`
	MaxChoices           *int    `json:"max_choices,omitempty"`
	MinChoices           *int    `json:"min_choices,omitempty"`
	MultipleChoicesMax   *int    `json:"multiple_choices_max,omitempty"`
	MultipleChoicesMin   *int    `json:"multiple_choices_min,omitempty"`
	NumberOfWinners      *int    `json:"number_of_winners,omitempty"`
	Randomize            *bool   `json:"randomize,omitempty"`
	RequireNames         *bool   `json:"require_names,omitempty"`
	ResultsVisibility    string  `json:"results_visibility,omitempty"`
}

// PollMeta holds poll metadata.
type PollMeta struct {
	Description      string `json:"description,omitempty"`
	Location         string `json:"location,omitempty"`
	Timezone         string `json:"timezone,omitempty"`          // IANA timezone (e.g. "Europe/Berlin")
	VoteCount        int    `json:"vote_count,omitempty"`
	ParticipantCount int    `json:"participant_count,omitempty"` // read-only
	ViewCount        int    `json:"view_count,omitempty"`        // read-only
}

// PollResults represents poll results from the API.
type PollResults struct {
	ID               string             `json:"id"`
	Version          int                `json:"version"`
	VoteCount        int                `json:"voteCount"`        // camelCase per OpenAPI spec
	ParticipantCount int                `json:"participantCount"` // camelCase per OpenAPI spec
	PollOptions      []*PollOption      `json:"poll_options"`
	PollParticipants []*PollParticipant `json:"poll_participants"`
}

// PollParticipant represents a participant in a poll.
type PollParticipant struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	CountryCode   string `json:"country_code"`
	IsEditAllowed bool   `json:"is_edit_allowed"`
	PollVotes     []*int `json:"poll_votes"`
	CreatedAt     string `json:"created_at"`
}

// CreatePollRequest is the request body for creating a poll.
type CreatePollRequest struct {
	Title       string       `json:"title"`
	Type        string       `json:"type"`
	PollOptions []*PollOption `json:"poll_options"`
	PollConfig  *PollConfig  `json:"poll_config,omitempty"`
	PollMeta    *PollMeta    `json:"poll_meta,omitempty"`
}

// UpdatePollRequest is the request body for updating a poll.
type UpdatePollRequest struct {
	Title       string       `json:"title,omitempty"`
	PollOptions []*PollOption `json:"poll_options,omitempty"`
	PollConfig  *PollConfig  `json:"poll_config,omitempty"`
	PollMeta    *PollMeta    `json:"poll_meta,omitempty"`
}

// Pagination holds page-based pagination metadata.
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// PollListResponse represents a paginated list of polls.
type PollListResponse struct {
	Data       []Poll     `json:"data"`
	Pagination Pagination `json:"pagination"`
}
