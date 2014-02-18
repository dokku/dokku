# Dokku core delete command

This plugin provides the `delete` command, which stops the container for a
running app and deletes its Docker image.

Additionally, the pre-delete hook in this plugin deletes the app's build cache,
and the post-delete hook deletes the app's repository directory.
