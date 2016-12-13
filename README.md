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
1. In `$GOPATH/src/github.com/getcarina/carina/glide.lock` update the commit hash for libcarina. Make sure you are editing the libcarina package and not libmakeswarm. They are same repository, but represent different branches.
1. In `$GOPATH/src/github.com/getcarina/carina`, run `make` and verify that libcarina is restored the right vendored commit hash and everything still works.

## Releasing

### Beta Builds
Here's how to release a beta build:

1. Checkout the release branch: `git checkout release/v2.0.0`
1. Create a tag for the beta release: `git tag v2.0.0-beta.8 -a -m ""`
1. Push the tag: `git push --follow-tags`
1. Watch the Travis CI build, and wait for a successful deploy.
1. Validate the uploaded binary

    ```
    https://download.getcarina.com/carina/beta/$(uname -s)/$(uname -m)/carina -o carina
    chmod +x carina
    ./carina --version
    # Should print the new version
    ```
