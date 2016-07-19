Common make tasks
=================

Usage
-----

### Seting Up the common scripts

Add these common tasks to your project by using git subtree from the root of your project.

First add the remote.

```
git remote add common_makefiles git@github.com:pantheon-systems/common_makefiles.git --no-tags
```

Now add the subtree

**note:** it is important that you keep the import path set to `scripts/make` as the makefiles assume this structure.

```
git subtree add --prefix scripts/make common_makefiles master --squash
```

### Using in your Makefile

you simply need to include the common makefiles you want in your projects root Makefile

```
APP=baryon
PROJECT := $$GOOGLE_PROJECT

include scripts/make/common.mk
include scripts/make/common-kube.mk
include scripts/make/common-go.mk
```

### Extending Tasks

All the common makefile tasks can be extended in your top level Makefile by defining them again. Each common task that can be extended has a `::` target. e.g. `deps::`

for example if I want to do something after the default build target from common-go.mk I can add to it in my Makefile like so:

```
build::
  @echo "this is after the common build"
```

Updating the Common tasks
-------------------------

The `common.mk` file includes a task named `update-makefiles` which you can invoke to pull and squash the latest versions of the common tasks into your project.

```
make update-makefiles
```

Adding more tasks and common files
----------------------------------

make edits here and open a PR against this repo. Please do not push from your subtree on your project.

### Common Patterns for adding to the repo

Tasks should follow the form of `<action>-<context>-<object>` for example if I have a build task  and you want to add windows support you would add a `build-windows` or if you wanted to add a build for onebox you might do `build-onebox-linux` or simply `build-onebox`. 

There is the expectation that if you are doing a context specific task you add the context to your tasks. I.E. `test-circle`

For now these are the default/expected tasks for any common-LANG file:
* build
* deps
* test

This isn't written in stone, but I think it is a reasonable expectation that any engineer should be able to checkout any project and run:  `make deps && make build && make test` to get things running / testing.
