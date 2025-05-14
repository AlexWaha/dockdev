package internal

import (
    "fmt"
    "os"
    "os/exec"
    "time"
    "strings"
)

// CheckDockerRunning verifies if Docker is running and available
func CheckDockerRunning() error {
    cmd := exec.Command("docker", "info")
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        outputStr := string(output)
        if strings.Contains(outputStr, "Could not connect to the Docker daemon") || 
           strings.Contains(outputStr, "could not be found in this WSL 2 distro") {
            return fmt.Errorf("Docker is not running or not properly configured with WSL 2")
        }
        return fmt.Errorf("docker error: %s", outputStr)
    }
    
    return nil
}

// StartDockerDesktop attempts to start Docker Desktop on Windows
func StartDockerDesktop() error {
    PrintSectionDivider("STARTING DOCKER DESKTOP")
    fmt.Println(Highlight("Attempting to start Docker Desktop..."))
    
    // Path to Docker Desktop on Windows
    dockerPath := "C:\\Program Files\\Docker\\Docker\\Docker Desktop.exe"
    
    // Check if Docker Desktop exists at the expected path
    if _, err := os.Stat("/mnt/c/Program Files/Docker/Docker/Docker Desktop.exe"); os.IsNotExist(err) {
        return fmt.Errorf("Docker Desktop not found at expected location")
    }
    
    // Start Docker Desktop
    cmd := exec.Command("cmd.exe", "/c", "start", `"Docker Desktop"`, dockerPath)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to start Docker Desktop: %w", err)
    }
    
    PrintDivider()
    fmt.Println(Info("Docker Desktop is starting. Please wait..."))
    
    // Wait for Docker to become available
    maxAttempts := 30
    for i := 0; i < maxAttempts; i++ {
        time.Sleep(2 * time.Second)
        fmt.Printf(Info("Checking Docker availability (attempt %d/%d)...\n"), i+1, maxAttempts)
        
        if err := CheckDockerRunning(); err == nil {
            fmt.Println(Success("Docker is now running."))
            return nil
        }
    }
    
    return fmt.Errorf("Docker did not start within the expected time")
}

// EnsureDockerRunning checks if Docker is running and offers to start it if it's not
func EnsureDockerRunning() error {
    if err := CheckDockerRunning(); err == nil {
        // Docker is running
        return nil
    }
    
    fmt.Println(Warning("Docker is not running."))
    
    if IsTerminal() {
        if YesNoPrompt("Would you like to start Docker Desktop?", true) {
            return StartDockerDesktop()
        } else {
            return fmt.Errorf("Docker is required but not running. Operation cancelled")
        }
    } else {
        return fmt.Errorf("Docker is not running and cannot be started in non-interactive mode")
    }
}

func runDockerComposeUp(dir string) error {
	PrintDivider()
	fmt.Println(Bold("STARTING DOCKER SERVICES"))
	fmt.Println(Highlight("Starting services with docker-compose in"), Info(dir))
	
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(Error("Failed to start containers."))
		return err
	}
	fmt.Println(Success("Containers started successfully."))
	return nil
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
    
    PrintDivider()
    fmt.Println(Bold("STOPPING DOCKER SERVICES"))
    fmt.Println(Highlight("Running docker-compose down in"), Info(dir))
    
    // Create a custom writer that can filter Docker Compose warnings
    stdoutBuf := NewFilteredWriter(os.Stdout, func(line string) bool {
        return !strings.Contains(line, "Warning: No resource found to remove")
    })
    
    cmd := exec.Command("docker", "compose", "down")
    cmd.Dir = dir
	cmd.Stdout = stdoutBuf
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// restartNginxReverseProxy attempts to reload the Nginx configuration.
// If reload fails, it will restart the container.
func restartNginxReverseProxy() error {
    PrintDivider()
    fmt.Println(Bold("RESTARTING NGINX PROXY"))
    fmt.Println(Highlight("Reloading reverse proxy configuration..."))
    
    // First try to reload Nginx configuration
    reloadCmd := exec.Command("docker", "exec", ReverseProxyName, "nginx", "-s", "reload")
    reloadCmd.Stdout = os.Stdout
    reloadCmd.Stderr = os.Stderr
    err := reloadCmd.Run()
    
    if err == nil {
        fmt.Println(Success("Nginx configuration reloaded successfully."))
        return nil
    }
    
    // If reload fails, try to restart the container
    fmt.Println(Warning("Reload failed, restarting Nginx container..."))
    restartCmd := exec.Command("docker", "restart", ReverseProxyName)
    restartCmd.Stdout = os.Stdout
    restartCmd.Stderr = os.Stderr
    err = restartCmd.Run()
    
    if err != nil {
        return fmt.Errorf("failed to restart Nginx container: %w", err)
    }
    
    // Give the container a moment to restart
    time.Sleep(2 * time.Second)
    fmt.Println(Success("Nginx container restarted successfully."))
    return nil
}
