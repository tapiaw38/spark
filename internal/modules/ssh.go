package modules

import (
	"bufio"
	"os"
	"strings"

	"github.com/tapiaw38/spark/internal/config"
)

func SSHSearch(query string) []Result {
	arg, ok := MatchCommand(query, "ssh")
	if !ok {
		return nil
	}
	filter := strings.ToLower(arg)

	var out []Result
	hosts, err := sshHosts()
	if err != nil {
		return []Result{{
			Type:  TypeSSH,
			Title: "No SSH config",
			Desc:  "Create ~/.ssh/config with Host entries",
			Icon:  "dialog-warning",
		}}
	}
	for _, host := range hosts {
		if filter != "" && !strings.Contains(strings.ToLower(host), filter) {
			continue
		}
		out = append(out, Result{
			Type:       TypeSSH,
			Title:      "SSH: " + host,
			Desc:       "Connect in terminal",
			Icon:       "utilities-terminal",
			ActionSpec: TerminalAction("ssh " + shellQuote(host)),
		})
		if len(out) >= MaxCompactResults {
			break
		}
	}
	if len(out) == 0 {
		return []Result{{
			Type:  TypeSSH,
			Title: "No SSH host: " + filter,
			Desc:  "Add Host " + filter + " to ~/.ssh/config",
			Icon:  "dialog-warning",
		}}
	}
	return out
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func sshHosts() ([]string, error) {
	f, err := os.Open(config.HomeFile(".ssh", "config"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hosts []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(strings.ToLower(line), "host ") {
			continue
		}
		for _, name := range strings.Fields(line)[1:] {
			if strings.ContainsAny(name, "*?") {
				continue
			}
			hosts = append(hosts, name)
		}
	}
	return hosts, nil
}
