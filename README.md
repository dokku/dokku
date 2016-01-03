# Dokku [![Build Status](https://img.shields.io/circleci/project/dokku/dokku/master.svg?style=flat-square "Build Status")](https://circleci.com/gh/dokku/dokku/tree/master) [![Ubuntu Package](https://img.shields.io/badge/package-ubuntu-brightgreen.svg?style=flat-square "Ubuntu Package")](https://packagecloud.io/dokku/dokku) [![IRC Network](https://img.shields.io/badge/irc-freenode-blue.svg?style=flat-square "IRC Freenode")](https://webchat.freenode.net/?channels=dokku) [![Documentation](https://img.shields.io/badge/docs-viewdocs-blue.svg?style=flat-square "Viewdocs")](http://dokku.viewdocs.io/dokku/) [![Gratipay](https://img.shields.io/gratipay/dokku.svg?style=flat-square)](https://gratipay.com/dokku/)

Docker powered mini-Heroku. The smallest PaaS implementation you've ever seen. Sponsored by our friends at [Deis](http://deis.io/).

## Requirements

- A fresh VM running Ubuntu `14.04 x64`

## Installing

To install the latest stable release, you can run the following commands as a user that has access to `sudo`:

    wget https://raw.githubusercontent.com/dokku/dokku/v0.4.9/bootstrap.sh
    sudo DOKKU_TAG=v0.4.9 bash bootstrap.sh

### Upgrading

[View the docs for upgrading](http://dokku.viewdocs.io/dokku/upgrading) from an older version of Dokku.

## Documentation

Full documentation - including advanced installation docs - are available online at [docs](http://dokku.viewdocs.io/dokku/)

## Support

You can use [Github Issues](https://github.com/dokku/dokku/issues), check [Troubleshooting](http://dokku.viewdocs.io/dokku/troubleshooting) in the documentation, or join us on [freenode in #dokku](https://webchat.freenode.net/?channels=%23dokku)

## Contribution

After checking [Github Issues](https://github.com/dokku/dokku/issues), the [Troubleshooting Guide](http://dokku.viewdocs.io/dokku/troubleshooting) or having a chat with us on [freenode in #dokku](https://webchat.freenode.net/?channels=%23dokku), feel free to fork and create a Pull Request.

While we may not merge your PR as is, they serve to start conversations and improve the general dokku experience for all users.

## Sponsors

Dokku is currently sponsored by the enterprise grade, multi-host PaaS project [Deis](http://deis.io/).

## License

[MIT License](https://github.com/dokku/dokku/blob/master/LICENSE) © Jeff Lindsay
