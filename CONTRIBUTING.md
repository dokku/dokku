# How to contribute

The Dokku project would love to welcome your contributions. There are
several ways to help out:

* Create an [issue](https://github.com/dokku/dokku/issues) on GitHub, if you
  have found a bug
* Write [test cases](https://dokku.com/docs/development/testing/) for open bug issues
* Write patches for open bug/feature issues, preferably with test cases
  included
* Contribute to the [documentation](https://dokku.com/docs/)
* Come up with new ways, non-commercial to show off our [lovely logo](https://avatars1.githubusercontent.com/u/13455795)
* Blog about different ways you are using dokku
* Sponsor the Dokku project financially on [OpenCollective](https://opencollective.com/dokku#support) or [Patreon](https://www.patreon.com/dokku)

There are a few guidelines that we need contributors to follow so that we have
a chance of keeping on top of things.

## Topics

* [Reporting Security Issues](#reporting-security-issues)
* [Reporting Issues](#reporting-other-issues)
* [Contributing](#contributing)
  * [Making Changes](#making-changes)
  * [Which branch to base the work](#which-branch-to-base-the-work)
  * [Submitting Changes](#submitting-changes)
  * [When will my change be merged?](#when-will-my-change-be-merged)
  * [Running tests locally](#running-tests-locally)
* [Additional Resources](#additional-resources)

## Reporting security issues

The Dokku maintainers take security seriously. If you discover a security
issue, please bring it to their attention right away!

Please **DO NOT** file a public issue, instead send your report privately to
[dokku@josediazgonzalez.com](mailto:dokku@josediazgonzalez.com),

Security reports are greatly appreciated and we will publicly thank you for it.

## Reporting other issues

A great way to contribute to the project is to send a detailed report when you
encounter an issue. We always appreciate a well-written, thorough bug report,
and will thank you for it!

Sometimes  Dokku  is missing a feature you need. In some cases, those features can
be found in pre-existing [plugins](https://dokku.com/docs/plugins/),
and we encourage our users to create and contribute such packages. From time to
time, we will also pull plugins into the Dokku core when the task they solve is
a common one for our users.

Check that [our issue database](https://github.com/dokku/dokku/issues)
doesn't already include that problem or suggestion before submitting an issue.
If you find a match, add a quick "+1" or "I have this problem too." Doing this
helps prioritize the most common problems and requests.

When reporting issues, please include all the information we ask for in our
[issue template](https://github.com/dokku/dokku/blob/master/ISSUE_TEMPLATE.md).
Not doing so will prolongue the support period, making it more difficult to support
you.

## Contributing

Before you contribute to the Dokku project, there are a few things that you'll
need to do

* Make sure you have a [GitHub account](https://github.com/signup/free).
* Submit an [issue](https://github.com/dokku/dokku/issues), assuming one
  does not already exist.
  * Clearly describe the issue including steps to reproduce when it is a bug.
  * Make sure you fill in the earliest version that you know has the issue.
* Fork the repository on GitHub.

### Making Changes

* Create a topic branch from where you want to base your work.
  * This is usually the master branch.
  * Only target an existing branch if you are certain your fix must be on that
    branch.
  * To quickly create a topic branch based on master; `git checkout -b my_contribution origin/master`.
    It is best to avoid working directly on the `master` branch. Doing so will
    help avoid conflicts if you pull in updates from origin.
* Make commits of logical units. Implementing a new function and calling it in
  another file constitute a single logical unit of work.
  * Before you make a pull request, squash your commits into logical units of work
    using `git rebase -i` and `git push -f`.
  * A majority of submissions should have a single commit, so if in doubt,
    squash your commits down to one commit.
* Check for unnecessary whitespace with `git diff --check` before committing.
* Use descriptive commit messages and reference the #issue number.
* Core test cases should continue to pass. You can run tests locally or enable
  [circle-ci](https://circleci.com/gh/dokku/dokku) for your fork, so all
  tests and codesniffs will be executed.
* Your work should apply the [Dokku coding standards](https://github.com/progrium/bashstyle)
* Pull requests must be cleanly rebased on top of master without multiple branches
  mixed into the PR.
  * **Git tip**: If your PR no longer merges cleanly, use `rebase master` in your
    feature branch to update your pull request rather than `merge master`.

### Which branch to base the work

All changes should be be based on the latest master commit.

### Submitting Changes

* Push your changes to a topic branch in your fork of the repository.
* Submit a pull request to the repository on github, with the correct target
  branch.

### When will my change be merged?

Be patient! The Dokku maintainers will review all pull requests and comment as
quickly as possible. There may be some back and forth while the details of your
pull request are discussed.

In the unlikely event that your pull request does not get merged, the Dokku
maintainers will either provide an alternative patch or guide you towards a
better solution to the problem at hand.

During our pre-1.0 cycle, we will follow these general rules when merging pull
requests:

- bugfix (patch)
- security (patch/minor)
- minor feature (patch)
- backwards incompatible change (minor)
- major feature (minor)

### Running tests locally

Please read the [testing docs](https://dokku.com/docs/development/testing/),
which contains test setup information as well as tips for running tests.

# Additional Resources

* [Dokku coding standards](https://github.com/progrium/bashstyle)
* [Existing issues](https://github.com/dokku/dokku/issues)
* [General GitHub documentation](https://help.github.com/)
* [GitHub pull request documentation](https://help.github.com/send-pull-requests/)
* [#dokku IRC channel on freenode.org](https://webchat.freenode.net/?channels=dokku)
