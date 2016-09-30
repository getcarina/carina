# !!!! PRIVATE RELEASE !!!!
See the [GitHub releases](https://github.com/rackerlabs/carina-cli/releases) for the private binaries. You can't use curl easily to get them because the repo is private.

# Carina™ client
Create and interact with clusters on both Rackspace Public and Private Cloud.

![Carina Constellation](https://cloud.githubusercontent.com/assets/836375/10503963/e5bcca8c-72c0-11e5-8e14-2c1697297d7e.png)

## Installation

To download and install the `carina` CLI, use the appropriate instructions for your operating system.

#### Linux and OS X

Downloads for the latest release of `carina` are available in [releases](https://github.com/rackerlabs/carina-cli/releases) for 64-bit Linux and OS X. Manually download the zip file, unzip, grab the right binary for your OS, move it to a directory on your `$PATH`, and make it executable:

```bash
$ mv v2.0.0-alpha.3/Linux/x86_64/carina ~/bin/carina
$ chmod u+x ~/bin/carina
```

#### Windows

Downloads for the latest release of `carina` are available in [releases](https://github.com/rackerlabs/carina-cli/releases). Manually download the zip file, unzip, and move the carina to a directory on your `$PATH`:

```plain
> mv v2.0.0-alpha.3\Windows\x86_64\carina.exe C:\some\place\on\the\windows\path\carina.exe
```

## Authentication
The user credentials are used to automatically detect the cloud with which the cli should communicate. First, it looks for the Rackspace Public Cloud environment variables, such as CARINA_USERNAME/CARINA_APIKEY or RS_USERNAME/RS_API_KEY. Then it looks for Rackspace Private Cloud environment variables, such as OS_USERNAME/OS_PASSWORD. Use --cloud flag to explicitly select a cloud.

In the following example, the detected cloud is 'private' because --password is specified:
```carina --username bob --password ilovepuppies --project admin --auth-endpoint http://example.com/auth/v3 ls```

In the following example, the detected cloud is 'public' because --apikey is specified:
```carina --username bob --apikey abc123 ls```

In the following example, 'private' is used, even though the Rackspace Public Cloud environment variables may be present, because the --cloud is specified:
```carina --cloud private ls```

### Profiles
Credentials can be saved under a profile name in ~/.carina/config then used with the --profile flag. If --profile is not specified, and the config file contains a profile named 'default', it will be used when no other credential flags are provided. Use `--no-profile` to disable profiles.

The configuration file is in [TOML](https://github.com/toml-lang/toml) syntax.

Below is a sample config file:

```toml
# The following profile stores its credentials in plain text.
[prod]
cloud="public"
username="alicia"
api-key="abc123"

# The following profile retrieves its credentials from environment variables defined in your openrc file.
[dev]
cloud="private"
username-var="OS_USERNAME"
password-var="OS_PASSWORD"
auth-endpoint-var="OS_AUTH_URL"
tenant-var="OS_TENANT_NAME"
project-var="OS_PROJECT_NAME"
domain-var="OS_PROJECT_DOMAIN_NAME"

# The following profile is used when no --profile is specified.
# The default profile takes precedence over auto-discovered environment variables
[default]
cloud="public"
username-var="RS_USERNAME"
apikey-var="RS_API_KEY"
```

In the following example, the default profile is used to authenticate because no other credentials were explicitly provided:
```carina ls```

In the following example, the dev profile is used to authenticate:
```carina --profile dev ls```

## Getting started

```
$ carina list
ID		Name		Status	Type		Nodes
abc123	mycluster	active	kubernetes	1

$ carina create --template swarm-dev newone
ID		Name		Status	Type		Nodes
def456	newone		active	swarm		1

$ carina create --template kubernetes-dev --nodes 3 --wait another
ID		Name		Status	Type		Nodes
geh978	another		active	kubernetes	3

$ carina list
ID		Name		Status	Type		Nodes
abc123	mycluster	active	kubernetes	1
def456	newone		active	swarm		1
geh978	another		active	kubernetes	3

$ carina credentials mycluster
#
# Credentials written to "~/.carina/clusters/public-alicia/mycluster"
# To see how to connect to your cluster, run: carina env mycluster
#

$ eval "$( carina env mycluster )"
$ kubectl cluster-info

$ eval "$( carina env newone )"
$ docker info
```


## Usage

```
Create and interact with clusters on both Rackspace Public and Private Cloud

Usage:
  carina [command]

Available Commands:
  create          Create a cluster
  credentials     Download a cluster's credentials
  delete          Delete a cluster
  env             Show the command to load a cluster's credentials
  get             Show information about a cluster
  grow            Add nodes to a cluster
  list            List clusters
  quotas          Show the user's quotas
  rebuild         Rebuild a cluster
  version         Show the application version

Flags:
      --api-key string         Public Cloud API Key [CARINA_APIKEY/RS_API_KEY]
      --auth-endpoint string   Private Cloud Authentication endpoint [OS_AUTH_URL]
      --cache                  Cache API tokens and update times (default true)
      --cloud string           The cloud type: public or private
      --config string          config file (default is CARINA_HOME/config.toml)
      --debug                  Print additional debug messages to stdout
      --domain string          Private Cloud Domain Name [OS_DOMAIN_NAME]
  -h, --help                   help for carina
      --no-profile             Ignore profiles and use flags and/or environment variables only
      --password string        Private Cloud Password [OS_PASSWORD]
      --profile string         Use saved credentials for the specified profile
      --project string         Private Cloud Project Name [OS_PROJECT_NAME]
      --region string          Private Cloud Region Name [OS_REGION_NAME]
      --silent                 Do not print to stdout
      --username string        Username [CARINA_USERNAME/RS_USERNAME/OS_USERNAME]

Environment Variables:
  CARINA_HOME
    directory that stores your cluster tokens and credentials
    current setting: ~/.carina
```

## Building

The build script assumes you're running go 1.5 or later. If not, upgrade or use
something like [gimme](https://github.com/travis-ci/gimme).

```bash
make
```

This creates `carina` in the current directory (there is no `make install` currently).

If you want to build without running validation or tests, use `make quick`.
