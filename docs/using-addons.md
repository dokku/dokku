# Addons
Addons provide services. They are similar to heroku's addons.

Heroku provides a good description of how to use and manage addons.
Most commands are similar in dokku.

https://devcenter.heroku.com/articles/managing-add-ons

## Usage
In the following examples, `myapp` is the name of an application, `mariadb` is the addon's name.

First, you should list all available addons :

    dokku addons -a

An addon marked as disabled may not be used by applications. You must enable it first :

    dokku addons:enable mariadb

Depending on the addon, you might need to type your password into sudo.

Once the addon is enabled, add it to your application :

    dokku addons:add myapp mariadb

Congratulations, you can now use a mariadb database in your application.
The url to which your application should connect is located in the `DOKKU_MARIADB_URL`.
You can visualize this url by running the `addons:url` command.

    dokku addons:url myapp mariadb

The contents and format of the url is addon-specific.
Note that the indicated host is from the application's container point of view.
It may not be accessible from outside.
In the case of the mariadb addon, it has the following format :

    mysql://USER@PASSWORD/HOST:PORT/DATABASE


You can remove an addon from your application :

    dokku addons:remove myapp mariadb

WARNING: This removes all data (such as a database) associated with this application.
Use with care.

If none of your applications uses an addon, you can disable it :

    dokku addons:disable mariadb

