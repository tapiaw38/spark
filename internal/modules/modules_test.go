package modules

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestAllResultTypesHasNoDuplicates(t *testing.T) {
	seen := make(map[ResultType]bool)
	for _, rt := range AllResultTypes() {
		if seen[rt] {
			t.Errorf("duplicate entry in AllResultTypes: %q", rt)
		}
		seen[rt] = true
	}
}

func TestAllResultTypesAreLowerKebabCase(t *testing.T) {
	valid := regexp.MustCompile(`^[a-z]+(-[a-z]+)*$`)
	for _, rt := range AllResultTypes() {
		if !valid.MatchString(string(rt)) {
			t.Errorf("ResultType %q is not lower-kebab-case", rt)
		}
	}
}

func TestEveryTypeConstantUsedIsRegistered(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}

	registered := make(map[string]bool)
	for _, rt := range AllResultTypes() {
		registered[constantNameFor(rt)] = true
	}

	usage := regexp.MustCompile(`\bType:\s*(Type[A-Za-z]+)`)
	found := 0
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(".", e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		for _, m := range usage.FindAllStringSubmatch(string(src), -1) {
			found++
			if !registered[m[1]] {
				t.Errorf("%s uses %s, which is missing from AllResultTypes()", e.Name(), m[1])
			}
		}
	}
	if found == 0 {
		t.Fatal("no Type: constant usages found; the scan regex is stale")
	}
}

func TestNoStringLiteralResultTypes(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}

	literal := regexp.MustCompile(`\bType:\s*"`)
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(".", e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		if literal.MatchString(string(src)) {
			t.Errorf("%s assigns Result.Type from a string literal; use a ResultType constant", e.Name())
		}
	}
}

func TestResultTypesUsedByPreviewAreRegistered(t *testing.T) {
	for _, rt := range []ResultType{TypeFile, TypeDirectory, TypeClipboard, TypeSnippet, TypeDictionary, TypeCalc} {
		if !slices.Contains(AllResultTypes(), rt) {
			t.Errorf("%q drives preview behaviour but is not in AllResultTypes()", rt)
		}
	}
}

func constantNameFor(rt ResultType) string {
	switch rt {
	case TypeApp:
		return "TypeApp"
	case TypeBookmark:
		return "TypeBookmark"
	case TypeCalc:
		return "TypeCalc"
	case TypeClipboard:
		return "TypeClipboard"
	case TypeContact:
		return "TypeContact"
	case TypeDevtool:
		return "TypeDevtool"
	case TypeDictionary:
		return "TypeDictionary"
	case TypeDirectory:
		return "TypeDirectory"
	case TypeEmail:
		return "TypeEmail"
	case TypeEmoji:
		return "TypeEmoji"
	case TypeFile:
		return "TypeFile"
	case TypeFileAction:
		return "TypeFileAction"
	case TypeFileBuffer:
		return "TypeFileBuffer"
	case TypeFileBufferAction:
		return "TypeFileBufferAction"
	case TypeFileOp:
		return "TypeFileOp"
	case TypeHelp:
		return "TypeHelp"
	case TypeKill:
		return "TypeKill"
	case TypeLargeType:
		return "TypeLargeType"
	case TypeMediaControl:
		return "TypeMediaControl"
	case TypeMusic:
		return "TypeMusic"
	case TypeNavigation:
		return "TypeNavigation"
	case TypePass:
		return "TypePass"
	case TypeRecent:
		return "TypeRecent"
	case TypeScreenshot:
		return "TypeScreenshot"
	case TypeShell:
		return "TypeShell"
	case TypeSnippet:
		return "TypeSnippet"
	case TypeSpell:
		return "TypeSpell"
	case TypeSpotify:
		return "TypeSpotify"
	case TypeSSH:
		return "TypeSSH"
	case TypeStats:
		return "TypeStats"
	case TypeStatus:
		return "TypeStatus"
	case TypeSync:
		return "TypeSync"
	case TypeSystem:
		return "TypeSystem"
	case TypeTimer:
		return "TypeTimer"
	case TypeUndo:
		return "TypeUndo"
	case TypeWeather:
		return "TypeWeather"
	case TypeWeb:
		return "TypeWeb"
	case TypeWindow:
		return "TypeWindow"
	case TypeWorkspace:
		return "TypeWorkspace"
	case TypeYouTube:
		return "TypeYouTube"
	case TypeYouTubePlayer:
		return "TypeYouTubePlayer"
	default:
		return ""
	}
}
