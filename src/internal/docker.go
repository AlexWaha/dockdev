package internal

import (
    "os"
    "os/exec"
)

func runDockerComposeUp(dir string) error {
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}


func runDockerComposeDown(dir string) error {
    cmd := exec.Command("docker", "compose", "down")
    cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
