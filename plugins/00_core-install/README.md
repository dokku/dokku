# Dokku core installation

This plugin provides a couple components of Dokku's setup:

- Sets the dokku HOSTNAME file from the server's hostname, if missing.
- Creates an Upstart rule to restart all apps after a reboot (redeploying them
  to refresh their port mappings for virtual hosting).
