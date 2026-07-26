package modules

import "testing"

func TestActionSpecConstructors(t *testing.T) {
	tests := []struct {
		name string
		got  ActionSpec
		want ActionSpec
	}{
		{"open", OpenAction("/tmp/file"), ActionSpec{Kind: ActionOpen, Target: "/tmp/file"}},
		{"copy", CopyAction("text"), ActionSpec{Kind: ActionCopy, Target: "text"}},
		{"start", StartAction("xdg-open", "/tmp/file"), ActionSpec{Kind: ActionStart, Target: "xdg-open", Args: []string{"/tmp/file"}}},
		{"run", RunAction("wl-copy", "text"), ActionSpec{Kind: ActionRun, Target: "wl-copy", Args: []string{"text"}}},
		{"terminal", TerminalAction("ssh host"), ActionSpec{Kind: ActionTerm, Target: "ssh host"}},
		{"app", AppAction("Firefox", "firefox", "firefox"), ActionSpec{Kind: ActionApp, Target: "Firefox", Args: []string{"firefox", "firefox"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got.Kind != tt.want.Kind || tt.got.Target != tt.want.Target {
				t.Fatalf("action = %#v, want %#v", tt.got, tt.want)
			}
			if len(tt.got.Args) != len(tt.want.Args) {
				t.Fatalf("args = %#v, want %#v", tt.got.Args, tt.want.Args)
			}
			for i := range tt.want.Args {
				if tt.got.Args[i] != tt.want.Args[i] {
					t.Fatalf("args = %#v, want %#v", tt.got.Args, tt.want.Args)
				}
			}
		})
	}
}

func TestActionSpecIsZero(t *testing.T) {
	if !(ActionSpec{}).IsZero() {
		t.Fatal("empty ActionSpec should be zero")
	}
	if OpenAction("/tmp/file").IsZero() {
		t.Fatal("open action should not be zero")
	}
}

func TestActionSpecWithStatus(t *testing.T) {
	action := StartAction("sh", "-c", "true").WithStatus("done")
	if action.SuccessStatus != "done" {
		t.Fatalf("status = %q, want done", action.SuccessStatus)
	}
}
