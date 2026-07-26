package main

import "testing"

func TestLookupCommand(t *testing.T) {
	tests := []struct {
		name     string
		argv     []string
		wantOK   bool
		wantArgs []string
	}{
		{"no args runs the launcher", nil, false, nil},
		{"unknown flag runs the launcher", []string{"--nope"}, false, nil},
		{"flag with no argument", []string{"--stats-window"}, true, []string{}},
		{"flag with its argument", []string{"--large-type", "hello"}, true, []string{"hello"}},
		{"missing argument", []string{"--large-type"}, false, nil},
		{"file op needs three", []string{"--file-op-window", "copy", "a"}, false, nil},
		{"file op with three", []string{"--file-op-window", "copy", "a", "b"}, true, []string{"copy", "a", "b"}},
		{"email with none", []string{"--email-window"}, true, []string{}},
		{"email with some", []string{"--email-window", "me@example.com"}, true, []string{"me@example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, args, ok := lookupCommand(tt.argv)
			if ok != tt.wantOK {
				t.Fatalf("lookupCommand(%q) ok = %v, want %v", tt.argv, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("lookupCommand(%q) args = %q, want %q", tt.argv, args, tt.wantArgs)
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestArgAt(t *testing.T) {
	args := []string{"a", "b"}
	for i, want := range []string{"a", "b", "", ""} {
		if got := argAt(args, i); got != want {
			t.Errorf("argAt(%q, %d) = %q, want %q", args, i, got, want)
		}
	}
}
