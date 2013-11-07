# User interface
A few commands are available to manage add-ons : 
* `dokku addons` Lists all add-ons an application has.
* `dokku addons:enable` Enable a add-on. This installs required dependencies, creates required files, ...
* `dokku addons:disable` Remove a add-on from all applications and disable it. This removes unnecessary files, ...
* `dokku addons:add` Adds an add-on to an application. This causes the add-on to allocate resources for this application (eg an account, database, ..). The add-on must be enabled first.
* `dokku addons:remove` Removes an add-on from an application. This causes the add-on to destroy all associated resources.
* `dokku addons:url` Shows the URL an add-on has provided. This is only meant for debugging. In production, there's no need to read this url manually, as it is set automatically in an environment variable.

The URL is passed to the application through the `DOKKU_${ADDON}_URL` where `${ADDON}` is the name of the add-on in uppercase.

# Add-on development
Add-ons are located in `/var/lib/dokku/add-ons/`. Every subfolder is an add-on.
When they provision an app, add-ons should generate a unique ID, and a "private" value.
The ID is used to identify the app within the add-on. This would typically be a username or database name.
The "private" value is used to generate the URL. This would be a password for example.

IDs and private values can take any value. They however must not include semicolons, as it is used internally by the plugin. (See the internal section)

Add-ons must provide a few executable files : 
* `$ADDON/enable` This script takes no arguments. It should setup everything the add-on needs, eg install dependencies, create docker containers, ...
* `$ADDON/disable` This script takes no arguments. It should clean up all the data used by the add-on, eg destroy docker containers. All applications have already been unprovisonised before this script is called.
* `$ADDON/provision` This script takes one argument, the app name. It should only be used to generate understandable IDs, which contain the app name. Add-ons shouldn't use it the access the app's config files, ...
The script should output the generated id and private value on stdout, separated by a semicolon.
* `$ADDON/unprovision` This script takes one argument, the ID. It should destroy all resources associated with this ID.
* `$ADDON/url` This script takes two arguments, the ID and the private value. It should output the url on stdout.

Apart from `$ADDON/enable` all scripts are run only with the add-on in enabled state.
Add-on's are free to run the service in the way they like. They can run it on the cloud, on the host, on a single docker container or on a container per provisioned app. The plugin doesn't care, as long as the add-on provides a URL which is accessible.
Because URLs might change (docker can assign different IPs/ports after reboot), the `url` script is called each time the app is released. (BTW, this means we should release all apps at startup, rather than deploy them)

## Guidelines
### Add-ons internal files
Add-ons may want to store files for their own internal use, for eg. database storage.
Before any add-on script is ran, a `$ADDON_DATA` environment variable is exported, which contains the path to a directory where the add-on can safely save any file. The directory is guaranteed to exist before add-on scripts are ran.
Currently it is set to `$DOKKU_ROOT/.add-ons/$ADDON` where `$ADDON` is the add-on's name. However, this might change, therefore add-ons should use the `$ADDON_DATA` variable rather than hardcoding it.

The add-on plugin automatically creates and removes a `$ADDON_DATA/enabled` file to keep track of which plugin is installed. Add-ons mustn't create, modify nor delete this file themselves.
It is created after the `$ADDON/enable` script completes, and deleted after the `$ADDON/disabled` script completes.

# Internals
The add-on plugin uses the `$APP/ADDONS` file to save which add-ons are in use.
Each line has the following format : 

    name;id;private

In the `pre-release` hook, the plugin generates environment variables based on the ids and private values, using the `url` script, and saves them to `$APP/ENV`

