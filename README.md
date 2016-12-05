# Carina™ client
Create and interact with clusters on both Rackspace Public and Private Cloud.

![Carina Constellation](https://cloud.githubusercontent.com/assets/836375/10503963/e5bcca8c-72c0-11e5-8e14-2c1697297d7e.png)

See the [getting started tutorial](https://getcarina.com/docs/getting-started/getting-started-carina-cli/),
and [full documentation](https://getcarina.com/docs/reference/carina-cli/).

## Building

The build script assumes you're running go 1.6 or later. If not, upgrade or use
something like [gimme](https://github.com/travis-ci/gimme).

```bash
make
```

This creates `carina` in the current directory (there is no `make install` currently).

**Make Targets**

* `make`: First run for newcomers.
* `make get-deps`: Retrieves dependencies into the `vendor/` directory using glide.
* `make validate`: Run tools like `fmt`.
* `make test`: Run unit tests.
* `make local`: Build for the current dev env, using whatever dependencies that happen to be on the local machine.
* `make cross-build`: The official build.

### Editing libcarina
Here's how to work on libcarina and the cli at the same time locally:

1. Run `go get github.com/getcarina/libcarina`.
1. Make required changes to libcarina in `$GOPATH/src/github.com/getcarina/libcarina`.
1. In `$GOPATH/src/github.com/getcarina/carina`, run `rm $GOPATH/src/github.com/getcarina/carina/vendor/github.com/getcarina/libcarina` so that Go will pickup your local edits, and not use the vendored version. Use `make local` to build. Don't use `make` as it will restore the vendored copy of libcarina, overriding your local changes.
1. When everything is looking good, run `make` in `$GOPATH/src/github.com/getcarina/libcarina` to validate and format your changes.
1. Submit a PR to libcarina and once it is merged to master, note the commit hash.
4. In `$GOPATH/src/github.com/getcarina/carina/glide.lock` update the commit hash for libcarina. Make sure you are editing the libcarina package and not libmakeswarm. They are same repository, but represent different branches.
5. In `$GOPATH/src/github.com/getcarina/carina`, run `make` and verify that libcarina is restored the right vendored commit hash and everything still works.