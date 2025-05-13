package internal

// File paths
const (
	IPMapPath           = ".ipmap.env"
	TemplateDir         = "templates"
	ProjectDirPrefix    = "domains"
	WindowsHostsPath    = "/mnt/c/Windows/System32/drivers/etc/hosts"
	SharedServicesDir   = "shared-services"
)

// Docker container names
const (
	ReverseProxyName = "nginx-reverse-proxy"
	SharedMySQLName  = "shared_mysql"
)

// Directory structure
const (
	CertsDir          = "shared-services/certs"
	SitesDir          = "sites"
	NginxConfFileName = "nginx.conf"
	DockerComposeFile = "docker-compose.yml"
)

// Project structure folders
var ProjectFolders = []string{"image", "conf", "logs", "data"}

// Environment variable names
const (
	EnvNetworkName       = "NETWORK_NAME"
	EnvProjectStartIP    = "PROJECT_START_IP"
	EnvSharedMySQLIP     = "SHARED_MYSQL_IP"
	EnvReverseProxyIP    = "REVERSE_PROXY_IP"
	EnvMySQLRootPassword = "MYSQL_ROOT_PASSWORD"
	EnvMySQLUser         = "MYSQL_USER"
	EnvMySQLPassword     = "MYSQL_PASSWORD"
)

// Reserved IP suffixes
const (
	ReservedIPNetwork    = 0  // Network identifier
	ReservedIPGateway    = 1  // Default gateway
	ReservedIPBroadcast1 = 254 // Broadcast address (for some networks)
	ReservedIPBroadcast2 = 255 // Broadcast address
)

// Feature flags
const (
	// SSLEnabled controls whether SSL is enabled for projects
	// Currently, this must be true as SSL is required for the application to work
	SSLEnabled = true
) 