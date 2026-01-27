package utils

import "os/exec"

// runcommand executes a shell command
// args ...stirng - variadic parameter - allows 0 or more arguments
func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
