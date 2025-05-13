package internal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func LoadUsedIPs(ipmapPath string) (map[string]bool, error) {
	used := make(map[string]bool)

	file, err := os.Open(ipmapPath)
	if err != nil {
		if os.IsNotExist(err) {
			return used, nil // empty map if file doesn't exist
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			used[parts[1]] = true
		}
	}

	return used, scanner.Err()
}

// isReservedIPSuffix checks if the IP last octet is reserved
func isReservedIPSuffix(octet int) bool {
	return octet == ReservedIPNetwork || 
	       octet == ReservedIPGateway || 
	       octet == ReservedIPBroadcast1 || 
	       octet == ReservedIPBroadcast2
}

func FindNextFreeIP(base string, used map[string]bool) (string, error) {
	prefix := base[:strings.LastIndex(base, ".")]
	start, _ := strconv.Atoi(base[strings.LastIndex(base, ".")+1:])

	// Ensure start is at least 2 to avoid .0 and .1
	if start < 2 {
		start = 2
	}

	for i := start; i < 254; i++ {
		// Skip reserved IP suffixes
		if isReservedIPSuffix(i) {
			continue
		}
		
		candidate := fmt.Sprintf("%s.%d", prefix, i)
		if !used[candidate] {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no available IPs in subnet")
}

func AppendIPMapping(ipmapPath, domain, ip string) error {
	f, err := os.OpenFile(ipmapPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%s=%s\n", domain, ip))
	return err
}
