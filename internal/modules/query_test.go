package modules

import "testing"

func TestMatchesAny(t *testing.T) {
	tests := []struct {
		query string
		names []string
		want  bool
	}{
		{"stats", []string{"stats", "usage"}, true},
		{"  USAGE ", []string{"stats", "usage"}, true},
		{"stats today", []string{"stats", "usage"}, false},
		{"statsy", []string{"stats"}, false},
		{"", []string{"stats"}, false},
	}
	for _, tt := range tests {
		if got := MatchesAny(tt.query, tt.names...); got != tt.want {
			t.Errorf("MatchesAny(%q, %v) = %v, want %v", tt.query, tt.names, got, tt.want)
		}
	}
}

func TestMatchCommand(t *testing.T) {
	tests := []struct {
		query   string
		names   []string
		wantArg string
		wantOK  bool
	}{
		{"ssh", []string{"ssh"}, "", true},
		{"ssh web-01", []string{"ssh"}, "web-01", true},
		{"SSH  web-01  ", []string{"ssh"}, "web-01", true},
		{"define Ephemeral", []string{"define", "def"}, "Ephemeral", true},
		{"def Ephemeral", []string{"define", "def"}, "Ephemeral", true},
		{"sshd", []string{"ssh"}, "", false},
		{"", []string{"ssh"}, "", false},
	}
	for _, tt := range tests {
		arg, ok := MatchCommand(tt.query, tt.names...)
		if arg != tt.wantArg || ok != tt.wantOK {
			t.Errorf("MatchCommand(%q, %v) = (%q, %v), want (%q, %v)",
				tt.query, tt.names, arg, ok, tt.wantArg, tt.wantOK)
		}
	}
}
