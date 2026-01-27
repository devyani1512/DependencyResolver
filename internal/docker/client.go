package docker

//os/exec is standard go library package for running external system commands
import (
	"os/exec"
)

// starts docker compose services
func StartServices(projectPath string) error {
	cmd := exec.Command("docker-compose", "up", "-d")
	cmd.Dir = projectPath // this line takes to the docker folder and then runs docker compose
	return cmd.Run()
}

// stops docker compose services
func StopServices(projectPath string) error {
	cmd := exec.Command("docker-compose", "down")
	cmd.Dir = projectPath
	return cmd.Run()
}
