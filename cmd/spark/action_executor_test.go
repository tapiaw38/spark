package main

import (
	"testing"

	"github.com/tapiaw38/spark/internal/modules"
)

func TestEveryActionKindHasHandler(t *testing.T) {
	for _, kind := range modules.AllActionKinds() {
		if _, ok := actionHandlers[kind]; !ok {
			t.Errorf("ActionKind %q has no handler in actionHandlers", kind)
		}
	}
}

func TestNoHandlerForUnlistedKind(t *testing.T) {
	listed := make(map[modules.ActionKind]bool)
	for _, kind := range modules.AllActionKinds() {
		listed[kind] = true
	}
	for kind := range actionHandlers {
		if !listed[kind] {
			t.Errorf("handler registered for %q, which AllActionKinds() does not list", kind)
		}
	}
}

func TestZeroActionIsNotExecuted(t *testing.T) {
	if executeActionSpec(modules.ActionSpec{}) {
		t.Error("a zero ActionSpec must not report success")
	}
}

func TestUnknownActionFails(t *testing.T) {
	if executeActionSpec(modules.ActionSpec{Kind: modules.ActionKind("nope")}) {
		t.Error("an unknown ActionKind must not report success")
	}
}

func TestCopyToClipboardFailsWhenNoWriterInstalled(t *testing.T) {
	original := clipboardWriters
	t.Cleanup(func() { clipboardWriters = original })
	clipboardWriters = [][]string{{"spark-no-such-clipboard-tool"}}

	if copyToClipboard("text") {
		t.Error("copyToClipboard must report failure when no clipboard tool exists")
	}
}

func TestCopyActionDoesNotReportSuccessWhenCopyFails(t *testing.T) {
	original := clipboardWriters
	t.Cleanup(func() { clipboardWriters = original })
	clipboardWriters = [][]string{{"spark-no-such-clipboard-tool"}}

	action := modules.CopyAction("text").WithStatus("Copied!")
	if executeActionSpec(action) {
		t.Error("a failed copy must not report success, or SuccessStatus overwrites the error")
	}
}

func TestOpenTerminalFailsWhenNoTerminalInstalled(t *testing.T) {
	original := terminals
	t.Cleanup(func() { terminals = original })
	terminals = []string{"spark-no-such-terminal"}

	if openTerminal("true") {
		t.Error("openTerminal must report failure when no terminal exists")
	}
}

func TestPasteDoesNotFireWhenCopyFails(t *testing.T) {
	original := clipboardWriters
	t.Cleanup(func() { clipboardWriters = original })
	clipboardWriters = [][]string{{"spark-no-such-clipboard-tool"}}

	if pasteText("snippet body") {
		t.Error("pasteText must not paste stale clipboard content after a failed copy")
	}
}
