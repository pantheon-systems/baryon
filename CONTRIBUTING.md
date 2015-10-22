Contribute
==========

Creating Issues
---------------

Search existing open Issues and Pull Requests, and add your debug info before creating a new issue.

Setting Up
----------

1. Make sure you have Go 1.5 or greater installed
2. Clone this Git repository on your local machine.
3. Run `go get github.com/pantheon-systems/baryon` to pull the source
4. Run `cd $GOPATH/src/github.com/pantheon-systems/baryon && make` to pull deps and build

Submitting Patches
------------------

Whether you want to fix a bug or implement a new feature, the process is pretty much the same:

0. [Search existing issues](https://github.com/pantheon-systems/baryon/issues); if you can't find anything related to what you want to work on, open a new issue so that you can get some initial feedback.
1. [Fork](https://github.com/pantheon-systems/baryon/fork) the repository.
2. Push the code changes from your local clone to your fork.
3. Open a pull request.

It doesn't matter if the code isn't perfect. The idea is to get it reviewed early and iterate on it.

If you're adding a new feature, please add unit tests for it.

### Versions

In keeping with the standards of semantic versioning, backward-incompatible fixes are targeted to "major" versions. "Minor" versions are reserved for significant feature/bug releases needed between major versions. "Patch" releases are reserved only for critical security issues and other bugs critical to stabilizing the release.

### Release Stability

If you are using baryon in a production environment, you should be deploying the executable for [the latest release](https://github.com/pantheon-systems/baryon/releases).

Feedback
--------

Writing for Baryon should be fun. If you find any of this hard to figure out, let us know so we can improve our process or documentation!
