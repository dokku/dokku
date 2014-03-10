# Dokku core

This plugin provides the base Dokku installation steps, as well as the `help`
command for many commands implemented as plugins (in lieu of a proper `help`
command system for plugins).

## Dokku core installation

- Sets the dokku HOSTNAME file from the server's hostname, if missing.
- Creates an Upstart rule to restart all apps after a reboot (redeploying them
  to refresh their port mappings for virtual hosting).
