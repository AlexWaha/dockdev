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
	sharedServicesDir := "shared-services"

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Are you sure you want to delete domain '%s'? [y/N]: ", domain)
	answer, _ := reader.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer != "y" {
		fmt.Println("Aborted.")
		return
	}

	projectPath := filepath.Join("domains", domain)
	composeFile := filepath.Join(projectPath, "docker-compose.yml")

	if _, err := os.Stat(composeFile); err == nil {
		fmt.Println("Stopping containers for", domain, "...")
		stopCmd := exec.Command("docker", "compose", "down")
		stopCmd.Dir = projectPath
		stopCmd.Stdout = os.Stdout
		stopCmd.Stderr = os.Stderr
		if err := stopCmd.Run(); err != nil {
			fmt.Println("Warning: failed to stop containers:", err)
		}
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

	certDir := filepath.Join("shared-services", "certs", domain)
	if err := os.RemoveAll(certDir); err == nil {
		fmt.Println("Deleted domain certs folder:", certDir)
	} else {
		fmt.Println("Failed to delete cert folder:", err)
	}

	checkCmd := exec.Command("powershell.exe", "-Command",
	fmt.Sprintf(`certutil -store Root ^| Select-String "%s"`, domain))

	output, err := checkCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Warning: failed to check Windows cert store: %v\n", err)
	} else if strings.Contains(string(output), domain) {
		fmt.Println("Removing domain cert from Windows Root store...")

		delCmd := exec.Command("powershell.exe", "-Command",
            fmt.Sprintf(`Start-Process powershell -Verb runAs -ArgumentList 'certutil -delstore Root "%s"'`, domain))
		delCmd.Stdin = os.Stdin
		delCmd.Stdout = os.Stdout
		delCmd.Stderr = os.Stderr

		if err := delCmd.Run(); err != nil {
			fmt.Println("Failed to remove trusted cert:", err)
		} else {
			fmt.Println("Removed domain cert from Windows Root store.")
		}
	} else {
		fmt.Println("Domain cert not found in Windows Root store — skipping removal.")
	}

	ipmap := ".ipmap.env"
	lines, err := os.ReadFile(ipmap)
	if err == nil {
		var kept []string
		for _, line := range strings.Split(string(lines), "\n") {
			if !strings.HasPrefix(line, domain+"=") && !strings.HasPrefix(line, domain+"_") {
				kept = append(kept, line)
			}
		}
		os.WriteFile(ipmap, []byte(strings.Join(kept, "\n")), 0644)
		fmt.Println("Updated:", ipmap)
	}

	sitePath := filepath.Join(sharedServicesDir, "sites", domain+".conf")
	if err := os.Remove(sitePath); err == nil {
		fmt.Println("Removed reverse proxy config:", sitePath)
	}

	hosts := "/mnt/c/Windows/System32/drivers/etc/hosts"
	hfile, err := os.ReadFile(hosts)
	if err == nil {
		var out []string
		for _, line := range strings.Split(string(hfile), "\n") {
			if !strings.Contains(line, domain) {
				out = append(out, line)
			}
		}
		os.WriteFile(hosts, []byte(strings.Join(out, "\n")), 0644)
		fmt.Println("Updated Windows hosts file.")
	}

	fmt.Println("Domain", domain, "was successfully deleted.")
}
