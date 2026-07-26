package apps

import (
	"strings"
	"testing"
)

func TestLaunchRejectsAppWithoutCommand(t *testing.T) {
	for _, exec := range []string{"", "   ", "\t\n"} {
		app := App{Name: "Broken", Exec: exec}
		if err := app.Launch(); err == nil {
			t.Errorf("Launch() with Exec=%q must return an error, not panic or succeed", exec)
		}
	}
}

func TestLaunchReportsTheAppNameInTheError(t *testing.T) {
	err := App{Name: "Firefox"}.Launch()
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "Firefox") {
		t.Errorf("error %q should name the app", err)
	}
}
