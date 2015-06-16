
# Dokku

Docker powered mini-Heroku. The smallest PaaS implementation you've ever seen. Sponsored by our friends at [Deis](http://deis.io/).

## Requirements

- A fresh VM running Ubuntu `14.04 x64`

## Installing

To install the latest frog staging version of dokku, you can run the following bootstrapper command as a user that has `sudo`:

    $ wget -qO- https://raw.githubusercontent.com/frog-eXPeriMeNTaL/dokku/frog-production/bootstrap.sh
    $ sudo DOKKU_BRANCH=frog-staging bash

To install the latest frog production version of dokku, you can run the following bootstrapper command:

    $ wget -qO- https://raw.githubusercontent.com/frog-eXPeriMeNTaL/dokku/frog-production/bootstrap.sh
    $ sudo DOKKU_BRANCH=frog-production bash

### Upgrading

[View the docs for upgrading](http://progrium.viewdocs.io/dokku/upgrading) from an older version of Dokku.

## Documentation

Full documentation - including advanced installation docs - are available online at [docs](http://progrium.viewdocs.io/dokku/index)

## Support

You can use [Github Issues](https://github.com/progrium/dokku/issues), check [Troubleshooting](http://progrium.viewdocs.io/dokku/troubleshooting) in the documentation, or join us on [freenode in #dokku](https://webchat.freenode.net/?channels=%23dokku)

## Contribution

After checking [Github Issues](https://github.com/progrium/dokku/issues), the [Troubleshooting Guide](http://progrium.viewdocs.io/dokku/troubleshooting) or having a chat with us on [freenode in #dokku](https://webchat.freenode.net/?channels=%23dokku), feel free to fork and create a Pull Request.

While we may not merge your PR as is, they serve to start conversations and improve the general dokku experience for all users.

## Sponsors

Dokku is currently sponsored by the enterprise grade, multi-host PaaS project [Deis](http://deis.io/).

## License

MIT
