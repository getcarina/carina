# !!!! PRIVATE RELEASE !!!!
See the [GitHub releases](https://github.com/rackerlabs/carina-cli/releases) for the private binaries. You can't use curl easily to get them because the repo is private.

# Carina™ client
Create and interact with clusters on both Rackspace Public and Private Cloud.

![Carina Constellation](https://cloud.githubusercontent.com/assets/836375/10503963/e5bcca8c-72c0-11e5-8e14-2c1697297d7e.png)

See the [getting started tutorial](https://getcarina.com/docs/getting-started/getting-started-carina-cli/),
and [full documentation](https://getcarina.com/docs/reference/carina-cli/).

## Building

The build script assumes you're running go 1.5 or later. If not, upgrade or use
something like [gimme](https://github.com/travis-ci/gimme).

```bash
make
```

This creates `carina` in the current directory (there is no `make install` currently).

If you want to build without running validation or tests, use `make quick`.
