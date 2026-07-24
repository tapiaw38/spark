package modules

import (
	"regexp"
	"strconv"
	"strings"
)

var unitRe = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)\s*([a-z°]+)\s+(?:to|in)\s+([a-z°]+)$`)

// factors maps a unit to (base-unit, multiplier-to-base) within a category.
var unitFactors = map[string]struct {
	base   string
	factor float64
}{
	"mm": {"m", 0.001}, "cm": {"m", 0.01}, "m": {"m", 1}, "km": {"m", 1000},
	"in": {"m", 0.0254}, "ft": {"m", 0.3048}, "yd": {"m", 0.9144}, "mi": {"m", 1609.344},
	"mg": {"g", 0.001}, "g": {"g", 1}, "kg": {"g", 1000}, "oz": {"g", 28.3495}, "lb": {"g", 453.592},
	"b": {"b", 1}, "kb": {"b", 1024}, "mb": {"b", 1048576}, "gb": {"b", 1073741824}, "tb": {"b", 1099511627776},
}

// UnitSearch converts "100 km to mi", "50f to c", "5 gb to mb".
func UnitSearch(query string) []Result {
	m := unitRe.FindStringSubmatch(strings.ToLower(strings.TrimSpace(query)))
	if m == nil {
		return nil
	}
	val, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return nil
	}
	res, ok := convertUnit(val, m[2], m[3])
	if !ok {
		return nil
	}
	str := formatNumber(res)
	return []Result{{
		Type:   "calc",
		Title:  str + " " + m[3],
		Desc:   "Copy to clipboard",
		Icon:   "accessories-calculator",
		Action: func() { copyToClipboard(str) },
	}}
}

func convertUnit(val float64, from, to string) (float64, bool) {
	if t, ok := convertTemp(val, from, to); ok {
		return t, true
	}
	f, okF := unitFactors[from]
	t, okT := unitFactors[to]
	if !okF || !okT || f.base != t.base {
		return 0, false
	}
	return val * f.factor / t.factor, true
}

func convertTemp(val float64, from, to string) (float64, bool) {
	norm := func(u string) string { return strings.TrimPrefix(strings.TrimSuffix(u, "°"), "°") }
	from, to = norm(from), norm(to)
	temps := map[string]bool{"c": true, "f": true, "k": true}
	if !temps[from] || !temps[to] {
		return 0, false
	}
	var c float64
	switch from {
	case "c":
		c = val
	case "f":
		c = (val - 32) * 5 / 9
	case "k":
		c = val - 273.15
	}
	switch to {
	case "c":
		return c, true
	case "f":
		return c*9/5 + 32, true
	case "k":
		return c + 273.15, true
	}
	return 0, false
}
