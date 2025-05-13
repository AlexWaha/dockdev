package internal

import (
    "fmt"
    "os"
    "os/exec"
    "time"
)

func runDockerComposeUp(dir string) error {
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runDockerComposeDown(dir string) error {
    // Verify directory exists before trying to run docker compose
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        return fmt.Errorf("directory does not exist: %s", dir)
    }
    
    // Verify docker-compose.yml exists in the directory
    composeFile := fmt.Sprintf("%s/%s", dir, DockerComposeFile)
    if _, err := os.Stat(composeFile); os.IsNotExist(err) {
        return fmt.Errorf("docker-compose file not found: %s", composeFile)
    }
    
    cmd := exec.Command("docker", "compose", "down")
    cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// restartNginxReverseProxy attempts to reload the Nginx configuration.
// If reload fails, it will restart the container.
func restartNginxReverseProxy() error {
    fmt.Println("Reloading reverse proxy configuration...")
    
    // First try to reload Nginx configuration
    reloadCmd := exec.Command("docker", "exec", ReverseProxyName, "nginx", "-s", "reload")
    reloadCmd.Stdout = os.Stdout
    reloadCmd.Stderr = os.Stderr
    err := reloadCmd.Run()
    
    if err == nil {
        fmt.Println("Nginx configuration reloaded successfully.")
        return nil
    }
    
    // If reload fails, try to restart the container
    fmt.Println("Reload failed, restarting Nginx container...")
    restartCmd := exec.Command("docker", "restart", ReverseProxyName)
    restartCmd.Stdout = os.Stdout
    restartCmd.Stderr = os.Stderr
    err = restartCmd.Run()
    
    if err != nil {
        return fmt.Errorf("failed to restart Nginx container: %w", err)
    }
    
    // Give the container a moment to restart
    time.Sleep(2 * time.Second)
    fmt.Println("Nginx container restarted successfully.")
    return nil
}
