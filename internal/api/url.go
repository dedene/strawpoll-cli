package api

import (
	"net/url"
	"path"
	"strings"
)

// ParsePollID extracts a poll ID from a URL or returns the input as-is if it's a raw ID.
func ParsePollID(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	// Normalize: add scheme if missing so url.Parse works correctly.
	raw := input
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return input
	}

	host := strings.ToLower(u.Hostname())
	if host == "strawpoll.com" || host == "www.strawpoll.com" {
		// Last segment of path is the poll ID.
		segment := path.Base(u.Path)
		if segment != "" && segment != "." && segment != "/" {
			return segment
		}
	}

	// Not a recognized URL; return original trimmed input as raw poll ID.
	return strings.TrimSpace(input)
}
