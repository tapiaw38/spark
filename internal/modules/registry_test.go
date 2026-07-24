package modules

import "testing"

func withProviders(t *testing.T, p []Provider) {
	orig := providers
	t.Cleanup(func() { providers = orig })
	providers = p
}

func TestSearchAll_OrderAndAccumulate(t *testing.T) {
	withProviders(t, []Provider{
		{Search: func(string) []Result { return []Result{{Title: "a"}} }},
		{Search: func(string) []Result { return nil }},
		{Search: func(string) []Result { return []Result{{Title: "b"}, {Title: "c"}} }},
	})

	results, terminal := SearchAll("q")
	if terminal {
		t.Fatal("expected non-terminal")
	}
	got := []string{}
	for _, r := range results {
		got = append(got, r.Title)
	}
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order: got %v, want %v", got, want)
		}
	}
}

func TestSearchAll_TerminalShortCircuits(t *testing.T) {
	reached := false
	withProviders(t, []Provider{
		{Search: func(string) []Result { return []Result{{Title: "nav"}} }, Terminal: true},
		{Search: func(string) []Result { reached = true; return []Result{{Title: "after"}} }},
	})

	results, terminal := SearchAll("q")
	if !terminal {
		t.Fatal("expected terminal")
	}
	if reached {
		t.Fatal("provider after terminal must not run")
	}
	if len(results) != 1 || results[0].Title != "nav" {
		t.Fatalf("got %v, want [nav]", results)
	}
}

func TestSearchAll_TerminalNilFallsThrough(t *testing.T) {
	withProviders(t, []Provider{
		{Search: func(string) []Result { return nil }, Terminal: true},
		{Search: func(string) []Result { return []Result{{Title: "x"}} }},
	})

	results, terminal := SearchAll("q")
	if terminal {
		t.Fatal("nil terminal must not short-circuit")
	}
	if len(results) != 1 || results[0].Title != "x" {
		t.Fatalf("got %v, want [x]", results)
	}
}
