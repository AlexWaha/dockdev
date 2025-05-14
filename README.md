# 🚀 DockDev — Instant Docker Dev Domains

> ⚡️ DockDev is a fast CLI tool that helps you create isolated Docker-based development environments with reverse proxy and custom local domains.
>
> It lets you work locally in Windows + WSL with multiple projects — and access them in the browser using friendly domain names like https://app.test.

![Docker](https://img.shields.io/badge/Docker-ready-blue)
![Go](https://img.shields.io/badge/Built%20with-Go-informational)
![WSL2](https://img.shields.io/badge/WSL2-supported-green)
![License](https://img.shields.io/badge/license-MIT-lightgrey)

---

## 📦 What is DockDev?

**DockDev** is a developer CLI utility written in Go that helps you instantly:

✅ Features
- 🔧 Spin up isolated Docker-based dev environments with NGINX, PHP, Node.js, Redis, ElasticMQ etc.
- 🌐 All traffic is routed through a shared reverse proxy (nginx-reverse-proxy) for seamless local domain support
- 🔐 **Automatic SSL certificate generation** with Windows trust store integration (HTTPS ready!)
- 🛠 Assign static IPs via a shared user-defined Docker network (bridge mode)
- 🌍 Access each project via clean local domains like https://app.test
- 🗂 Automatically add the domain to your Windows hosts file
- ⚙️ Reverse proxy configs are generated per-project and hot-reloaded or restarted as needed
- 🗃️ Includes a shared MySQL container for all projects — connect via native MySQL GUI clients on Windows
- 🖥️ Interactive CLI interface for creating and managing projects
- 🚀 Automatic Docker Desktop startup and container management

✅ Ideal for full-stack development inside **WSL2 + Docker Desktop** environments.

---

## 📋 Prerequisites

- Windows with WSL2
- Docker Desktop
- Go 1.24+ (for building from source)

## 🛠 Installation & Build

### Option 1: Build from Source 🔨

1. Clone the repository
2. Run the build script:

```bash
./build.bat
```

The build script will:
- Compile the Go code for Linux (WSL compatibility)
- Create a distribution in the `../dist` folder with the executable

3. Manually copy the required files to your WSL user folder for example to `/home/user/dockdev`:
   - `dist/dockdev` (executable)
   - `dist/templates/` (directory with all templates)
   - Create a `.env` configuration file (see Configuration section)

4. Set proper executable permissions in your WSL environment:
```bash
chmod +x dockdev
```

### Option 2: Use Pre-built Binary 📦

1. Download the latest release from the releases page
2. Extract the archive to your preferred location
3. Run the executable from WSL2

---

## ⚙️ Setup

### Required files and folders in your WSL environment:

- `.env` (configuration file)
- `dockdev` (executable)
- `templates/` (directory with templates)

---
## 💡 Before you start!

> If your index application folder is `public` or another, update `nginx.conf`
>
> For example: `root /var/www/html/public;`
>
> 1. You can update it before adding new project in `templates/nginx.conf.tmpl` for all projects
>
> 2. or after, directly in `domains/YOUR_DOMAIN/conf/nginx/default.conf`
> #### If #2 - Don't forget to remove and run project containers manually!

---

## 🚀 Usage

### 📘 Available Commands

| Command | Description |
|---------|-------------|
| `./dockdev` | Start interactive mode |
| `./dockdev domain.test` | Create a new project with the specified domain |
| `./dockdev domain.test --no-ssl` | Create a project without SSL (not recommended) |
| `./dockdev rm domain.test` | Delete an existing project |
| `./dockdev -H` or `--help` | Show help message |

### 💬 Interactive Mode

Run the tool without arguments to enter interactive mode:

```bash
./dockdev
```

Interactive mode will:
- Show existing projects
- Allow you to create new projects
- Allow you to delete existing projects
- Guide you through all options with prompts

### 🆕 Create a New Project

```bash
./dockdev mydomain.test
```

🔧 It will:

- Create `domains/mydomain.test/`
- Generate SSL certificates and add them to Windows trust store (**HTTPS ready**)
- Assign next free IP like `172.20.0.12`
- Assign IP's for all project containers
- Generate:
    - `docker-compose.yml`
    - `conf/nginx/default.conf`
    - `app/index.html`
    - reverse proxy config in `shared-services/sites`
- Update:
    - `.ipmap.env` > This file just FYI
    - `Windows hosts` file
- Automatically start all services
- Open your browser to https://mydomain.test

> Your application must be in  `app` folder: `domains/YOUR_DOMAIN/app`

### 🗑️ Delete a Project

```bash
./dockdev rm mydomain.test
```

You'll be prompted:

```
Are you sure you want to delete domain 'mydomain.test'? [y/N]
```

Deletes:

- Domain folder
- Reverse proxy `.conf`
- IP mapping entry
- Hosts file entry
- SSL certificates from disk and Windows trust store
- Drop all domain containers

### ❓ Show Help

```bash
./dockdev --help
```

---

## 🏗️ Project Structure

Each project environment includes:

- 🌐 Nginx web server
- 🐘 PHP-FPM 8.3
- 📦 Node.js 23
- ⚡ Redis server
- 📨 ElasticMQ (SQS-compatible message queue)
- 🗃️ Shared MySQL database (across all projects)

---

## 🔐 SSL Certificates

DockDev automatically:

1. Generates a root CA certificate and adds it to Windows trust store
2. Creates domain-specific SSL certificates for each project
3. Configures Nginx to use HTTPS
4. Makes all projects accessible via secure HTTPS connections

This means you can develop with:
- HTTPS by default
- No browser security warnings
- Proper SSL testing in your local environment
- Full compatibility with secure-only features

---

## ⚙️ Configuration

The tool uses an `.env` file for configuration:

```
NETWORK_NAME=local_net
SUBNET=10.0.100.0/24
REVERSE_PROXY_IP=10.0.100.2
SHARED_MYSQL_IP=10.0.100.3
PROJECT_START_IP=10.0.100.10
MYSQL_ROOT_PASSWORD=root
MYSQL_USER=user
MYSQL_PASSWORD=userpass
```

---

## 📁 Directory Structure

```
├── dockdev                  # Main executable
├── templates/               # Project templates
│   ├── app/                 # Default web application files
│   ├── conf/                # Configuration templates
│   ├── image/               # Docker image definitions
│   │   ├── node/            # Node.js image
│   │   └── php/             # PHP image
│   └── shared-services/     # Shared services templates
├── domains/                 # Generated project directories
└── shared-services/         # Shared services (MySQL, Nginx proxy)
    ├── certs/               # SSL certificates
    ├── sites/               # Nginx site configurations
    └── data/                # Persistent data
```

---

## 🧱 Architecture

- 🔧 `dockdev`: CLI manager (Go)
- 📁 `templates/`: reusable template files
> You can extend docker-compose.yml.tmpl with your containers
- 🌍 `shared-services/`: reverse proxy & shared MySQL DB
- 🛠 `.ipmap.env`
>📘 Reference file for developers, to track which containers was assigned to which IP
- 🔌 All containers in one shared Docker `bridge` network

---

## ✅ Platform Compatibility

| Platform              | Supported      |
|-----------------------|----------------|
| ✅ Windows 10/11 + WSL | ✔️ Recommended |

---

All routed via NGINX with shared IP space and automatic DNS mapping.

---

## 📝 License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details. 
