# Carina

CLI tool for Carina, the [Rackspace container service](https://mycluster.rackspacecloud.com) that's currently in Beta.

![Carina Constellation](https://cloud.githubusercontent.com/assets/836375/10503963/e5bcca8c-72c0-11e5-8e14-2c1697297d7e.png)

:warning: This is temporary tooling until we have integration into `rack` :warning:

## Installation

There are downloads of the built binaries over in [releases](https://github.com/rackerlabs/carina/releases).

After downloading the version for your system, you'll probably need to rename it,
set it as executable, and put it on a `PATH` you have:

### OS X

```
$ mv carina-darwin-amd64 ~/bin/carina
$ chmod u+x ~/bin/carina
```

### Linux

```
$ mv carina-linux-amd64 ~/bin/carina
$ chmod u+x ~/bin/carina
```

### Windows

TODO: Instructions for Windows. Care to add some?

## Usage

```
usage: carina [<flags>] <command> [<args> ...]

command line interface to launch and work with Docker Swarm clusters

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).
  --username="jupyter"
          Rackspace username - can also set env var RACKSPACE_USERNAME
  --api-key=RACKSPACE_APIKEY
          Rackspace API Key - can also set env var RACKSPACE_APIKEY
  --endpoint="https://mycluster.rackspacecloud.com"
          Carina API endpoint

Commands:
  help [<command>...]
    Show help.


  list
    list swarm clusters


  get <cluster-name>
    get information about a swarm cluster


  delete <cluster-name>
    delete a swarm cluster


  create [<flags>] <cluster-name>
    create a swarm cluster

    --wait       wait for swarm cluster completion
    --nodes=1    number of nodes for the initial cluster
    --autoscale  whether autoscale is on or off

  credentials [<flags>] <cluster-name>
    download credentials

    --path=PATH  path to write credentials out to

  grow --nodes=NODES <cluster-name>
    Grow a cluster by the requested number of nodes

    --nodes=NODES  number of nodes to increase the cluster by

```

## Building

The build script assumes you're running go 1.5 or later. If not, upgrade or use
something like [gimme](https://github.com/travis-ci/gimme).

```bash
make carina
```

This creates `carina` in the current directory (there is no `make install` currently).

If you want it to build on prior releases of go, we'd need a PR to change up how
the `Makefile` sets the `LDFLAGS` conditionally based on Go version.

## Releasing

### Prerequisites

The release script relies on [github-release](https://github.com/aktau/github-release). Get it, configure it.

Make sure you're on `master` then run `release.sh` with the next tag and release name.

```bash
./release.sh 0.2.0 "Acute Aquarius"
```

How do you pick the release name?

### Naming things

The hardest problem in computer science is picking names. For releases, we take
an adjective attached combined with the next constellation from an
[alphabetical list of constellations](http://www.astro.wisc.edu/~dolan/constellations/constellation_list.html).
It can be alliterative if you like.
