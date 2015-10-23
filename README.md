# Carina™ client

[![Travis Build Status](https://travis-ci.org/getcarina/carina.svg)](https://travis-ci.org/getcarina/carina)
[![Circle CI Build Status](https://circleci.com/gh/getcarina/carina.svg?style=svg)](https://circleci.com/gh/getcarina/carina)
[![Appveyor Build Status](https://ci.appveyor.com/api/projects/status/8qjckvi0cvfgf1qr?svg=true)](https://ci.appveyor.com/project/rgbkrk/carina)

Command line client for [Carina by Rackspace](https://getcarina.com), a container service that's currently in Beta.

![Carina Constellation](https://cloud.githubusercontent.com/assets/836375/10503963/e5bcca8c-72c0-11e5-8e14-2c1697297d7e.png)

## Installation

There are downloads of the built binaries over in [releases](https://github.com/getcarina/carina/releases).

After downloading the version for your system, you'll probably need to rename it,
set it as executable, and put it on a `PATH` you have:

### OS X

```bash
$ mv carina-darwin-amd64 ~/bin/carina
$ chmod u+x ~/bin/carina
```

### Linux

```bash
$ mv carina-linux-amd64 ~/bin/carina
$ chmod u+x ~/bin/carina
```

### Windows

```
> carina.exe
```

## Getting started

```
$ export CARINA_USERNAME=trythingsout
$ export CARINA_APIKEY=$RACKSPACE_APIKEY
$ carina list
ClusterName Flavor        Nodes AutoScale Status
mycluster   container1-4G 1     false     active
$ carina create newone
newone      container1-4G 1     false     new
$ carina create another --wait --autoscale
another     container1-4G 1     true      active
$ carina list
ClusterName Flavor        Nodes AutoScale Status
mycluster   container1-4G 1     false     active
newone      container1-4G 1     false     active
another     container1-4G 1     true      active
$ carina credentials another
#
# Credentials written to "/Users/rgbkrk/.carina/clusters/trythingsout/another"
#
source "/Users/rgbkrk/.carina/clusters/trythingsout/another/docker.env"
# Run the command above to get your Docker environment variables set

$ eval "$( carina credentials another )"
$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
$ docker run -d --name whoa -p 8080:8080 whoa/tiny
0e857826144194fd089310279915b1a052de9fb878d6d4f61420a0c64ee06c53
$ curl $( docker port whoa 8080 )
👊  I know kung fu  👊
```


## Usage

```
usage: carina [<flags>] <command> [<args> ...]

command line interface to launch and work with Docker Swarm clusters

Flags:
  --help               Show context-sensitive help (also try --help-long and --help-man).
  --version            Show application version.
  --username=USERNAME  Carina username - can also set env var CARINA_USERNAME
  --api-key=CARINA_APIKEY
                       Carina API Key - can also set env var CARINA_APIKEY
  --endpoint="https://mycluster.rackspacecloud.com"
                       Carina API endpoint

Commands:
  help [<command>...]
    Show help.

  create [<flags>] <cluster-name>
    Create a swarm cluster

    --wait       wait for swarm cluster to come online (or error)
    --nodes=1    number of nodes for the initial cluster
    --autoscale  whether autoscale is on or off

  get <cluster-name>
    Get information about a swarm cluster

  list
    List swarm clusters

  grow --nodes=NODES <cluster-name>
    Grow a cluster by the requested number of nodes

    --nodes=NODES  number of nodes to increase the cluster by

  credentials [<flags>] <cluster-name>
    download credentials

    --path=<cluster-name>
      path to write credentials out to

  rebuild [<flags>] <cluster-name>
    Rebuild a swarm cluster

    --wait  wait for swarm cluster to come online (or error)

  delete <cluster-name>
    Delete a swarm cluster
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
