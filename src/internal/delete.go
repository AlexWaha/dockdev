package internal

import (
    "os"
    "os/exec"
    "fmt"
    "path/filepath"
    "strings"
)

func DeleteProject(domain string) {
	PrintSectionDivider("DELETING PROJECT: " + domain)

	if IsTerminal() {
		// Use YesNoPrompt for interactive confirmation
		if !YesNoPrompt(fmt.Sprintf("Are you sure you want to delete domain '%s'?", Bold(domain)), false) {
			fmt.Println(Info("Aborted."))
			return
		}
	} else {
		// Non-interactive mode always proceeds without confirmation
		fmt.Printf(Info("Deleting domain '%s'...\n"), Bold(domain))
	}

	// Ensure Docker is running if we need to stop containers
	projectPath := filepath.Join(ProjectDirPrefix, domain)
	composeFile := filepath.Join(projectPath, DockerComposeFile)
	
	PrintDivider()
	fmt.Println(Bold("STEP 1: Stopping containers"))
	
	// Stop containers if the project exists
	if _, err := os.Stat(composeFile); err == nil {
		// Check if Docker is running before trying to stop containers
		if err := EnsureDockerRunning(); err != nil {
			fmt.Printf(Warning("Warning: Docker is not available: %v\n"), err)
			fmt.Println(Info("Skipping container shutdown, will continue with deletion."))
		} else {
			fmt.Println(Highlight("Stopping containers for"), Bold(domain), Highlight("..."))
			if err := runDockerComposeDown(projectPath); err != nil {
				fmt.Println(Warning("Warning: failed to stop containers:"), Error(err.Error()))
			} else {
				fmt.Println(Success("Containers stopped successfully."))
			}
		}
	} else {
		fmt.Println(Info("No docker-compose file found, skipping container shutdown."))
	}

	PrintDivider()
	fmt.Println(Bold("STEP 2: Removing project files"))
	
	err := os.RemoveAll(projectPath)
	if err != nil {
		fmt.Println(Warning("Remove Domain failed, retrying with sudo rm -rf"))

		cmd := exec.Command("sudo", "rm", "-rf", projectPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println(Error("Hard delete failed (sudo):"), Error(err.Error()))
		} else {
			fmt.Println(Success("Deleted folder with sudo rm -rf:"), Info(projectPath))
		}
	} else {
		fmt.Println(Success("Deleted domain folder:"), Info(projectPath))
	}

	PrintDivider()
	fmt.Println(Bold("STEP 3: Removing SSL certificates"))
	
	certDir := filepath.Join(CertsDir, domain)
	if err := os.RemoveAll(certDir); err == nil {
		fmt.Println(Success("Deleted domain certs folder:"), Info(certDir))
	} else {
		fmt.Println(Error("Failed to delete cert folder:"), Error(err.Error()))
	}

	PrintDivider()
	fmt.Println(Bold("STEP 4: Removing Windows certificates"))
	
	// Use -NoProfile and -ExecutionPolicy Bypass for more efficient PowerShell execution
	checkCmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command",
		fmt.Sprintf(`certutil -store Root ^| Select-String "%s"; exit`, domain))

	output, err := checkCmd.CombinedOutput()
	// Force garbage collection for PowerShell process
	if checkCmd.Process != nil {
		checkCmd.Process.Kill()
	}
	
	if err != nil {
		fmt.Printf(Warning("Warning: failed to check Windows cert store: %v\n"), err)
	} else if strings.Contains(string(output), domain) {
		fmt.Println(Highlight("Removing domain cert from Windows Root store..."))

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
			fmt.Println(Error("Failed to remove trusted cert:"), Error(err.Error()))
		} else {
			fmt.Println(Success("Removed domain cert from Windows Root store."))
		}
	} else {
		fmt.Println(Info("Domain cert not found in Windows Root store — skipping removal."))
	}

	PrintDivider()
	fmt.Println(Bold("STEP 5: Updating configuration files"))
	
	lines, err := os.ReadFile(IPMapPath)
	if err == nil {
		var kept []string
		for _, line := range strings.Split(string(lines), "\n") {
			if !strings.HasPrefix(line, domain+"=") && !strings.HasPrefix(line, domain+"_") {
				kept = append(kept, line)
			}
		}
		os.WriteFile(IPMapPath, []byte(strings.Join(kept, "\n")), 0644)
		fmt.Println(Success("Updated:"), Info(IPMapPath))
	}

	sitePath := filepath.Join(SharedServicesDir, SitesDir, domain+".conf")
	siteConfigRemoved := false
	if err := os.Remove(sitePath); err == nil {
		fmt.Println(Success("Removed reverse proxy config:"), Info(sitePath))
		siteConfigRemoved = true
	}

	// Remove domain from Windows hosts file
	if err := removeFromWindowsHosts(domain, WindowsHostsPath); err != nil {
		fmt.Println(Warning("Warning: failed to update Windows hosts file:"), Error(err.Error()))
	}

	PrintDivider()
	fmt.Println(Bold("STEP 6: Restarting services"))
	
	// Restart Nginx reverse proxy if site config was removed
	if siteConfigRemoved {
		// Check if Docker is running before trying to restart Nginx
		if err := CheckDockerRunning(); err != nil {
			fmt.Println(Warning("Warning: Docker is not available, skipping Nginx restart."))
		} else {
			if err := restartNginxReverseProxy(); err != nil {
				fmt.Println(Warning("Warning: failed to restart Nginx reverse proxy:"), Error(err.Error()))
			}
		}
	}

	PrintSectionDivider("OPERATION COMPLETE")
	fmt.Println(Success("Domain"), Bold(domain), Success("was successfully deleted."))
}
