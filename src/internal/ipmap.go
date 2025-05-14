package internal

import (
    "fmt"
    "os"
    "strings"
    "regexp"
    "sort"
)

func InsertIPMappingAtTop(filePath, key, ip string) error {
	entry := fmt.Sprintf("%s=%s", key, ip)

	content, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := strings.Split(string(content), "\n")

	var filtered []string
	for _, line := range lines {
		if !strings.HasPrefix(line, key+"=") && strings.TrimSpace(line) != "" {
			filtered = append(filtered, line)
		}
	}

	final := append([]string{entry}, filtered...)
	return os.WriteFile(filePath, []byte(strings.Join(final, "\n")+"\n"), 0644)
}

func ExtractIPKeysFromTemplate(path string) ([]string, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	regex := regexp.MustCompile(`{{\s*index\s+\.IPsByService\s+"([^"]+)"\s*}}`)
	matches := regex.FindAllStringSubmatch(string(content), -1)

	set := make(map[string]bool)
	for _, m := range matches {
		set[m[1]] = true
	}

	keys := make([]string, 0, len(set))

	for k := range set {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys, nil
}
