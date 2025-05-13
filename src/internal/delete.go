package internal

import (
    "bufio"
    "os"
    "os/exec"
    "fmt"
    "path/filepath"
    "strings"
)

func DeleteProject(domain string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Are you sure you want to delete domain '%s'? [y/N]: ", domain)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "y" {
		fmt.Println("Aborted.")
		return
	}

	projectPath := filepath.Join(ProjectDirPrefix, domain)
	composeFile := filepath.Join(projectPath, DockerComposeFile)

	// Stop containers if the project exists
	if _, err := os.Stat(composeFile); err == nil {
		fmt.Println("Stopping containers for", domain, "...")
		if err := runDockerComposeDown(projectPath); err != nil {
			fmt.Println("Warning: failed to stop containers:", err)
		} else {
			fmt.Println("Containers stopped successfully.")
		}
	} else {
		fmt.Println("No docker-compose file found, skipping container shutdown.")
	}

	err := os.RemoveAll(projectPath)
	if err != nil {
		fmt.Println("Remove Domain failed, retrying with sudo rm -rf")

		cmd := exec.Command("sudo", "rm", "-rf", projectPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println("Hard delete failed (sudo):", err)
		} else {
			fmt.Println("Deleted folder with sudo rm -rf:", projectPath)
		}
	} else {
		fmt.Println("Deleted domain folder:", projectPath)
	}

	certDir := filepath.Join(CertsDir, domain)
	if err := os.RemoveAll(certDir); err == nil {
		fmt.Println("Deleted domain certs folder:", certDir)
	} else {
		fmt.Println("Failed to delete cert folder:", err)
	}

	// Use -NoProfile and -ExecutionPolicy Bypass for more efficient PowerShell execution
	checkCmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command",
		fmt.Sprintf(`certutil -store Root ^| Select-String "%s"; exit`, domain))

	output, err := checkCmd.CombinedOutput()
	// Force garbage collection for PowerShell process
	if checkCmd.Process != nil {
		checkCmd.Process.Kill()
	}
	
	if err != nil {
		fmt.Printf("Warning: failed to check Windows cert store: %v\n", err)
	} else if strings.Contains(string(output), domain) {
		fmt.Println("Removing domain cert from Windows Root store...")

		delCmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command",
			fmt.Sprintf(`Start-Process powershell -Verb runAs -ArgumentList '-NoProfile','-ExecutionPolicy','Bypass','-Command','certutil -delstore Root "%s"; exit'`, domain))
		delCmd.Stdin = os.Stdin
		delCmd.Stdout = os.Stdout
		delCmd.Stderr = os.Stderr

		err := delCmd.Run()
		// Force garbage collection for PowerShell process
		if delCmd.Process != nil {
			delCmd.Process.Kill()
		}
		
		if err != nil {
			fmt.Println("Failed to remove trusted cert:", err)
		} else {
			fmt.Println("Removed domain cert from Windows Root store.")
		}
	} else {
		fmt.Println("Domain cert not found in Windows Root store — skipping removal.")
	}

	lines, err := os.ReadFile(IPMapPath)
	if err == nil {
		var kept []string
		for _, line := range strings.Split(string(lines), "\n") {
			if !strings.HasPrefix(line, domain+"=") && !strings.HasPrefix(line, domain+"_") {
				kept = append(kept, line)
			}
		}
		os.WriteFile(IPMapPath, []byte(strings.Join(kept, "\n")), 0644)
		fmt.Println("Updated:", IPMapPath)
	}

	sitePath := filepath.Join(SharedServicesDir, SitesDir, domain+".conf")
	siteConfigRemoved := false
	if err := os.Remove(sitePath); err == nil {
		fmt.Println("Removed reverse proxy config:", sitePath)
		siteConfigRemoved = true
	}

	// Remove domain from Windows hosts file
	if err := removeFromWindowsHosts(domain, WindowsHostsPath); err != nil {
		fmt.Println("Warning: failed to update Windows hosts file:", err)
	}

	// Restart Nginx reverse proxy if site config was removed
	if siteConfigRemoved {
		if err := restartNginxReverseProxy(); err != nil {
			fmt.Println("Warning: failed to restart Nginx reverse proxy:", err)
		}
	}

	fmt.Println("Domain", domain, "was successfully deleted.")
}
