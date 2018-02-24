# Release Process

Dokku is released in intervals *at most* three weeks apart, though may be released much quicker.

To propose a release, the following tasks need to be performed:

- Update all blockquote references of `not yet released and only available in master` to point to release version.
- The installable version must be changed in the `contrib/dokku-installer.py` file.
- The installable version must be changed in the `debian/control` file.
- The installable version must be changed in the `docs/home.html` file
- The installable version must be changed in the `docs/installation.md` file.
- The installable version must be changed in the `docs/template.html` file.
- The installable version must be changed in the `README.md` file.
- The embedded css should be cleared in the `docs/template.html` file.
- The versioned links should be updated in the `docs/assets/favicons/browserconfig.xml` file.
- The versioned links should be updated in the `docs/assets/favicons/manifest.json` file.
- The versioned links should be updated in the `docs/assets/style.css` file.
- The versioned links should be updated in the `docs/home.html` file.
- The versioned links should be updated in the `docs/template.html` file.
- The versioned links should be updated or added to the `docs/assets/versions.json` file.
- A list of changes must be made in the `HISTORY.md`.
- A tag must be created locally with your release version
- Debian and RPM packages *must* be created via `vagrant up build`
- The packages should be uploaded to packagecloud.io
- All changes are pushed to master and the tag should be turned into a release which will contain the changelog.
- ArchLinux package description *must* be updated via `vagrant up build-arch` (needs to be done after the tag is pushed to GitHub, because it is based on that)

## Versioning

Dokku follows semver standards. As we are not yet at a stable release, breaking changes will require *only* a minor release, while all other changes only require a patch release. Once we hit stable, breaking changes will require a major release.

Tags should be created via the following method:

```shell
git tag v0.9.9
```

At the moment, tags need not be signed, though that may change in the future.

## Debian and RPM packages

The `build` target in the Dokku `Vagrantfile` creates debian and rpm packages for Dokku at a point in time. The version will be based upon the latest local tag - you may create your own, internal tags/releases if that is so desired.

Debian package information is held in the `debian` directory of the Dokku project.

For the public project, releases should be pushed to packagecloud.io *after* a tag is created but *before* said tag is pushed to github. The following may be the release workflow:


```shell
git tag v0.9.9
vagrant up build
export PACKAGECLOUD_TOKEN=SOME_TOKEN
package_cloud push dokku/dokku/ubuntu/trusty dokku_0.9.9_amd64.deb
package_cloud push dokku/dokku/el/7 dokku-0.9.9-1.x86_64.rpm
```

If new versions of other packages were created, these should also be pushed at this time.

> If you are a maintainer and need the PACKAGECLOUD_TOKEN in order to make a release, please contact @josegonzalez to get this information.

## ArchLinux Packages

ArchLinux packages are not really build, because all that is needed for an Arch User Repo (AUR) package is the description of how to build the package. To make this process as easy as possible there is a vagrant box called `build-arch` that updates the version of this build description (a file called `PKGBUILD`), then runs some helper scripts to fill all additional information and does test if the package could be build. Then only those changes need to be pushed to the AUR repo and an updated version of the package is ready for usage for our ArchLinux users. For detailed information see the section below.

The workflow looks like this:

```shell
# having dokku-arch in ../dokku-arch
vagrant up build-arch
# wait for "==> build-arch: ==> Finished making: dokku 0.11.4-2 (Mon Feb 22 23:20:37 CET 2016)"
cd ../dokku-arch
git add PKGBUILD .SRCINFO
git commit -m 'Update to dokku 0.9.9'
git push aur master
```

> If you are a maintainer and need access to the AUR repositories in order to make a release, please contact @morrisjobke or @josegonzalez to get this co-maintainership.

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

## Detailed information for ArchLinux packages

All of the information to build the ArchLinux package is in the AUR git repository (see [dokku AUR page](https://aur.archlinux.org/packages/dokku/)). The release of a AUR package only consists of pushing the package information into the AUR git repo. Then users could use that information to build the package on their machines.

To update the package clone the repository and adjust the files in the repository. Then a helper script - `updpkgsums` - to update the SHA sum could  be called (check against the original SHA sum). Another helper script - `mksrcinfo` - needs to be called to update the meta information of the package in a file called `.SRCINFO`. The next step builds the package locally for verification - `makepkg`. As last step commit your changes and push the commit.

* dependencies are defined in the `depends` attribute in `PKGBUILD`
* build steps during package build time are defined in the `package()` method in `PKGBUILD`
* steps that should be executed during install/update/remove time are defined in the file `dokku.install`
* detailed information about all attributes in `PKGBUILD` could be found in the [ArchLinux wiki](https://wiki.archlinux.org/index.php/PKGBUILD)
* detailed information about the AUR workflow could be found in the [AUR article](https://wiki.archlinux.org/index.php/Arch_User_Repository) in the ArchLinux wiki

That is the usual workflow:

```shell
updpkgsums # update sha sums - compare them with the original ones
mksrcinfo # update package metadata for AUR
makepkg # test package builds
git add PKGBUILD .SRCINFO
git commit -m 'Update to dokku 0.9.9'
git push
```

> If there is something unclear simply ask @morrisjobke for help.
