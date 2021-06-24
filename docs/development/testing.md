# Running Tests

Dokku has a full test suite to assist in quick iterating development. These tests include a linter using [shellcheck](https://github.com/koalaman/shellcheck), functional unit tests using the [Bats testing framework](https://github.com/bats-core/bats-core), and a deployment suite of example apps that use the most popular languages and frameworks.

We maintain the Dokku test harness within the `tests` directory:

- `tests/unit/*.bats`: Bats tests
- `tests/apps/`: Example applications that can be used for tests

## Continuous Integration

All pull requests have tests run against them on [GitHub Actions](https://github.com/features/actions), a continuous integration platform that provides Docker support for Ubuntu Trusty 18.04.

If you wish to skip tests for a particular commit, e.g. documentation changes, you may add the `[ci skip]` designator to your commit message. Commits that _should_ be tested but have the above designator will not be merged.

While we do provide official packages for a variety of platforms, as our test suite currently runs on Ubuntu Trusty 18.04, we only provide official installation support for that platform and the latest LTS release of Ubuntu (currently 20.04).

## Local Test Execution

- Setup Dokku in a [Vagrant VM](/docs/getting-started/install/vagrant.md).
- Run the following to setup tests and execute them:

  ```shell
  vagrant ssh
  sudo su -
  cd ~/dokku
  make ci-dependencies setup-deploy-tests

  # execute the entire test suite (linter, bats tests, and app deployment tests)
  make test

  # run linter
  make lint

  # execute all bats tests
  make unit-tests

  # execute all app deployment tests
  make deploy-tests
  ```

After making changes to your local Dokku clone, don't forget to update the Vagrant Dokku install.

```shell
# update vagrant dokku install from local git clone
make copyfiles

# build a specific plugin
make go-build-plugin copyplugin PLUGIN_NAME=apps
```

Additionally you may run a specific app deployment tests with a target similar to:

```shell
make deploy-test-nodejs-express
```

For a full list of test make targets check out `tests.mk` in the root of the Dokku repository.

## Executing a single test suite

When working on a particular plugin, it may be useful to run _only_ a particular test suite. This can be done by specifying the test suite path:

```shell
bats tests/unit/10_apps.bats
```

It is also possible to target multiple test suites at a time.

```shell
bats tests/unit/10_apps.bats tests/unit/10_certs.bats
```

## Executing a single test

In order to increase testing velocity, a wrapper script around Bats is available that can be used to run a single test case within a suite.

Tests within a suite may be listed by specifying the suite as a parameter to `bats`.

```shell
bats tests/unit/10_apps.bats
```

A single test can be specified via the `--filter` argument. The tests are selected via regex match, and all matches are executed.

```shell
bats --filter clone tests/unit/10_apps.bats
```
