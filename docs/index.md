# Dokku Documentation

Docker powered mini-Heroku. The smallest PaaS implementation you've ever seen.

## Table Of Contents

### Getting Started

- [Installation](http://progrium.viewdocs.io/dokku/installation)
- [Upgrading](http://progrium.viewdocs.io/dokku/upgrading)
- [Troubleshooting](http://progrium.viewdocs.io/dokku/troubleshooting)

### Deployment

- [Application Deployment](http://progrium.viewdocs.io/dokku/deployment/index)
- [Configuration management](http://progrium.viewdocs.io/dokku/deployment/configuration)
- [Process management](http://progrium.viewdocs.io/dokku/deployment/process-management)
- [DNS Configuration](http://progrium.viewdocs.io/dokku/deployment/dns)
- [Nginx Configuration](http://progrium.viewdocs.io/dokku/deploymentnginx)
- [Running Remote commands](http://progrium.viewdocs.io/dokku/deployment/remote-commands)

### Community Contributions

- [Dokku Client](http://progrium.viewdocs.io/dokku/community/client)
- [Plugins](http://progrium.viewdocs.io/dokku/community/plugins)

### Development

- [Contributing](http://progrium.viewdocs.io/dokku/development/contributing)
- [Plugin Creation](http://progrium.viewdocs.io/dokku/development/plugin-creation)
- [Pluginhooks](http://progrium.viewdocs.io/dokku/development/pluginhooks)

## Things this project won't do

 * **Multi-host.** It runs on one host. If you need more, have a look at [Deis](http://deis.io/).
 * **Multitenancy.** Multi-app, and loosely multi-user based on SSH keys, but that's it.
 * **Client app.** Given the constraints, running commands remotely via SSH is fine.

## Sponsors

Though we love everybody doing open source, we especially love [Deis](http://deis.io/) for sponsoring Dokku.
