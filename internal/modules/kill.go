package modules

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// KillSearch lists running processes matching a name and sends SIGTERM.
func KillSearch(query string) []Result {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower != "kill" && !strings.HasPrefix(lower, "kill ") {
		return nil
	}
	if strings.HasPrefix(lower, "kill confirm ") {
		pid := strings.TrimSpace(query[len("kill confirm "):])
		proc, ok := procByPID(pid)
		if !ok {
			return []Result{{
				Type:   "kill",
				Title:  "Process not found: " + pid,
				Desc:   "It may have already exited",
				Icon:   "dialog-warning",
				Action: func() {},
			}}
		}
		return []Result{confirmKillResult(proc)}
	}
	name := strings.TrimSpace(query[len("kill"):])

	var out []Result
	for _, p := range listProcs(strings.ToLower(name)) {
		p := p
		out = append(out, Result{
			Type:          "kill",
			Title:         "Kill " + p.name + " (" + p.pid + ")",
			Desc:          "Press Enter to confirm",
			Icon:          "process-stop",
			NavigateQuery: "kill confirm " + p.pid,
		})
		if len(out) >= 8 {
			break
		}
	}
	return out
}

type procInfo struct{ pid, name string }

func confirmKillResult(p procInfo) Result {
	return Result{
		Type:  "kill",
		Title: "Confirm Kill " + p.name + " (" + p.pid + ")",
		Desc:  "Send SIGTERM",
		Icon:  "dialog-warning",
		Action: func() {
			if err := exec.Command("kill", p.pid).Run(); err != nil {
				SetStatus(false, "Kill failed: "+err.Error())
			} else {
				SetStatus(true, "Killed "+p.name+" ("+p.pid+")")
			}
		},
	}
}

func procByPID(pid string) (procInfo, bool) {
	for _, p := range listProcs("") {
		if p.pid == pid {
			return p, true
		}
	}
	return procInfo{}, false
}

func listProcs(match string) []procInfo {
	out, err := exec.Command("ps", "-eo", "pid=,comm=").Output()
	if err != nil {
		return nil
	}
	self := strconv.Itoa(os.Getpid())

	var procs []procInfo
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, name := fields[0], strings.Join(fields[1:], " ")
		if pid == self {
			continue
		}
		if match != "" && !strings.Contains(strings.ToLower(name), match) {
			continue
		}
		procs = append(procs, procInfo{pid, name})
	}
	return procs
}
