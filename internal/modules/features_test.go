package modules

import (
	"regexp"
	"strings"
	"testing"
)

func TestConvertUnit(t *testing.T) {
	cases := []struct {
		val      float64
		from, to string
		want     float64
	}{
		{1, "km", "m", 1000},
		{1000, "m", "km", 1},
		{1, "mi", "km", 1.609344},
		{1, "kg", "g", 1000},
		{1, "gb", "mb", 1024},
		{0, "c", "f", 32},
		{100, "c", "f", 212},
		{32, "f", "c", 0},
		{0, "c", "k", 273.15},
	}
	for _, c := range cases {
		got, ok := convertUnit(c.val, c.from, c.to)
		if !ok {
			t.Fatalf("convertUnit(%v,%s,%s) not ok", c.val, c.from, c.to)
		}
		if diff := got - c.want; diff > 0.001 || diff < -0.001 {
			t.Errorf("convertUnit(%v,%s,%s)=%v want %v", c.val, c.from, c.to, got, c.want)
		}
	}
	if _, ok := convertUnit(1, "km", "kg"); ok {
		t.Error("cross-category km->kg should fail")
	}
}

func TestUnitSearchParsing(t *testing.T) {
	if r := UnitSearch("100 km to mi"); len(r) != 1 {
		t.Fatalf("expected 1 result, got %d", len(r))
	}
	if r := UnitSearch("hello world"); r != nil {
		t.Error("non-conversion query should return nil")
	}
}

func TestDevTools(t *testing.T) {
	if r := DevToolsSearch("b64 hello"); len(r) != 1 || r[0].Title != "aGVsbG8=" {
		t.Fatalf("b64 encode wrong: %+v", r)
	}
	if r := DevToolsSearch("b64d aGVsbG8="); len(r) != 1 || r[0].Title != "hello" {
		t.Fatalf("b64 decode wrong: %+v", r)
	}
	if r := DevToolsSearch("url a b&c"); len(r) != 1 || !strings.Contains(r[0].Title, "%26") {
		t.Fatalf("url encode wrong: %+v", r)
	}
	if r := DevToolsSearch("brave browser"); r != nil {
		t.Error("unknown dev command should return nil")
	}
}

func TestEmojiSearchUsesIconText(t *testing.T) {
	r := EmojiSearch("emoji fire")
	if len(r) == 0 {
		t.Fatal("expected emoji result")
	}
	if r[0].IconText != "🔥" {
		t.Fatalf("expected fire IconText, got %+v", r[0])
	}
	if strings.Contains(r[0].Title, "🔥") {
		t.Fatalf("title should not duplicate emoji glyph: %+v", r[0])
	}
}

func TestKillSearchListsWithoutFilter(t *testing.T) {
	r := KillSearch("kill")
	if len(r) == 0 {
		t.Fatal("kill should list processes without filter")
	}
	if r[0].NavigateQuery == "" || r[0].Action != nil {
		t.Fatalf("kill list row should navigate to confirmation: %+v", r[0])
	}
	if r := KillSearch("killer"); r != nil {
		t.Fatal("non kill query should return nil")
	}
}

func TestSSHSearchReturnsPrefixFeedback(t *testing.T) {
	r := SSHSearch("ssh prod")
	if len(r) == 0 {
		t.Fatal("ssh prefix should return feedback instead of falling through")
	}
}

func TestGenUUID(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	for i := 0; i < 100; i++ {
		if u := genUUID(); !re.MatchString(u) {
			t.Fatalf("invalid uuid v4: %q", u)
		}
	}
}
