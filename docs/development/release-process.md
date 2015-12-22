# Release Process

Dokku is released in intervals *at most* three weeks apart, though may be released much quicker.

To propose a release, the following tasks need to be performed:

- The installable version must be changed in the `contrib/dokku-installer.py` file.
- The installable version must be changed in the `debian/control` file.
- The installable version must be changed in the `docs/home.html` file
- The installable version must be changed in the `docs/index.md` file
- The installable version must be changed in the `docs/installation.md` file.
- The installable version must be changed in the `docs/template.html` file.
- The installable version must be changed in the `README.md` file.
- The embedded css should be cleared in the `docs/template.html` file.
- The versioned links should be updated in the `docs/assets/favicons/browserconfig.xml` file.
- The versioned links should be updated in the `docs/assets/favicons/manifest.json` file.
- The versioned links should be updated in the `docs/assets/style.css` file.
- The versioned links should be updated in the `docs/home.html` file.
- The versioned links should be updated in the `docs/template.html` file.
- The versioned links should be updated in the `docs/template.html` file.
- A list of changes must be made in the `HISTORY.md`.
- A tag must be created locally with your release version
- Debian packages *must* be created via `vagrant up build`
- The packages should be uploaded to packagecloud.io
- All changes are pushed to master and the tag should be turned into a release which will contain the changelog.

## Versioning

Dokku follows semver standards. As we are not yet at a stable release, breaking changes will require *only* a minor release, while all other changes only require a patch release. Once we hit stable, breaking changes will require a major release.

Tags should be created via the following method:

```shell
git tag v0.9.9
```

At the moment, tags need not be signed, though that may change in the future.

## Debian Packages

The `build` target in the dokku `Vagrantfile` creates debian packages for dokku at a point in time. The version will be based upon the latest local tag - you may create your own, internal tags/releases if that is so desired.

Debian package information is held in the `debian` directory of the dokku project.

For the public project, releases should be pushed to packagecloud.io *after* a tag is created but *before* said tag is pushed to github. The following may be the release workflow:


```shell
git tag v0.9.9
vagrant up build
export PACKAGECLOUD_TOKEN=SOME_TOKEN
package_cloud push dokku/dokku/ubuntu/trusty dokku_0.9.9_amd64.deb
```

If new versions of other packages were created, these should also be pushed at this time.

> If you are a maintainer and need the PACKAGECLOUD_TOKEN in order to make a release, please contact @josegonzalez to get this information.

## Changelog format

The `HISTORY.md` should be added to based on the changes made since the previous release. This can be done by reviewing all merged pull requests to the master branch on github. The format is as follows:

```
## 0.9.9

Some description concerning major changes in this release, or potential incompatibilities.

### New Features

- #SOME_ID: @pull-request-creator Description

### Bug Fixes

- #SOME_ID: @pull-request-creator Description

### Docs Changes

- #SOME_ID: @pull-request-creator Description
```
