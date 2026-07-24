package modules

import "strings"

var emojiTable = []struct{ char, keywords string }{
	{"😀", "grinning smile happy"},
	{"😃", "smiley happy joy"},
	{"😄", "laughing happy"},
	{"😁", "grin beaming"},
	{"😂", "joy laugh tears funny lol"},
	{"🤣", "rofl rolling laughing"},
	{"🙂", "slight smile"},
	{"😉", "wink"},
	{"😊", "blush smile happy"},
	{"😍", "heart eyes love"},
	{"😘", "kiss love"},
	{"😎", "cool sunglasses"},
	{"🤔", "thinking hmm"},
	{"😐", "neutral meh"},
	{"🙄", "eye roll"},
	{"😴", "sleep tired"},
	{"😭", "cry sob sad tears"},
	{"😢", "cry sad tear"},
	{"😡", "angry mad rage"},
	{"🤯", "mind blown exploding"},
	{"😱", "scream shock fear"},
	{"🥳", "party celebrate"},
	{"😅", "sweat nervous laugh"},
	{"🙃", "upside down"},
	{"😇", "angel halo innocent"},
	{"🤗", "hug"},
	{"🤞", "fingers crossed luck"},
	{"👍", "thumbs up yes good like"},
	{"👎", "thumbs down no bad dislike"},
	{"👏", "clap applause"},
	{"🙏", "pray thanks please"},
	{"💪", "muscle strong flex"},
	{"👀", "eyes look"},
	{"🔥", "fire lit hot"},
	{"⭐", "star"},
	{"✨", "sparkles shiny"},
	{"🎉", "party tada celebrate"},
	{"❤️", "heart love red"},
	{"💔", "broken heart"},
	{"💯", "hundred perfect"},
	{"✅", "check done ok yes"},
	{"❌", "cross no wrong error"},
	{"⚠️", "warning caution"},
	{"🚀", "rocket launch ship fast"},
	{"💡", "idea lightbulb"},
	{"📌", "pin"},
	{"🐛", "bug insect"},
	{"💩", "poop"},
	{"🤖", "robot bot"},
	{"👋", "wave hello hi bye"},
	{"🎂", "cake birthday"},
	{"☕", "coffee tea"},
	{"🍺", "beer drink"},
	{"🌙", "moon night"},
	{"☀️", "sun sunny"},
	{"🌧️", "rain cloud"},
	{"⏰", "alarm clock time"},
	{"💰", "money bag cash"},
	{"📈", "chart up growth"},
	{"🎯", "target dart goal"},
}

// EmojiSearch handles "emoji <query>" and copies the chosen emoji.
func EmojiSearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "emoji" && !strings.HasPrefix(q, "emoji ") {
		return nil
	}
	term := strings.TrimSpace(strings.TrimPrefix(q, "emoji"))

	var out []Result
	for _, e := range emojiTable {
		if term != "" && !strings.Contains(e.keywords, term) {
			continue
		}
		e := e
		out = append(out, Result{
			Type:     "emoji",
			Title:    firstWord(e.keywords),
			Desc:     "Copy emoji",
			IconText: e.char,
			KeepOpen: false,
			Action:   func() { copyToClipboard(e.char) },
		})
		if len(out) >= 8 {
			break
		}
	}
	return out
}

func firstWord(s string) string {
	w, _, _ := strings.Cut(s, " ")
	return w
}
