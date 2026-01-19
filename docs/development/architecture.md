# Dokku Architecture

This document provides a developer-focused overview of Dokku's internal architecture. It is intended for contributors who want to understand how Dokku works under the hood before diving into the codebase.

## High-Level Architecture

Dokku is a Docker-powered Platform as a Service (PaaS) that provides a Heroku-like deployment experience. The following diagram shows the main components and their interactions:

```
+------------------+
|     User/Client  |
+--------+---------+
         |
         | SSH / git push
         v
+--------+---------+
|   dokku binary   |
|  (bash script)   |
+--------+---------+
         |
         | parse_args / execute_dokku_cmd
         v
+--------+---------+
|   Plugin System  |
|      (plugn)     |
+--------+---------+
         |
         | triggers / subcommands
         v
+--------+---------+     +------------------+
|  Core Plugins    |---->|  Docker / K8s    |
|  (apps, git,     |     |    Runtime       |
|   config, etc.)  |     +------------------+
+------------------+
```

**Key components:**

- **dokku binary**: The main entry point (`/dokku`), a bash script that handles authentication, argument parsing, and routes commands to plugins
- **Plugin System**: Uses [plugn](https://github.com/dokku/plugn) to execute triggers and discover commands
- **Core Plugins**: Implement all Dokku functionality (apps, git, config, builders, schedulers, proxies, etc.)
- **Runtime**: Docker (default via `scheduler-docker-local`) or Kubernetes (via `scheduler-k3s`)

## Directory Structure

### Source Tree

```
dokku/
├── dokku                 # Main CLI entry point (bash script)
├── plugins/              # All plugin source code
│   ├── apps/            # App management
│   ├── git/             # Git push handling
│   ├── config/          # Environment variables
│   ├── builder-*/       # Build backends (herokuish, pack, dockerfile, etc.)
│   ├── scheduler-*/     # Deployment schedulers (docker-local, k3s)
│   ├── *-vhosts/        # Proxy implementations (nginx, traefik, caddy, etc.)
│   └── common/          # Shared functions and utilities
├── docs/                 # Documentation (markdown)
├── debian/               # Debian packaging files
├── contrib/              # Installation scripts and helpers
└── tests/                # Integration tests (bats)
```

### Runtime Directories

When Dokku is installed, it creates the following directory structure:

```
/var/lib/dokku/
├── core-plugins/         # Core plugin binaries (installed from source)
│   ├── available/       # All available core plugins
│   └── enabled/         # Symlinks to enabled plugins
├── plugins/              # Community plugins
│   ├── available/
│   └── enabled/
└── data/                 # Plugin data storage
    └── <plugin>/        # Per-plugin data
        └── <app>/       # Per-app properties

$DOKKU_ROOT (~dokku by default)
├── <app>/               # Per-app data
│   ├── refs/           # Git refs
│   ├── HEAD            # Current git HEAD
│   ├── tls/            # SSL certificates
│   └── ...             # Other app-specific files
└── .dokkurc/            # Global configuration overrides
```

### Plugin Directory Layout

Each plugin follows a consistent structure:

```
plugins/<plugin-name>/
├── plugin.toml           # Plugin metadata (description, version)
├── commands              # Help output and catch-all command handler
├── subcommands/          # Individual command implementations
│   ├── default          # Default command (e.g., `dokku apps`)
│   └── <command>        # Named commands (e.g., `dokku apps:create`)
├── functions             # Public functions for other plugins to source
├── internal-functions    # Private functions
├── triggers.go           # Go-based trigger implementations
├── *.go                  # Additional Go code
└── Makefile              # Build configuration
```

## Plugin System Architecture

Dokku's functionality is entirely implemented through plugins. This architecture provides:

- **Extensibility**: Add new features without modifying core code
- **Loose coupling**: Plugins communicate via well-defined triggers
- **Composability**: Mix and match builders, schedulers, and proxies

### Plugin Communication via Triggers

Plugins communicate through the **trigger system** powered by [plugn](https://github.com/dokku/plugn). When a trigger is fired, plugn executes matching scripts from all enabled plugins:

```
+---------------+     plugn trigger     +---------------+
|    Plugin A   | ------------------->  |    Plugin B   |
| (fires event) |    "post-deploy"      | (listens for  |
+---------------+                       |    event)     |
                                        +---------------+
                                              |
                                              v
                                        +---------------+
                                        |    Plugin C   |
                                        | (also listens)|
                                        +---------------+
```

**Trigger execution flow:**

1. A plugin calls `plugn trigger <trigger-name> [args...]`
2. plugn searches `$PLUGIN_ENABLED_PATH/*/` for files named `<trigger-name>`
3. Each matching executable is run with the provided arguments
4. Triggers can return data via stdout or signal errors via exit codes

**Key trigger categories:**

| Category | Examples | Purpose |
|----------|----------|---------|
| App lifecycle | `post-create`, `pre-delete`, `post-deploy` | React to app events |
| Build | `builder-detect`, `pre-build`, `builder-build` | Control build process |
| Deploy | `scheduler-deploy`, `check-deploy` | Manage deployments |
| Proxy | `proxy-build-config`, `nginx-pre-reload` | Configure reverse proxy |
| Git | `git-pre-pull`, `receive-app` | Handle git operations |

For the complete list of triggers, see [Plugin Triggers](/docs/development/plugin-triggers.md).

### Calling Triggers from Go Code

Go-based plugins use the `common.CallPlugnTrigger` function:

```go
result, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
    Trigger: "post-deploy",
    Args:    []string{appName, port, ip, imageTag},
})
```

## Command Flow

When a user runs a command like `dokku apps:list`, here's what happens:

```
User runs: ssh dokku@host apps:list
                    |
                    v
+-------------------+-------------------+
|           dokku bash script           |
|  1. Source /etc/default/dokku         |
|  2. Set DOKKU_ROOT, PLUGIN_PATH       |
|  3. Source common/functions           |
|  4. Call parse_args()                 |
|  5. Check user permissions            |
+-------------------+-------------------+
                    |
                    v
+-------------------+-------------------+
|          execute_dokku_cmd()          |
|  1. Handle plugin aliases             |
|  2. Check for subcommands/default     |
|  3. Check for subcommands/<cmd>       |
|  4. Fall back to commands scripts     |
+-------------------+-------------------+
                    |
                    v
+-------------------+-------------------+
|  plugins/apps/subcommands/list        |
|  (executes the actual command)        |
+-------------------+-------------------+
```

**Command resolution order:**

1. Check `$PLUGIN_ENABLED_PATH/<plugin>/subcommands/default` for the default command
2. Check `$PLUGIN_ENABLED_PATH/<plugin>/subcommands/<command>` for named commands
3. Iterate through all `$PLUGIN_ENABLED_PATH/*/commands` scripts as fallback

## Deployment Pipeline

When code is pushed via `git push dokku@host:myapp`, Dokku executes a multi-stage pipeline:

```
+-------------+    +-------------+    +-------------+    +-------------+
|   Receive   |--->|    Build    |--->|   Release   |--->|   Deploy    |
+-------------+    +-------------+    +-------------+    +-------------+
      |                  |                  |                  |
      v                  v                  v                  v
  git-hook          builder-detect    builder-release    scheduler-deploy
  receive-app       pre-build                            core-post-deploy
  post-extract      builder-build                        post-deploy
```

### Stage 1: Receive

The git receive stage handles the incoming push and prepares the source code:

1. `git-hook` receives the push and validates the branch
2. `receive-app` trigger is fired
3. Source code is extracted to a temporary directory
4. `post-extract` trigger allows modification of source

### Stage 2: Build

The build stage creates a Docker image from the source:

1. `builder-detect` determines the builder type (herokuish, pack, dockerfile, etc.)
2. `pre-build` trigger runs pre-build hooks
3. `builder-build` creates the Docker image
4. `post-build` trigger runs post-build hooks

### Stage 3: Release

The release stage prepares the image for deployment:

1. `pre-release-builder` allows image modifications
2. `builder-release` sets environment variables in the image
3. `post-release-builder` runs final release hooks

### Stage 4: Deploy

The deploy stage starts containers and configures networking:

1. `scheduler-deploy` starts new containers
2. `check-deploy` runs health checks
3. `core-post-deploy` switches traffic to new containers
4. `post-deploy` runs deployment tasks
5. Old containers are retired

## State Management

Dokku uses a file-based state system for simplicity and transparency.

### Property System

Plugin-specific configuration is stored in the property system:

```
/var/lib/dokku/data/<plugin>/<app>/<property>
```

Properties are managed via helper functions:

```bash
# Shell
fn-plugin-property-write "git" "$APP" "deploy-branch" "main"
fn-plugin-property-get "git" "$APP" "deploy-branch"
```

```go
// Go
common.PropertyWrite("git", appName, "deploy-branch", "main")
common.PropertyGet("git", appName, "deploy-branch")
```

### Per-App Data

Application-specific data sometimes lives in `$DOKKU_ROOT/<app>/`:

- `refs/` - Git references
- `tls/` - SSL certificates
- `ENV` - Environment file
- Container IDs, port mappings, etc.

All non-git code is being migrated to the property system.

### Global Configuration

Global settings can be configured via:

- `/etc/default/dokku` - System-level defaults
- `$DOKKU_ROOT/dokkurc` - User-level configuration
- `$DOKKU_ROOT/.dokkurc/*` - Additional configuration files

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Plugin-based architecture** | Enables extensibility without modifying core code. Community plugins can add databases, caching, and other services. |
| **Bash + Go hybrid** | Bash for orchestration and simple commands; Go for performance-critical operations and complex logic. |
| **Trigger system** | Loose coupling between plugins. Plugins don't need to know about each other; they just fire and respond to events. |
| **File-based state** | Simple, transparent, and easy to debug. No database dependency. State can be inspected with standard Unix tools. |
| **Docker as foundation** | Leverages Docker's container runtime, networking, and image management. Allows multiple scheduler backends. |

## Further Reading

- [Plugin Creation](/docs/development/plugin-creation.md) - How to create custom plugins
- [Plugin Triggers](/docs/development/plugin-triggers.md) - Complete list of available triggers
- [Testing](/docs/development/testing.md) - How to test Dokku and plugins
- [Application Deployment](/docs/deployment/application-deployment.md) - User-facing deployment guide
- [Builder Management](/docs/deployment/builders/builder-management.md) - Available build backends
- [Scheduler Management](/docs/deployment/schedulers/scheduler-management.md) - Available scheduler backends
