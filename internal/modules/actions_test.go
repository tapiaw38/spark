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
		{"terminal", TerminalAction("ssh host"), ActionSpec{Kind: ActionTerminal, Target: "ssh host"}},
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

func TestOpenPathsActionCarriesEveryPath(t *testing.T) {
	action := OpenPathsAction("/a", "/b", "/c")
	if action.Kind != ActionOpen {
		t.Fatalf("kind = %q, want %q", action.Kind, ActionOpen)
	}
	got := append([]string{action.Target}, action.Args...)
	want := []string{"/a", "/b", "/c"}
	if len(got) != len(want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("paths = %v, want %v", got, want)
		}
	}
}

func TestOpenPathsActionWithNoPathsIsZero(t *testing.T) {
	if !OpenPathsAction().IsZero() {
		t.Error("OpenPathsAction() with no paths must be a zero action")
	}
}

func TestOpenBufferedFilesDoesNotDependOnTheMusicQueue(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	fileBufferMu.Lock()
	fileBuffer, bufferLoaded = nil, false
	fileBufferMu.Unlock()

	AddFileToBuffer("/tmp/one.txt")
	AddFileToBuffer("/tmp/two.txt")
	t.Cleanup(ClearFileBuffer)

	results := FileBufferSearch("buffer")
	var open *Result
	for i := range results {
		if results[i].Title == "Open Buffered Files" {
			open = &results[i]
			break
		}
	}
	if open == nil {
		t.Fatal("FileBufferSearch did not return an Open Buffered Files row")
	}
	if open.ActionSpec.Kind == ActionMusic {
		t.Fatal("opening buffered files must not route through the music executor, which aborts on an empty music queue")
	}
	if open.ActionSpec.Kind != ActionOpen {
		t.Fatalf("kind = %q, want %q", open.ActionSpec.Kind, ActionOpen)
	}
	paths := append([]string{open.ActionSpec.Target}, open.ActionSpec.Args...)
	if len(paths) != 2 {
		t.Fatalf("carried %d paths, want 2: %v", len(paths), paths)
	}
}
