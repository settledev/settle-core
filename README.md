# Settle Core

[![Go Version](https://img.shields.io/badge/Go-1.23.0+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/settlectl/settle-core)](https://goreportcard.com/report/github.com/settlectl/settle-core)

**Settle** is a modern, minimal, state-aware configuration automation tool built for infrastructure and platform engineers who want the power of Ansible and the safety of Terraform — without the overhead.

## 🚀 Features

- ** Agentless**: No daemons or remote installations. Uses native SSH to connect to your servers
- **📄 Declarative**: Define your desired configuration using a simple, readable DSL
- ** Safe-by-default**: Built-in lifecycle awareness with local state file
- **⛏️ Composable**: Resources like package, service, and file are first-class primitives
- ** Fast & Lightweight**: Designed to be small, understandable, and fast to adopt

## ️ Core Philosophy

Settle follows a hierarchical graph-based approach with clear dependency flows:

### Graph Layers (bottom to top)

- **Foundation Layer**: Hardware, OS, basic networking
- **Platform Layer**: Package managers, base services (systemd, networking)
- **Infrastructure Layer**: Databases, message queues, storage
- **Application Layer**: Your actual services and applications
- **Configuration Layer**: Service configs, users, permissions
- **Runtime Layer**: Active processes, connections, health checks

### Edge Types

- **Depends on**: Hard dependency - must exist first
- **Configures**: Soft dependency - can modify existing
- **Monitors**: Observational - reads state
- **Triggers**: Event-based - causes actions

### Core Rules

- **No circular dependencies**: If A depends on B, B cannot depend on A
- **Explicit before implicit**: All dependencies must be declared
- **Idempotent operations**: Running twice produces same result
- **Atomic changes**: Operations succeed completely or fail completely

## 📦 Installation

### Prerequisites

- Go 1.23.0 or later
- SSH access to target servers

### Quick Install

```bash
# Clone the repository
git clone https://github.com/settlectl/settle-core.git
cd settle-core

# Build the binary
go build -o settlectl

# Make it available in your PATH
sudo mv settlectl /usr/local/bin/
```

### From Source

```bash
go install github.com/settlectl/settle-core@latest
```

## Quick Start

1. **Define your hosts** in a `.stl` file:

```stl
host "web-server" {
  hostname = "192.168.1.100"
  user     = "ubuntu"
  port     = 22
  keyfile  = "/path/to/your/key.pem"
  group    = "web"
}
```

2. **Define your packages**:

```stl
package "nginx" {
    version = "latest"
    manager = "apt"
}
```

3. **Test connectivity**:

```bash
settlectl ping
```

4. **Plan your changes**:

```bash
settlectl plan
```

5. **Apply your configuration**:

```bash
settlectl create
```

## 📖 Usage

### Basic Commands

```bash
# Check SSH connectivity to all hosts
settlectl ping

# See what would change without applying
settlectl plan

# Apply changes from your config
settlectl create

# Safely remove config and reverse state
settlectl clean


```

### Configuration Files

Settle uses `.stl` files for configuration. These are declarative and describe your desired state:

```stl
# Define hosts
host "app-server" {
  hostname = "10.0.0.1"
  user     = "admin"
  port     = 22
  keyfile  = "/path/to/key"
  group    = "application"
}

# Define packages
package "docker" {
    version = "latest"
    manager = "apt"
}

# Define services
service "nginx" {
    state = "running"
    enabled = true
}
```

## ️ Project Structure
settle-core/
├── cmd/ # CLI commands (ping, plan, apply, etc.)
├── core/ # Core engine and graph logic
├── common/ # Shared types and constants
├── drivers/ # Package managers and service drivers
├── inventory/ # Host management and SSH connectivity
├── examples/ # Example configuration files
└── resources/ # Resource definitions