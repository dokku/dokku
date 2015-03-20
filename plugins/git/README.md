# Dokku core git handling

This plugin provides the `git` interface for creating and deploying apps in
Dokku by adding commands to Dokku to wrap Git's SSH interface, as well as
defining a hook to create the relevant git pre-receive hook for any apps being
imported from backup.

