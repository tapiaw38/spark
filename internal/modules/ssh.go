package modules

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// SSHSearch lists hosts from ~/.ssh/config and connects in a terminal.
func SSHSearch(query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "ssh" && !strings.HasPrefix(q, "ssh ") {
		return nil
	}
	filter := strings.TrimSpace(strings.TrimPrefix(q, "ssh"))

	var out []Result
	hosts, err := sshHosts()
	if err != nil {
		return []Result{{
			Type:   "ssh",
			Title:  "No SSH config",
			Desc:   "Create ~/.ssh/config with Host entries",
			Icon:   "dialog-warning",
			Action: func() {},
		}}
	}
	for _, host := range hosts {
		if filter != "" && !strings.Contains(strings.ToLower(host), filter) {
			continue
		}
		host := host
		out = append(out, Result{
			Type:   "ssh",
			Title:  "SSH: " + host,
			Desc:   "Connect in terminal",
			Icon:   "utilities-terminal",
			Action: func() { openTerminal("ssh " + shellQuote(host)) },
		})
		if len(out) >= 8 {
			break
		}
	}
	if len(out) == 0 {
		return []Result{{
			Type:   "ssh",
			Title:  "No SSH host: " + filter,
			Desc:   "Add Host " + filter + " to ~/.ssh/config",
			Icon:   "dialog-warning",
			Action: func() {},
		}}
	}
	return out
}

func sshHosts() ([]string, error) {
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "config"))
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
