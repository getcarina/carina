# carina

CLI tool for Carina, the [Rackspace container service](https://mycluster.rackspacecloud.com) that's currently in Beta.

![](clustergalaxy.png)

There are downloads of the built binaries over on [releases](https://github.com/rackerlabs/carina/releases).

:warning: This is temporary tooling until we have integration into `rack` :warning:

```
$ carina --help-long
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
