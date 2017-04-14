# Dokku test suite

Dokku has a full test suite to assist in quick iterating development. These tests include a linter using [shellcheck](https://github.com/koalaman/shellcheck), functional unit tests using the [bats testing framework](https://github.com/sstephenson/bats), and a deployment suite of example apps that use the most popular languages and frameworks.

We maintain the Dokku test harness within the `tests` directory:

- `tests/unit/*.bats`: Bats tests
- `tests/apps/`: Example applications that can be used for tests

## Continuous Integration

All pull requests have tests run against them on [CircleCI](https://circleci.com/), a continuous integration platform that provides Docker support for Ubuntu Trusty 14.04.

If you wish to skip tests for a particular commit - e.g. Documentation changes - you may add the `[ci skip]` designator to your commit message. Commits that *should* be tested but have the above designator will not be merged.

While we do provide official packages for a variety of platforms, as our test suite currently runs on Ubuntu Trusty 14.04, we only provide official installation support for that platform.

## Local Test Execution

- Setup Dokku in a [vagrant vm](/docs/getting-started/install/vagrant.md)
- Run the following to setup tests and execute them:

  ```shell
  vagrant ssh
  sudo su -
  cd ~/dokku
  make ci-dependencies setup-deploy-tests

  # execute the entire test suite (linter, bats tests, and app deployment tests)
  make test

  # run linter & update vagrant Dokku install from local git clone
  make lint copyfiles

  # execute all bats tests
  make unit-tests

  # execute all app deployment tests
  make deploy-tests
  ```

Additionally you may run a specific app deployment tests with a target similar to:

```shell
make deploy-test-nodejs-express
```

For a full list of test make targets check out `tests.mk` in the root of the Dokku repository.
