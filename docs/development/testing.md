# Dokku test suite

Dokku now has a full test suite to assist in quick iterating development. These tests include a linter using [shellcheck](https://github.com/koalaman/shellcheck), functional unit tests using the [bats testing framework](https://github.com/sstephenson/bats), and a deployment suite of example apps that use the most popular languages and frameworks.

Bats tests can be found here:
  ```
  tests/unit/*.bats
  ```

Example apps can be found here:
  ```
  tests/apps/
  ```

### Executing tests locally

- Setup dokku in a [vagrant vm](http://progrium.viewdocs.io/dokku/getting-started/install/vagrant)
- Test setup and execution

  ```shell
  $ vagrant ssh
  $ sudo su -
  $ cd ~/dokku
  $ make ci-dependencies setup-deploy-tests
  $ make test  # run the entire test suite (linter, bats tests, and app deployment tests)
  $
  $ make lint  # linter
  $ make unit-tests  # bats tests
  $ make deploy-tests  # app deployment tests
  ```
- Additionally you may run a specific app deployment tests with a target similar to:

  ```shell
  make deploy-test-nodejs-express
  ```
- For a full list of test make targets check out `tests.mk` in the root of the dokku repository.
