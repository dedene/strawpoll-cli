package cmd

import "testing"

func TestVersionString(t *testing.T) {
	tests := []struct {
		name    string
		ver     string
		com     string
		dat     string
		want    string
	}{
		{"defaults", "dev", "", "", "dev"},
		{"version-only", "v1.0.0", "", "", "v1.0.0"},
		{"version-commit", "v1.0.0", "abc1234", "", "v1.0.0 (abc1234)"},
		{"version-date", "v1.0.0", "", "2025-01-01", "v1.0.0 (2025-01-01)"},
		{"all-set", "v1.0.0", "abc1234", "2025-01-01", "v1.0.0 (abc1234 2025-01-01)"},
		{"empty-version", "", "", "", "dev"},
		{"whitespace-version", "  ", "", "", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore package vars.
			origV, origC, origD := version, commit, date
			defer func() { version, commit, date = origV, origC, origD }()

			version, commit, date = tt.ver, tt.com, tt.dat

			if got := VersionString(); got != tt.want {
				t.Errorf("VersionString() = %q, want %q", got, tt.want)
			}
		})
	}
}
