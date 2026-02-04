package cmd

import "testing"

func TestParseEnabledCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  map[string]bool
	}{
		{
			name:  "empty string",
			input: "",
			want:  map[string]bool{},
		},
		{
			name:  "single command",
			input: "accounts",
			want:  map[string]bool{"accounts": true},
		},
		{
			name:  "multiple commands",
			input: "accounts,transactions,sync",
			want:  map[string]bool{"accounts": true, "transactions": true, "sync": true},
		},
		{
			name:  "with spaces",
			input: " accounts , transactions , sync ",
			want:  map[string]bool{"accounts": true, "transactions": true, "sync": true},
		},
		{
			name:  "case insensitive",
			input: "Accounts,TRANSACTIONS,Sync",
			want:  map[string]bool{"accounts": true, "transactions": true, "sync": true},
		},
		{
			name:  "full paths",
			input: "accounts.list,transactions.get",
			want:  map[string]bool{"accounts.list": true, "transactions.get": true},
		},
		{
			name:  "wildcard",
			input: "*",
			want:  map[string]bool{"*": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := parseEnabledCommands(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("parseEnabledCommands(%q) = %v, want %v", tt.input, got, tt.want)
				return
			}

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseEnabledCommands(%q)[%q] = %v, want %v", tt.input, k, got[k], v)
				}
			}
		})
	}
}
