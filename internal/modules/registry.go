package modules

type Provider struct {
	Search   func(string) []Result
	Terminal bool
}

var providers = []Provider{
	{Search: NavigationSearch, Terminal: true},
	{Search: DestinationPickerSearch, Terminal: true},
	{Search: FileOperationSearch, Terminal: true},
	{Search: ShellSearch},
	{Search: HelpSearch},
	{Search: LargeTypeSearch},
	{Search: RecentSearch},
	{Search: ContactsSearch},
	{Search: EmailSearch},
	{Search: StatsSearch},
	{Search: SyncSearch},
	{Search: StatusSearch},
	{Search: SnippetSearch},
	{Search: DictionarySearch},
	{Search: SpellSearch},
	{Search: CalcSearch},
	{Search: ClipboardSearch},
	{Search: WebSearch},
	{Search: SystemSearch},
	{Search: SpotifySearch},
	{Search: MusicSearch},
	{Search: FileBufferSearch},
	{Search: FileSearch},
}

func SearchAll(query string) (results []Result, terminal bool) {
	for _, p := range providers {
		found := p.Search(query)
		if found == nil {
			continue
		}
		if p.Terminal {
			return found, true
		}
		results = append(results, found...)
	}
	return results, false
}
