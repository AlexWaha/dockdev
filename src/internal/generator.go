package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/joho/godotenv"
)

type TemplateData struct {
	Domain       string
	Prefix       string
	NetworkName  string
	IPsByService map[string]string
	UseSSL       bool
}

type SharedTemplateData struct {
	NetworkName        string
	ReverseProxyIP     string
	SharedMySQLIP      string
	MySQLRootPassword  string
	MySQLUser          string
	MySQLPassword      string
}

// GenerateProject creates a new project with the given domain name
// The useSSL parameter controls whether SSL is enabled for the project
// Currently, SSL is required for the application to work correctly
func GenerateProject(domain string, useSSL ...bool) error {
	// Ensure Docker is running before proceeding
	if err := EnsureDockerRunning(); err != nil {
		return fmt.Errorf("Docker check failed: %w", err)
	}

	enableSSL := SSLEnabled

	// If useSSL parameter is provided, use it
	if len(useSSL) > 0 {
		enableSSL = useSSL[0]
	}

	if err := godotenv.Load(".env"); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	network := os.Getenv(EnvNetworkName)
	baseIP := os.Getenv(EnvProjectStartIP)
	mysqlIP := os.Getenv(EnvSharedMySQLIP)
	projectDir := filepath.Join(ProjectDirPrefix, domain)
	prefix := strings.Split(domain, ".")[0]

	if mysqlIP != "" {
		_ = InsertIPMappingAtTop(IPMapPath, "shared-mysql", mysqlIP)
	}

	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		return fmt.Errorf("Project already exists: %s", projectDir)
	}
	if err := CreateDirIfNotExist(projectDir); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	usedIPs, err := LoadUsedIPs(IPMapPath)
	if err != nil {
		return err
	}

	ipKeys, err := ExtractIPKeysFromTemplate(filepath.Join(TemplateDir, DockerComposeFile+".tmpl"))
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
		if err := AppendIPMapping(IPMapPath, entry, ip); err != nil {
			return err
		}
	}

	data := TemplateData{
		Domain: domain,
		Prefix: prefix,
		NetworkName: network,
		IPsByService: ipMap,
		UseSSL: enableSSL,
	}

	if enableSSL {
		if err := ensureRootCA(CertsDir); err != nil {
			return fmt.Errorf("SSL rootCA failed: %w", err)
		}

		crtPath, keyPath, err := generateDomainCert(domain, CertsDir)
		if err != nil {
			return fmt.Errorf("Domain SSL generation failed: %w", err)
		}

		// Copy certificates to project's nginx ssl directory
		sslDstDir := filepath.Join(projectDir, "conf", "nginx", "ssl")
		if err := CopyCertificates(crtPath, keyPath, sslDstDir); err != nil {
			return err
		}
	}

	// Render docker-compose.yml from template
	if err := RenderTemplate(
		filepath.Join(TemplateDir, DockerComposeFile+".tmpl"),
		filepath.Join(projectDir, DockerComposeFile),
		data,
	); err != nil {
		return err
	}

	// Copy all required directories from template to project
	if err := CopyTemplatedDirectories(TemplateDir, projectDir, ProjectFolders); err != nil {
		return err
	}

	// Create and render nginx config
	confDir := filepath.Join(projectDir, "conf", "nginx")
	if err := CreateDirIfNotExist(confDir); err != nil {
		return fmt.Errorf("failed to create nginx config directory: %w", err)
	}

	if err := RenderTemplate(
		filepath.Join(TemplateDir, "nginx.conf.tmpl"),
		filepath.Join(confDir, "default.conf"),
		data,
	); err != nil {
		return err
	}

	// Create and render app/index.html
	appDstDir := filepath.Join(projectDir, "app")
	if err := CreateDirIfNotExist(appDstDir); err != nil {
		return fmt.Errorf("failed to create app directory: %w", err)
	}

	if err := RenderTemplate(
		filepath.Join(TemplateDir, "app", "index.html"),
		filepath.Join(appDstDir, "index.html"),
		data,
	); err != nil {
		return err
	}

	// Create and configure shared services
	sitesDir := filepath.Join(SharedServicesDir, SitesDir)
	if err := CreateDirIfNotExist(sitesDir); err != nil {
		return fmt.Errorf("failed to create sites directory: %w", err)
	}

	// Generate shared-services/docker-compose.yml if it doesn't exist
	sharedComposeTemplate := filepath.Join(TemplateDir, SharedServicesDir, DockerComposeFile+".tmpl")
	sharedComposePath := filepath.Join(SharedServicesDir, DockerComposeFile)
	if _, err := os.Stat(sharedComposePath); os.IsNotExist(err) {
		sharedTemplate := SharedTemplateData{
			NetworkName:        network,
			ReverseProxyIP:     os.Getenv(EnvReverseProxyIP),
			SharedMySQLIP:      os.Getenv(EnvSharedMySQLIP),
			MySQLRootPassword:  os.Getenv(EnvMySQLRootPassword),
			MySQLUser:          os.Getenv(EnvMySQLUser),
			MySQLPassword:      os.Getenv(EnvMySQLPassword),
		}

		if err := RenderTemplate(sharedComposeTemplate, sharedComposePath, sharedTemplate); err != nil {
			return fmt.Errorf("Failed to render %s: %w", filepath.Join(SharedServicesDir, DockerComposeFile), err)
		}

		fmt.Println("Generated " + filepath.Join(SharedServicesDir, DockerComposeFile))
	}

	// Render shared-services/nginx.conf
	nginxConfTemplate := filepath.Join(TemplateDir, SharedServicesDir, NginxConfFileName+".tmpl")
	nginxConfDest := filepath.Join(SharedServicesDir, NginxConfFileName)
	if err := RenderTemplate(nginxConfTemplate, nginxConfDest, data); err != nil {
		return fmt.Errorf("Failed to render nginx.conf: %w", err)
	}
	fmt.Println("Generated reverse proxy nginx.conf")

	// Copy shared-services/image if it exists
	sharedImageSrc := filepath.Join(TemplateDir, SharedServicesDir, "image")
	sharedImageDst := filepath.Join(SharedServicesDir, "image")
	if _, err := os.Stat(sharedImageSrc); err == nil {
		if err := CopyDir(sharedImageSrc, sharedImageDst); err != nil {
			return fmt.Errorf("Failed to copy shared-services image: %w", err)
		}
	}

    // Create site configuration
    siteConf := filepath.Join(sitesDir, domain+".conf")

    if _, err := os.Stat(siteConf); os.IsNotExist(err) {
        tmpl := "site.conf.tmpl"
        if enableSSL {
            tmpl = "site-ssl.conf.tmpl"
        }

        if err := RenderTemplate(filepath.Join(TemplateDir, tmpl), siteConf, data); err != nil {
            return err
        }

        fmt.Printf("Created reverse proxy %s config: %s\n",
            map[bool]string{true: "SSL", false: "no-SSL"}[enableSSL], siteConf)
    }

	fmt.Println("Starting shared-services...")
	if err := runDockerComposeUp(SharedServicesDir); err != nil {
		return err
	}

	root := os.Getenv(EnvMySQLRootPassword)
	user := os.Getenv(EnvMySQLUser)

	fmt.Println("Waiting for MySQL to become ready...")
	if err := waitForMySQL(SharedMySQLName, root); err != nil {
		return err
	}

	fmt.Println("Granting privileges to user...")
	if err := grantAllPrivileges(SharedMySQLName, root, user); err != nil {
		return fmt.Errorf("Failed to grant privileges: %w", err)
	}

	fmt.Println("Starting project containers...")
	if err := runDockerComposeUp(projectDir); err != nil {
		return err
	}

	// Reload or restart the Nginx reverse proxy
	if err := restartNginxReverseProxy(); err != nil {
		return err
	}

	if err := updateWindowsHosts(domain, WindowsHostsPath); err != nil {
		return err
	}

	// Display project information
	PrintSectionDivider("PROJECT CREATED SUCCESSFULLY")
	fmt.Println(Success("Your new development environment is ready!"))

	projectURL := GetProjectURL(domain)
	fmt.Println(Info("\nYou can access your project at:"), Bold(Highlight(projectURL)))

	// Ask to open in browser if in terminal mode
	if IsTerminal() {
		if YesNoPrompt("Would you like to open the project in your browser now?", true) {
			if err := OpenBrowser(projectURL); err != nil {
				fmt.Println(Warning("Could not open browser automatically."))
				fmt.Println(Info("Please open this URL manually:"), Highlight(projectURL))
			}
		} else {
			fmt.Println(Info("You can open this URL in your browser when ready:"), Highlight(projectURL))
		}
	}

	return nil
}
