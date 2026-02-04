package api

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	t.Parallel()

	// Use a fixed reference time for relative date tests
	now := time.Now()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "ISO 8601 date",
			input: "2024-01-15",
			want:  "2024-01-15",
		},
		{
			name:  "relative -1d",
			input: "-1d",
			want:  now.AddDate(0, 0, -1).Format("2006-01-02"),
		},
		{
			name:  "relative -30d",
			input: "-30d",
			want:  now.AddDate(0, 0, -30).Format("2006-01-02"),
		},
		{
			name:  "relative -7d",
			input: "-7d",
			want:  now.AddDate(0, 0, -7).Format("2006-01-02"),
		},
		{
			name:  "RFC3339 datetime",
			input: "2024-01-15T10:30:00Z",
			want:  "2024-01-15",
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "invalid relative",
			input:   "-abcd",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseDate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDate(%q) expected error, got nil", tt.input)
				}

				return
			}

			if err != nil {
				t.Errorf("parseDate(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got != tt.want {
				t.Errorf("parseDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
