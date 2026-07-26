package modules

import "testing"

func TestParsePlaybackStatusNormalizesCase(t *testing.T) {
	tests := []struct {
		raw  string
		want PlaybackStatus
	}{
		{"Playing", StatusPlaying},
		{"playing", StatusPlaying},
		{"PLAYING", StatusPlaying},
		{"Paused", StatusPaused},
		{"paused", StatusPaused},
		{"Stopped", StatusStopped},
		{"stopped", StatusStopped},
		{"  Playing\n", StatusPlaying},
		{"", ""},
	}
	for _, tt := range tests {
		if got := ParsePlaybackStatus(tt.raw); got != tt.want {
			t.Errorf("ParsePlaybackStatus(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestParsePlaybackStatusKeepsUnknownValues(t *testing.T) {
	got := ParsePlaybackStatus("  Buffering  ")
	if got != PlaybackStatus("Buffering") {
		t.Errorf("ParsePlaybackStatus(Buffering) = %q, want Buffering", got)
	}
	for _, known := range []PlaybackStatus{StatusPlaying, StatusPaused, StatusStopped} {
		if got == known {
			t.Errorf("unknown status must not collapse into %q", known)
		}
	}
}

func TestIsYouTubeMetaUsesNormalizedStatus(t *testing.T) {
	lower := playerMeta{
		name:   "firefox",
		url:    "https://youtube.com/feed",
		status: ParsePlaybackStatus("playing"),
	}
	if !isYouTubeMeta(lower) {
		t.Error("a lowercase playing status must still count as playing")
	}

	stopped := playerMeta{
		name:   "firefox youtube",
		status: ParsePlaybackStatus("STOPPED"),
	}
	if isYouTubeMeta(stopped) {
		t.Error("a stopped player must not match regardless of case")
	}
}
