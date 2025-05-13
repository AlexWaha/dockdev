package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	certsDir := filepath.Join("shared-services", "certs")
	if err := ensureRootCA(certsDir); err != nil {
		return fmt.Errorf("SSL rootCA failed: %w", err)
	}

	crtPath, keyPath, err := generateDomainCert(domain, certsDir)
	if err != nil {
		return fmt.Errorf("Domain SSL generation failed: %w", err)
	}

	// conf/nginx/ssl
	sslDstDir := filepath.Join(projectDir, "conf", "nginx", "ssl")
	os.MkdirAll(sslDstDir, 0755)
	copy(crtPath, filepath.Join(sslDstDir, "cert.crt"))
	copy(keyPath, filepath.Join(sslDstDir, "cert.key"))

	// docker-compose.yml
	if err := RenderTemplate(
		filepath.Join(templateDir, "docker-compose.yml.tmpl"),
		filepath.Join(projectDir, "docker-compose.yml"),
		data,
	); err != nil {
		return err
	}

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
