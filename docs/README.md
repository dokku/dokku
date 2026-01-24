# Dokku Documentation

This documentation covers the installation, configuration, and usage of Dokku - a Docker-powered PaaS that provides a Heroku-like experience.

## Getting Started

- [Installation](getting-started/installation/index.md) - Install Dokku on your server
    - [Debian](getting-started/install/debian.md)
    - [Docker](getting-started/install/docker.md)
    - [DigitalOcean](getting-started/install/digitalocean.md)
    - [Azure](getting-started/install/azure.md)
    - [DreamHost](getting-started/install/dreamhost.md)
    - [Vagrant](getting-started/install/vagrant.md)
- [Advanced Installation](getting-started/advanced-installation.md) - Custom installation options
- [Upgrading](getting-started/upgrading/index.md) - Upgrade to newer Dokku versions
- [Uninstalling](getting-started/uninstalling.md) - Remove Dokku from your server
- [Troubleshooting](getting-started/troubleshooting.md) - Common issues and solutions
- [Where to Get Help](getting-started/where-to-get-help.md) - Support resources

## Deployment

- [Application Deployment](deployment/application-deployment.md) - Deploy your first app
- [Application Management](deployment/application-management.md) - Manage deployed applications
- [Logs](deployment/logs.md) - View application logs
- [Remote Commands](deployment/remote-commands.md) - Run commands remotely
- [User Management](deployment/user-management.md) - Manage SSH access
- [Zero Downtime Deploys](deployment/zero-downtime-deploys.md) - Deploy without interruption

### Deployment Methods

- [Git](deployment/methods/git.md) - Deploy via git push
- [Archive](deployment/methods/archive.md) - Deploy from tar/zip archives
- [Image](deployment/methods/image.md) - Deploy from Docker images

### Builders

- [Builder Management](deployment/builders/builder-management.md) - Configure build systems
- [Buildpack Management](deployment/builders/buildpack-management.md) - Manage buildpack settings
- [Herokuish Buildpacks](deployment/builders/herokuish-buildpacks.md) - Heroku-compatible buildpacks
- [Cloud Native Buildpacks](deployment/builders/cloud-native-buildpacks.md) - CNB support
- [Dockerfiles](deployment/builders/dockerfiles.md) - Build from Dockerfile
- [Nixpacks](deployment/builders/nixpacks.md) - Nixpacks builder
- [Railpack](deployment/builders/railpack.md) - Railpack builder
- [Lambda](deployment/builders/lambda.md) - Lambda-style functions
- [Null Builder](deployment/builders/null.md) - Skip the build phase

### Schedulers

- [Scheduler Management](deployment/schedulers/scheduler-management.md) - Configure schedulers
- [Docker Local](deployment/schedulers/docker-local.md) - Default Docker scheduler
- [K3s](deployment/schedulers/k3s.md) - Lightweight Kubernetes
- [Kubernetes](deployment/schedulers/kubernetes.md) - External Kubernetes clusters
- [Nomad](deployment/schedulers/nomad.md) - HashiCorp Nomad
- [Null Scheduler](deployment/schedulers/null.md) - Skip scheduling

### Continuous Integration

- [Generic CI](deployment/continuous-integration/generic.md) - General CI/CD setup
- [GitHub Actions](deployment/continuous-integration/github-actions.md)
- [GitLab CI](deployment/continuous-integration/gitlab-ci.md)
- [Woodpecker CI](deployment/continuous-integration/woodpecker-ci.md)

## Configuration

- [Environment Variables](configuration/environment-variables.md) - Configure app environment
- [Domains](configuration/domains.md) - Custom domain management
- [SSL/TLS](configuration/ssl.md) - HTTPS and certificates

## Networking

- [Proxy Management](networking/proxy-management.md) - Reverse proxy configuration
- [Port Management](networking/port-management.md) - Port mapping and exposure
- [DNS](networking/dns.md) - DNS configuration
- [Network](networking/network.md) - Docker network management

### Proxies

- [Nginx](networking/proxies/nginx.md) - Default proxy
- [Caddy](networking/proxies/caddy.md)
- [HAProxy](networking/proxies/haproxy.md)
- [Traefik](networking/proxies/traefik.md)
- [OpenResty](networking/proxies/openresty.md)

## Processes

- [Process Management](processes/process-management.md) - Scale and manage processes
- [Scheduled Cron Tasks](processes/scheduled-cron-tasks.md) - Scheduled jobs
- [One-off Tasks](processes/one-off-tasks.md) - Run one-time commands
- [Entering Containers](processes/entering-containers.md) - Access running containers

## Advanced Usage

- [Persistent Storage](advanced-usage/persistent-storage.md) - Mount volumes
- [Docker Options](advanced-usage/docker-options.md) - Custom Docker run arguments
- [Resource Management](advanced-usage/resource-management.md) - CPU and memory limits
- [Plugin Management](advanced-usage/plugin-management.md) - Install and manage plugins
- [Registry Management](advanced-usage/registry-management.md) - Docker registry integration
- [Repository Management](advanced-usage/repository-management.md) - Git repository settings
- [Deployment Tasks](advanced-usage/deployment-tasks.md) - Pre/post deploy hooks
- [Event Logs](advanced-usage/event-logs.md) - Dokku event history
- [Backup and Recovery](advanced-usage/backup-recovery.md) - Data backup strategies

## Development

- [Architecture](development/architecture.md) - Internal architecture overview for contributors
- [Plugin Creation](development/plugin-creation.md) - Build custom plugins
- [Plugin Triggers](development/plugin-triggers.md) - Available plugin hooks
- [Testing](development/testing.md) - Test Dokku and plugins
- [Release Process](development/release-process.md) - How Dokku releases work

## Community

- [Plugins](community/plugins.md) - Community-maintained plugins
- [Clients](community/clients.md) - API clients and tools

## Enterprise

- [Dokku Pro](enterprise/pro.md) - Commercial features and support

## Appendices

### Migration Guides

- [0.37.0](appendices/0.37.0-migration-guide.md)
- [0.36.0](appendices/0.36.0-migration-guide.md)
- [0.35.0](appendices/0.35.0-migration-guide.md)
- [0.34.0](appendices/0.34.0-migration-guide.md)
- [0.33.0](appendices/0.33.0-migration-guide.md)
- [0.32.0](appendices/0.32.0-migration-guide.md)
- [0.31.0](appendices/0.31.0-migration-guide.md)
- [0.30.0](appendices/0.30.0-migration-guide.md)
- [0.29.0](appendices/0.29.0-migration-guide.md)
- [0.28.0](appendices/0.28.0-migration-guide.md)
- [0.27.0](appendices/0.27.0-migration-guide.md)
- [0.26.0](appendices/0.26.0-migration-guide.md)
- [0.25.0](appendices/0.25.0-migration-guide.md)
- [0.24.0](appendices/0.24.0-migration-guide.md)
- [0.23.0](appendices/0.23.0-migration-guide.md)
- [0.22.0](appendices/0.22.0-migration-guide.md)
- [0.21.0](appendices/0.21.0-migration-guide.md)
- [0.20.0](appendices/0.20.0-migration-guide.md)
- [0.10.0](appendices/0.10.0-migration-guide.md)
- [0.9.0](appendices/0.9.0-migration-guide.md)
- [0.8.0](appendices/0.8.0-migration-guide.md)
- [0.7.0](appendices/0.7.0-migration-guide.md)
- [0.6.0](appendices/0.6.0-migration-guide.md)
- [0.5.0](appendices/0.5.0-migration-guide.md)

### File Formats

- [app.json](appendices/file-formats/app-json.md) - Application configuration
- [Procfile](appendices/file-formats/procfile.md) - Process definitions
- [Dockerfile](appendices/file-formats/dockerfile.md) - Docker build instructions
- [.buildpacks](appendices/file-formats/buildpacks-file.md) - Buildpack configuration
- [project.toml](appendices/file-formats/project-toml.md) - CNB configuration
- [nixpacks.toml](appendices/file-formats/nixpacks-toml.md) - Nixpacks configuration
- [railpack.json](appendices/file-formats/railpack-json.md) - Railpack configuration
- [lambda.yml](appendices/file-formats/lambda-yml.md) - Lambda configuration
- [nginx.conf.sigil](appendices/file-formats/nginx-conf-sigil.md) - Nginx template customization
