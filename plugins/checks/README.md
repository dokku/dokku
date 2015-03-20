# Dokku core checks plugin

This plugin handles zero-downtime functionality by checking the application
container against a list of checks specified in CHECKS file.

The CHECKS file may contain empty lines, comments (lines starting with #),
settings (NAME=VALUE) and check instructions.

The format of a check instruction is a path, optionally followed by the
expected content.  For example:
  ```shell
   /                       My Amazing App
   /stylesheets/index.css  .body
   /scripts/index.js       $(function()
   /images/logo.png
   ```

To check an application that supports multiple hostnames, use relative URLs
that include the hostname, for example:
  ```shell
  //admin.example.com     Admin Dashboard
  //static.example.com/logo.png
  ```

You can also specify the protocol to explicitly check HTTPS requests.

The default behavior is to wait for 5 seconds before running the first check,
and timeout each check to 30 minutes.

You can change these by setting DOKKU_CHECKS_WAIT and DOKKU_CHECKS_TIMEOUT to different values, for
example:
   ```shell
   DOKKU_CHECKS_WAIT=30     # Wait 1/2 minute
   DOKKU_CHECKS_TIMEOUT=60  # Timeout after a minute
   ```
