package modules

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// WindowSearch lists open toplevel windows (via wlrctl) and focuses one.
func WindowSearch(query string) []Result {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower != "w" && !strings.HasPrefix(lower, "w ") {
		return nil
	}
	if _, err := exec.LookPath("wlrctl"); err != nil {
		return mangoWorkspaceSearch(lower)
	}
	filter := lower
	for _, p := range []string{"w ", "w"} {
		if strings.HasPrefix(lower, p) {
			filter = strings.TrimSpace(lower[len(p):])
			break
		}
	}

	out, err := exec.Command("wlrctl", "toplevel", "list").Output()
	if err != nil {
		return nil
	}

	var results []Result
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(line), filter) {
			continue
		}
		appID, title, _ := strings.Cut(line, ":")
		appID, title = strings.TrimSpace(appID), strings.TrimSpace(title)
		matcher := "app_id:" + appID
		if title != "" {
			matcher = "title:" + title
		}
		results = append(results, Result{
			Type:   "window",
			Title:  line,
			Desc:   "Focus window",
			Icon:   "preferences-system-windows",
			Action: func() { exec.Command("wlrctl", "toplevel", "focus", matcher).Run() },
		})
		if len(results) >= 10 {
			break
		}
	}
	return results
}

type mangoTag struct {
	output  string
	id      int
	state   int
	clients int
	focused int
}

func mangoWorkspaceSearch(query string) []Result {
	if _, err := exec.LookPath("mmsg"); err != nil {
		return []Result{{
			Type:   "window",
			Title:  "Window switcher unavailable",
			Desc:   "Install wlrctl or use MangoWM with mmsg",
			Icon:   "dialog-warning",
			Action: func() {},
		}}
	}

	filter := workspaceFilter(query)
	tags := mangoTags()
	if len(tags) == 0 {
		return []Result{{
			Type:   "workspace",
			Title:  "No MangoWM workspaces",
			Desc:   "mmsg did not return tags",
			Icon:   "view-grid",
			Action: func() {},
		}}
	}

	var results []Result
	for _, tag := range tags {
		tag := tag
		name := strconv.Itoa(tag.id)
		if filter != "" && !strings.Contains(name, filter) && !strings.Contains(strings.ToLower(tag.output), filter) {
			continue
		}
		title := "Workspace " + name
		if tag.focused != 0 {
			title = "Current: " + title
		}
		desc := fmt.Sprintf("%s | clients: %d", tag.output, tag.clients)
		if tag.state != 0 {
			desc += " | occupied"
		}
		results = append(results, Result{
			Type:  "workspace",
			Title: title,
			Desc:  desc,
			Icon:  "view-grid",
			Action: func() {
				if err := exec.Command("mmsg", "-s", "-t", strconv.Itoa(tag.id)).Run(); err != nil {
					SetStatus(false, "Workspace switch failed: "+err.Error())
					return
				}
				SetStatus(true, "Workspace: "+strconv.Itoa(tag.id))
			},
		})
	}
	return results
}

func workspaceFilter(query string) string {
	for _, p := range []string{"w ", "w"} {
		if strings.HasPrefix(query, p) {
			return strings.TrimSpace(query[len(p):])
		}
	}
	return ""
}

func mangoTags() []mangoTag {
	out, err := exec.Command("mmsg", "-g", "-t").Output()
	if err != nil {
		return nil
	}
	var tags []mangoTag
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 6 || fields[1] != "tag" {
			continue
		}
		id, errID := strconv.Atoi(fields[2])
		state, errState := strconv.Atoi(fields[3])
		clients, errClients := strconv.Atoi(fields[4])
		focused, errFocused := strconv.Atoi(fields[5])
		if errID != nil || errState != nil || errClients != nil || errFocused != nil {
			continue
		}
		tags = append(tags, mangoTag{
			output:  fields[0],
			id:      id,
			state:   state,
			clients: clients,
			focused: focused,
		})
	}
	return tags
}
