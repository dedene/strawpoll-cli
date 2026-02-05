package api

import "testing"

func TestParsePollID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"raw ID", "NPgxkzPqrn2", "NPgxkzPqrn2"},
		{"https URL", "https://strawpoll.com/NPgxkzPqrn2", "NPgxkzPqrn2"},
		{"https www URL", "https://www.strawpoll.com/NPgxkzPqrn2", "NPgxkzPqrn2"},
		{"polls path", "https://strawpoll.com/polls/NPgxkzPqrn2", "NPgxkzPqrn2"},
		{"http URL", "http://strawpoll.com/NPgxkzPqrn2", "NPgxkzPqrn2"},
		{"trimmed whitespace", "  NPgxkzPqrn2  ", "NPgxkzPqrn2"},
		{"no scheme", "strawpoll.com/NPgxkzPqrn2", "NPgxkzPqrn2"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePollID(tt.input)
			if got != tt.want {
				t.Errorf("ParsePollID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
