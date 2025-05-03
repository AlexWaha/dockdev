# 🚀 DockDev — Instant Docker Dev Domains

> ⚡️ DockDev is a fast CLI tool that helps you create isolated Docker-based development environments with reverse proxy and custom local domains. 
> 
> It lets you work locally in Windows + WSL with multiple projects — and access them in the browser using friendly domain names like http://app.local.

![Docker](https://img.shields.io/badge/Docker-ready-blue)
![Go](https://img.shields.io/badge/Built%20with-Go-informational)
![WSL2](https://img.shields.io/badge/WSL2-supported-green)
![License](https://img.shields.io/badge/license-MIT-lightgrey)

---

## 📦 What is DockDev?

**DockDev** is a developer CLI utility written in Go that helps you instantly:

✅ Features
- 🔧 Spin up isolated Docker-based dev environments with NGINX and any containers like PHP, Redis etc.
- 🌐 All traffic is routed through a shared reverse proxy (nginx-reverse-proxy) for seamless local domain support
- 🛠 Assign static IPs via a shared user-defined Docker network (bridge mode)
- 🌍 Access each project via clean local domains like http://app.local
- 🗂 Automatically add the domain to your Windows hosts file
- ⚙️ Reverse proxy configs are generated per-project and hot-reloaded or restarted as needed
- 🗃️ Includes a shared MySQL container for all projects — connect via native MySQL GUI clients on Windows (e.g. TablePlus, DBeaver, DataGrip etc.)

✅ Ideal for full-stack development inside **WSL2 + Docker Desktop** environments.

---

## 🛠 Installation & Build

💡 You don't need to build it, use the ready-to-run script `dockerdev`

1. Install [Go](https://go.dev/dl/)
2. Clone the repository and enter the folder:

```bash
git clone https://github.com/your-org/dockdev.git
cd dockdev
```

3. Build the binary (choose based on your OS)::

🪟 For Windows (CMD):
Run `build.bat`

> This produces a Linux-compatible binary you can run inside WSL2 or Linux servers.

---

## ⚙️ Setup

### Copy required files and folders:

> All from folder dist to your WSL

- `.env`
- `dockdev`
- `templates`

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


## 🚀 Usage

### ➕ Create a new domain/project

#### Interactive:

```bash
./dockdev
```

#### Direct:

```bash
./dockdev app.local
```

🔧 It will:

- Create `domains/app.local/`
- Assign next free IP like `10.0.100.12`
- Assign IP's for all project containers
- Generate:
  - `docker-compose.yml`
  - `conf/nginx/default.conf`
  - `app/index.html`
  - reverse proxy config in `shared-services/sites`
- Update:
  - `.ipmap.env` > This file just FYI
  - `Windows hosts` file

> Your application must be in  `app` folder: `domains/YOUR_DOMAIN/app`
---

### 🗑 Remove a project

```bash
./dockdev rm app.local
```

You'll be prompted:

```
Are you sure you want to delete domain 'app.local'? [y/N]
```

Deletes:

- Domain folder
- Reverse proxy `.conf`
- IP mapping entry
- Hosts file entry
- Drop all domain containers

---

## 🧱 Architecture

- 🔧 `dockdev`: CLI manager (Go)
- 📁 `templates/`: reusable template files 
> You can extend docker-compose.yml.tmpl with your containers
- 🌍 `shared-services/`: reverse proxy & shared Mysql DB
- 🛠 `.ipmap.env`
>📘 Reference file for developers, to track which containers was assigned to which IP
- 🔌 All containers in one shared Docker `bridge` network

---

## ✅ Platform Compatibility

| Platform              | Supported |
|-----------------------|-----------|
| ✅ Windows 10/11 + WSL | ✔️ Recommended |

---

All routed via NGINX with shared IP space and automatic DNS mapping.

---

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
