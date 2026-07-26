package modules

func HelpSearch(query string) []Result {
	if !MatchesAny(query, "help", "?") {
		return nil
	}

	return []Result{
		helpResult("Apps", "type app name, e.g. slack"),
		helpResult("Files", "f query, then Tab for actions"),
		helpResult("Navigation", "nav / nav ~/Downloads"),
		helpResult("Destination Picker", "pick copy|move source | folder"),
		helpResult("File Ops", "rename/copy/move source | target"),
		helpResult("Undo", "undo last file op"),
		helpResult("Status", "status / last"),
		helpResult("File Buffer", "buffer / buf"),
		helpResult("Recent Documents", "recent query / recent app name"),
		helpResult("Large Type", "large text / lt text"),
		helpResult("Contacts", "contact name"),
		helpResult("Email", "email contact | subject | body"),
		helpResult("Usage Stats", "stats / usage"),
		helpResult("Sync Settings", "sync / sync import zip"),
		helpResult("Web", "g, yt, gh, wiki, ddg, r, so"),
		helpResult("Clipboard", "clip query / cb query / paste query"),
		helpResult("Snippets", ";keyword"),
		helpResult("Dictionary", "define word / def word"),
		helpResult("Spelling", "spell word / spelling word"),
		helpResult("Music", "sp controls / m local song"),
		helpResult("Weather", "wt city / weather city"),
		helpResult("Shell", "> command"),
	}
}

func helpResult(title, desc string) Result {
	return Result{
		Type:  TypeHelp,
		Title: title,
		Desc:  desc,
		Icon:  "help-browser",
	}
}
