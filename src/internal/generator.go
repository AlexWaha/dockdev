package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"github.com/joho/godotenv"
)

type TemplateData struct {
	Domain       string
	Prefix       string
	NetworkName  string
	IPsByService map[string]string
}

type SharedTemplateData struct {
	NetworkName        string
	ReverseProxyIP     string
	SharedMySQLIP      string
	MySQLRootPassword  string
	MySQLUser          string
	MySQLPassword      string
}

func GenerateProject(domain string) error {
	if err := godotenv.Load(".env"); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	network := os.Getenv("NETWORK_NAME")
	baseIP := os.Getenv("PROJECT_START_IP")
	mysqlIP := os.Getenv("SHARED_MYSQL_IP")
	ipmapPath := ".ipmap.env"
	templateDir := "templates"
	projectDir := filepath.Join("domains", domain)
	prefix := strings.Split(domain, ".")[0]
	hostsPath := "/mnt/c/Windows/System32/drivers/etc/hosts"
	sharedServicesDir := "shared-services"
	reverseProxyName := "nginx-reverse-proxy"

	// Insert shared-mysql IP at the top
	if mysqlIP != "" {
		_ = InsertIPMappingAtTop(ipmapPath, "shared-mysql", mysqlIP)
	}

	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		return fmt.Errorf("Project already exists: %s", projectDir)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return err
	}

	usedIPs, err := LoadUsedIPs(ipmapPath)
	if err != nil {
		return err
	}

	ipKeys, err := ExtractIPKeysFromTemplate(filepath.Join(templateDir, "docker-compose.yml.tmpl"))
	if err != nil {
		return err
	}

	ipMap := map[string]string{}

	for _, key := range ipKeys {
		ip, err := FindNextFreeIP(baseIP, usedIPs)
		if err != nil {
			return err
		}
		usedIPs[ip] = true
		ipMap[key] = ip

		entry := domain
		if key != "main" {
			entry += "_" + key
		}
		if err := AppendIPMapping(ipmapPath, entry, ip); err != nil {
			return err
		}
	}

	data := TemplateData{
		Domain: domain, Prefix: prefix, NetworkName: network, IPsByService: ipMap,
	}

	// docker-compose.yml
	if err := RenderTemplate(
		filepath.Join(templateDir, "docker-compose.yml.tmpl"),
		filepath.Join(projectDir, "docker-compose.yml"),
		data,
	); err != nil {
		return err
	}

	// image/, conf/, logs/, data/
	for _, dir := range []string{"image", "conf", "logs", "data"} {
		src := filepath.Join(templateDir, dir)
		dst := filepath.Join(projectDir, dir)
		if _, err := os.Stat(src); err == nil {
			if err := CopyDir(src, dst); err != nil {
				return fmt.Errorf("Failed to copy %s: %w", dir, err)
			}
		}
	}

	// conf/nginx/default.conf
	confDir := filepath.Join(projectDir, "conf", "nginx")
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return err
	}

	if err := RenderTemplate(
		filepath.Join(templateDir, "nginx.conf.tmpl"),
		filepath.Join(confDir, "default.conf"),
		data,
	); err != nil {
		return err
	}

	// app/index.html
	appDstDir := filepath.Join(projectDir, "app")
	if err := os.MkdirAll(appDstDir, 0755); err != nil {
		return err
	}

	if err := RenderTemplate(
		filepath.Join(templateDir, "app", "index.html"),
		filepath.Join(appDstDir, "index.html"),
		data,
	); err != nil {
		return err
	}

	// shared-services/sites/domain.conf
	sitesDir := filepath.Join(sharedServicesDir, "sites")
	if err := os.MkdirAll(sitesDir, 0755); err != nil {
		return err
	}

	// Generate shared-services/docker-compose.yml if it doesn't exist
	sharedComposeTemplate := filepath.Join(templateDir, sharedServicesDir, "docker-compose.yml.tmpl")
	sharedComposePath := filepath.Join(sharedServicesDir, "docker-compose.yml")
	if _, err := os.Stat(sharedComposePath); os.IsNotExist(err) {
		sharedTemplate := SharedTemplateData{
			NetworkName:        network,
			ReverseProxyIP:     os.Getenv("REVERSE_PROXY_IP"),
			SharedMySQLIP:      os.Getenv("SHARED_MYSQL_IP"),
			MySQLRootPassword:  os.Getenv("MYSQL_ROOT_PASSWORD"),
			MySQLUser:          os.Getenv("MYSQL_USER"),
			MySQLPassword:      os.Getenv("MYSQL_PASSWORD"),
		}

		if err := RenderTemplate(sharedComposeTemplate, sharedComposePath, sharedTemplate); err != nil {
			return fmt.Errorf("Failed to render shared-services/docker-compose.yml: %w", err)
		}
		
		fmt.Println("Generated shared-services/docker-compose.yml")
	}

	// Render shared-services/nginx.conf
	nginxConfTemplate := filepath.Join(templateDir, sharedServicesDir, "nginx.conf.tmpl")
	nginxConfDest := filepath.Join(sharedServicesDir, "nginx.conf")
	if err := RenderTemplate(nginxConfTemplate, nginxConfDest, data); err != nil {
		return fmt.Errorf("Failed to render nginx.conf: %w", err)
	}
	fmt.Println("Generated reverse proxy nginx.conf")

	// Copy shared-services/image if it exists
	sharedImageSrc := filepath.Join(templateDir, sharedServicesDir, "image")
	sharedImageDst := filepath.Join(sharedServicesDir, "image")
	if _, err := os.Stat(sharedImageSrc); err == nil {
		if err := CopyDir(sharedImageSrc, sharedImageDst); err != nil {
			return fmt.Errorf("Failed to copy shared-services image: %w", err)
		}
	}

	siteConf := filepath.Join(sitesDir, domain+".conf")
	if _, err := os.Stat(siteConf); os.IsNotExist(err) {
		if err := RenderTemplate(
			filepath.Join(templateDir, "site.conf.tmpl"),
			siteConf,
			data,
		); err != nil {
			return err
		}
		
		fmt.Println("Created reverse proxy config:", siteConf)
	}

	fmt.Println("Starting shared-services...")
	if err := runDockerComposeUp(sharedServicesDir); err != nil {
		return err
	}

	root := os.Getenv("MYSQL_ROOT_PASSWORD")
	user := os.Getenv("MYSQL_USER")

	fmt.Println("Waiting for MySQL to become ready...")
	if err := waitForMySQL("shared_mysql", root); err != nil {
		return err
	}

	fmt.Println("Granting privileges to user...")
	if err := grantAllPrivileges("shared_mysql", root, user); err != nil {
		return fmt.Errorf("Failed to grant privileges: %w", err)
	}

	fmt.Println("Starting project containers...")
	if err := runDockerComposeUp(projectDir); err != nil {
		return err
	}

	fmt.Println("Reloading reverse proxy config...")
	err = exec.Command("docker", "exec", reverseProxyName, "nginx", "-s", "reload").Run()
	if err != nil {
		fmt.Println("Reload failed, restarting container...")
		err = exec.Command("docker", "restart", reverseProxyName).Run()
		if err != nil {
			return fmt.Errorf("Failed to reload or restart reverse proxy: %w", err)
		}
	}

	if err := updateWindowsHosts(domain, hostsPath); err != nil {
		return err
	}

	return nil
}

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

func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}

		defer srcFile.Close()
		dstFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}

		defer dstFile.Close()
		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

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

func runDockerComposeUp(dir string) error {
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitForMySQL(container, rootPass string) error {
	for i := 1; i <= 30; i++ {
		cmd := exec.Command("docker", "exec", container,
			"mysql", "-uroot", fmt.Sprintf("-p%s", rootPass),
			"-e", "SELECT 1;")

		if err := cmd.Run(); err == nil {
			fmt.Println("MySQL is ready.")
			return nil
		}

		fmt.Printf("Waiting for MySQL... (%d/30)\n", i)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("MySQL is not responding in container: %s", container)
}

func grantAllPrivileges(container, rootPass, user string) error {
	sql := fmt.Sprintf(`GRANT ALL PRIVILEGES ON *.* TO '%s'@'%%' WITH GRANT OPTION;`, user)
	cmd := exec.Command("docker", "exec", "-i", container, "mysql", "-uroot", fmt.Sprintf("-p%s", rootPass), "-e", sql)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

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