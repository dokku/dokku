# User interface
Four commands are available to manage plugins : 
* `dokku addons` Lists all addons an application has.
* `dokku addons:add` Adds an addon to an application. This causes the addon to allocate resources for this application (eg an account, database, ..)
* `dokku addons:url` Shows the URL a plugin has given. This is only meant for debugging. In production, there's no need to read this url manually, as it is provided in an environment variable.
* `dokku addons:remove` Removes an addon from an application. This causes the addon to destroy all associated resources.

The URL is passed to the application through the `DOKKU_${ADDON}_URL` where `${ADDON}` is the name of the add-on in uppercase.

# Add-on development
Add-ons are located in `/var/lib/dokku/addons/`. Every subfolder is an add-on.
When they provision an app, add-ons should generate a unique ID, and a "private" value.
The ID is used to identify the app within the add-on. This would typically be a username or database name.
The "private" value is used to generate the URL. This would be a password for example.

IDs and private values can take any value. They however must not include semicolons, as it is used internally by the plugin. (See the internal section)

Add-ons must provide three executable files : 
* `$ADDON/provision` This script takes one argument, the app name. It should only be used to generate understandable IDs, which contain the app name. Add-ons shouldn't use it the access the app's config files, ...
The script should output the generated id and private value on stdout, separated by a semicolon.
* `$ADDON/unprovision` This script takes one argument, the ID. It should destroy all resources associated with this ID.
* `$ADDON/url` This script takes two arguments, the ID and the private value. It should output the url on stdout.

Add-on's are free to run the service in the way they like. They can run it on the cloud, on the host, on a single docker container or on a container per provisioned app. The plugin doesn't care, as long as the add-on provides a URL which is accessible.
Because URLs might change (docker can assign different IPs/ports after reboot), the `url` script is called each time the app is released. (BTW, this means we should release all apps at startup, rather than deploy them)

## Guidelines
### Add-ons internal files
Add-ons may want to store files for their own internal use, for eg. database storage.
Before any add-on script is ran, a `$ADDON_ROOT` environment variable is exported, which contains the path to a directory where the addon can safely save any file. The directory is guaranteed to exist before add-on scripts are ran.
Currently it is set to `$DOKKU_ROOT/.addons/$ADDON` where `$ADDON` is the addon's name. However, this might change, therefore plugins should use the `$ADDON_ROOT` variable rather than hardcoding it.


# Internals
The add-on plugin uses the `$APP/ADDONS` file to save which add-ons are in use.
Each line has the following format : 

    name;id;private

In the `pre-release` hook, the plugin generates environment variables based on the ids and private values, using the `url` script, and saves them to `$APP/ENV`
