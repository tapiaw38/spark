package config

import (
	"strings"
	"testing"
)

func TestCSSUsesTheReceiverNotTheGlobal(t *testing.T) {
	original := Current
	t.Cleanup(func() { Current = original })
	Current = Config{BackgroundColor: "1, 1, 1", TextColor: "red", SelectionColor: "2, 2, 2"}

	other := Config{
		BackgroundColor: "9, 9, 9",
		BackgroundAlpha: 0.5,
		TextColor:       "lime",
		SelectionColor:  "8, 8, 8",
		BorderRadius:    3,
		FontSize:        7,
	}

	css := other.CSS()
	for _, want := range []string{"9, 9, 9", "lime", "8, 8, 8"} {
		if !strings.Contains(css, want) {
			t.Errorf("CSS() dropped receiver value %q", want)
		}
	}
	for _, unwanted := range []string{"1, 1, 1", "red", "2, 2, 2"} {
		if strings.Contains(css, unwanted) {
			t.Errorf("CSS() leaked global value %q; it must read the receiver", unwanted)
		}
	}
}

func TestCSSIncludesStaticTheme(t *testing.T) {
	if !strings.Contains(Current.CSS(), "@define-color spark-white-60") {
		t.Error("CSS() must carry the static theme colors")
	}
}
