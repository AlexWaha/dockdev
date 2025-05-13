package internal

import (
    "bufio"
    "os"
    "fmt"
    "strings"
)

func updateWindowsHosts(domain, path string) error {
	hostsEntry := fmt.Sprintf("127.0.0.1 %s", domain)

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), domain) {
			fmt.Println("Hosts entry already exists.")
			return nil
		}
	}

	if _, err := file.WriteString("\n" + hostsEntry + "\n"); err != nil {
		return err
	}

	fmt.Println("Domain added to Windows hosts file.")
	return nil
}
