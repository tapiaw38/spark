package modules

import "testing"

func titled(titles ...string) SearchFunc {
	return func(string) []Result {
		out := make([]Result, 0, len(titles))
		for _, t := range titles {
			out = append(out, Result{Title: t})
		}
		return out
	}
}

func titlesOf(results []Result) []string {
	out := make([]string, 0, len(results))
	for _, r := range results {
		out = append(out, r.Title)
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRegistryPreservesPriorityOrder(t *testing.T) {
	reg := NewRegistry(titled("first"), titled("second", "third"))

	got := titlesOf(reg.Search("anything"))
	want := []string{"first", "second", "third"}
	if !equal(got, want) {
		t.Errorf("Search() = %v, want %v", got, want)
	}
}

func TestRegistrySkipsNonMatchingModules(t *testing.T) {
	silent := SearchFunc(func(string) []Result { return nil })
	reg := NewRegistry(silent, titled("only"), silent)

	if got := titlesOf(reg.Search("q")); !equal(got, []string{"only"}) {
		t.Errorf("Search() = %v, want [only]", got)
	}
}

func TestRegistryFallbackOnlyWhenEmpty(t *testing.T) {
	fallback := titled("fallback")

	empty := NewRegistry(SearchFunc(func(string) []Result { return nil })).WithFallback(fallback)
	if got := titlesOf(empty.Search("q")); !equal(got, []string{"fallback"}) {
		t.Errorf("empty registry Search() = %v, want [fallback]", got)
	}

	matched := NewRegistry(titled("hit")).WithFallback(fallback)
	if got := titlesOf(matched.Search("q")); !equal(got, []string{"hit"}) {
		t.Errorf("matched registry Search() = %v, want [hit] (fallback must not run)", got)
	}
}

func TestRegistryAppendGoesLast(t *testing.T) {
	reg := NewRegistry(titled("first")).Append(titled("appended"))

	if got := titlesOf(reg.Search("q")); !equal(got, []string{"first", "appended"}) {
		t.Errorf("Search() = %v, want [first appended]", got)
	}
}

func TestDefaultSearchersAreWired(t *testing.T) {
	if got := len(DefaultRegistry().searchers); got == 0 {
		t.Fatal("DefaultRegistry() has no searchers")
	}
	for _, s := range DefaultRegistry().searchers {
		if s == nil {
			t.Error("DefaultRegistry() contains a nil searcher")
		}
	}
	for _, a := range DefaultAsyncSearchers() {
		if a.Match == nil || a.Loading == nil || a.Search == nil {
			t.Errorf("async searcher %q has a nil Match/Loading/Search", a.Name)
		}
	}
}
