# Dokku core backup system

This plugin provides the commands for exporting and importing backups of a
Dokku configuration (including the list of apps and any plugin configurations
for those apps), as well as hooks for the backup of core Dokku configuration
files.

Note that this does *not* backup the content of the apps themselves: to back up
your apps, you should be using another solution, such as pushing your apps to a
Git backup server. Similarly, you should have a relevant backup solution for
any data you may be keeping through a datastore plugin.

