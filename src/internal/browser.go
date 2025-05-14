package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OpenBrowser opens the specified URL in the default Windows browser
func OpenBrowser(url string) error {
	PrintDivider()
	fmt.Println(Bold("OPENING PROJECT IN BROWSER"))

	// Make sure URL has a scheme
	if url != "" && !hasScheme(url) {
		url = "https://" + url
	}

	fmt.Println(Info("Opening URL in default browser:"), Highlight(url))

	// Use PowerShell to open the default browser
	// Using Start-Process for better process management
	cmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command",
		fmt.Sprintf(`Start-Process "%s"`, url))

	err := cmd.Run()
	if err != nil {
		fmt.Println(Error("Failed to open browser:"), Error(err.Error()))
		return err
	}

	fmt.Println(Success("Browser launched successfully."))
	return nil
}

// hasScheme checks if a URL has a scheme (http:// or https://)
func hasScheme(url string) bool {
	for i := 0; i < len(url); i++ {
		if url[i] == ':' {
			return true
		}
		if url[i] == '/' || url[i] == ' ' {
			return false
		}
	}
	return false
}

// GetProjectURL returns the complete URL for a project domain
func GetProjectURL(domain string) string {
    if IsSSLEnabledForDomain(domain) {
        return "https://" + domain
    }
    return "http://" + domain
}

// IsSSLEnabledForDomain checks if SSL is enabled for a given domain based on its nginx config
func IsSSLEnabledForDomain(domain string) bool {
    confPath := filepath.Join(SharedServicesDir, SitesDir, domain+".conf")

    file, err := os.Open(confPath)
    if err != nil {
        return false
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if strings.HasPrefix(line, "listen 443 ssl;") {
            return true
        }
    }

    return false
}
