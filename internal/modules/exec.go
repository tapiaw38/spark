package modules

import "os/exec"

func Open(path string) bool {
	return Start("xdg-open", path)
}

func Start(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		SetStatus(false, "Failed to run "+name+": "+err.Error())
		return false
	}
	go func() {
		_ = cmd.Wait()
	}()
	return true
}

func Run(name string, args ...string) bool {
	if err := exec.Command(name, args...).Run(); err != nil {
		SetStatus(false, "Failed to run "+name+": "+err.Error())
		return false
	}
	return true
}
