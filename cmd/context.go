package cmd

import (
	"errors"
	"fmt"

	"os"

	"github.com/getcarina/carina/client"
	"github.com/getcarina/carina/common"
	"github.com/getcarina/carina/magnum"
	"github.com/getcarina/carina/make-coe"
	"github.com/getcarina/carina/makeswarm"
)

// CarinaUserNameEnvVar is the Carina username environment variable (1st)
const CarinaUserNameEnvVar = "CARINA_USERNAME"

// RackspaceUserNameEnvVar is the Rackspace username environment variable (2nd)
const RackspaceUserNameEnvVar = "RS_USERNAME"

// OpenStackUserNameEnvVar is the OpenStack username environment variable (3nd)
const OpenStackUserNameEnvVar = "OS_USERNAME"

// CarinaAPIKeyEnvVar is the Carina API key environment variable (1st)
const CarinaAPIKeyEnvVar = "CARINA_APIKEY"

// RackspaceAPIKeyEnvVar is the Rackspace API key environment variable (2nd)
const RackspaceAPIKeyEnvVar = "RS_API_KEY"

// OpenStackPasswordEnvVar is OpenStack password environment variable
// When set, this instructs the cli to communicate with Carina (private cloud) instead of the default Carina (public cloud)
const OpenStackPasswordEnvVar = "OS_PASSWORD"

// OpenStackAuthURLEnvVar is the OpenStack Identity URL (v2 and v3 supported)
const OpenStackAuthURLEnvVar = "OS_AUTH_URL"

// CarinaEndpointEnvVar overrides the default Carina endpoint
const CarinaEndpointEnvVar = "CARINA_ENDPOINT"

// OpenStackEndpointEnvVar overrides the default endpoint from the service catalog
const OpenStackEndpointEnvVar = "OS_ENDPOINT"

// OpenStackProjectEnvVar is the OpenStack project name, required for identity v3
const OpenStackProjectEnvVar = "OS_PROJECT_NAME"

// OpenStackDomainEnvVar is the OpenStack domain name, optional for identity v3
const OpenStackDomainEnvVar = "OS_DOMAIN_NAME"

// OpenStackRegionEnvVar is the OpenStack domain name, optional for identity v3
const OpenStackRegionEnvVar = "OS_REGION_NAME"

type context struct {
	// Values built from flags
	Client  *client.Client
	Account client.Account

	// Global Flags
	CacheEnabled bool
	ConfigFile   string
	Debug        bool
	Silent       bool

	// Account Flags
	CloudType    string
	Username     string
	APIKey       string
	Password     string
	Project      string
	Domain       string
	Region       string
	AuthEndpoint string
	Endpoint     string
}

func (cxt *context) buildAccount() client.Account {
	switch cxt.CloudType {
	case client.CloudMakeCOE:
		return &makecoe.Account{
			Endpoint: cxt.Endpoint,
			UserName: cxt.Username,
			APIKey:   cxt.APIKey,
		}
	case client.CloudMakeSwarm:
		return &makeswarm.Account{
			Endpoint: cxt.Endpoint,
			UserName: cxt.Username,
			APIKey:   cxt.APIKey,
		}
	case client.CloudMagnum:
		return &magnum.Account{
			AuthEndpoint: cxt.AuthEndpoint,
			Endpoint:     cxt.Endpoint,
			UserName:     cxt.Username,
			Password:     cxt.Password,
			Project:      cxt.Project,
			Domain:       cxt.Domain,
		}
	default:
		panic(fmt.Sprintf("Unsupported cloud type: %s", cxt.CloudType))
	}
}

func (cxt *context) initialize() error {
	if cxt.Silent {
		common.Log.SetSilent()
	} else if cxt.Debug {
		common.Log.SetDebug()
	}

	// Verify that we have enough information to identity the account: apikey or password
	apikeyFound := cxt.APIKey != "" || os.Getenv(CarinaAPIKeyEnvVar) != "" || os.Getenv(RackspaceAPIKeyEnvVar) != ""
	passwordFound := cxt.Password != "" || os.Getenv(OpenStackPasswordEnvVar) != ""
	if !apikeyFound && !passwordFound {
		return errors.New("No credentials provided. An --apikey or --password must either be specified or the equivalent environment variables must be set. Run carina --help for more information.")
	}

	// Detect the cloud provider
	switch cxt.CloudType {
	case client.CloudMakeCOE, client.CloudMagnum, client.CloudMakeSwarm:
		break
	case "":
		common.Log.WriteDebug("No cloud type specified, detecting with the provided credentials. Use --cloud to skip detection.")
		if apikeyFound {
			cxt.CloudType = client.CloudMakeCOE
			common.Log.WriteDebug("Cloud: public")
		} else {
			cxt.CloudType = client.CloudMagnum
			common.Log.WriteDebug("Cloud: private")
		}
	default:
		return fmt.Errorf("Invalid --cloud value: %s. Allowed values are public, private and make-swarm", cxt.CloudType)
	}

	// Initialize the remaining flags based on the cloud provider
	var err error
	switch cxt.CloudType {
	case client.CloudMakeSwarm, client.CloudMakeCOE:
		err = initCarinaFlags(cxt)
	case client.CloudMagnum:
		err = initMagnumFlags(cxt)
	}
	if err != nil {
		return err
	}

	cxt.Client = client.NewClient(cxt.CacheEnabled)
	cxt.Account = cxt.buildAccount()
	return nil
}

func initCarinaFlags(cxt *context) error {
	// auth-endpoint = --auth-endpoint -> rackspace identity endpoint
	if cxt.AuthEndpoint == "" {
		common.Log.WriteDebug("AuthEndpoint: default")
	} else {
		common.Log.WriteDebug("AuthEndpoint: --auth-endpoint")
	}

	// endpoint = --endpoint -> public carina endpoint
	if cxt.Endpoint == "" {
		cxt.Endpoint = os.Getenv(CarinaEndpointEnvVar)
		if cxt.Endpoint == "" {
			common.Log.WriteDebug("Endpoint: default")
		} else {
			common.Log.WriteDebug("Endpoint: %s", CarinaEndpointEnvVar)
		}
	} else {
		common.Log.WriteDebug("Endpoint: --endpoint")
	}

	// username = --username -> CARINA_USERNAME -> RS_USERNAME
	if cxt.Username == "" {
		cxt.Username = os.Getenv(CarinaUserNameEnvVar)
		if cxt.Username == "" {
			cxt.Username = os.Getenv(RackspaceUserNameEnvVar)
			if cxt.Username == "" {
				return fmt.Errorf("UserName was not specified. Either use --username or set %s, or %s.", CarinaUserNameEnvVar, RackspaceUserNameEnvVar)
			}
			common.Log.WriteDebug("UserName: %s", RackspaceUserNameEnvVar)
		} else {
			common.Log.WriteDebug("UserName: %s", CarinaUserNameEnvVar)
		}
	} else {
		common.Log.WriteDebug("UserName: --username")
	}

	// api-key = --api-key -> CARINA_APIKEY -> RS_API_KEY
	if cxt.APIKey == "" {
		cxt.APIKey = os.Getenv(CarinaAPIKeyEnvVar)
		if cxt.APIKey == "" {
			cxt.APIKey = os.Getenv(RackspaceAPIKeyEnvVar)
			if cxt.APIKey == "" {
				return fmt.Errorf("API Key was not specified. Either use --api-key or set %s or %s", CarinaAPIKeyEnvVar, RackspaceAPIKeyEnvVar)
			}
			common.Log.WriteDebug("API Key: %s", RackspaceAPIKeyEnvVar)
		} else {
			common.Log.WriteDebug("API Key: %s", CarinaAPIKeyEnvVar)
		}
	} else {
		common.Log.WriteDebug("API Key: --api-key")
	}

	return nil
}

func initMagnumFlags(cxt *context) error {
	// auth-endpoint = --auth-endpoint -> OS_AUTH_URL
	if cxt.AuthEndpoint == "" {
		cxt.AuthEndpoint = os.Getenv(OpenStackAuthURLEnvVar)
		if cxt.AuthEndpoint == "" {
			return fmt.Errorf("AuthEndpoint was not specified via --auth-endpoint or %s", OpenStackAuthURLEnvVar)
		}
		common.Log.WriteDebug("AuthEndpoint: %s", OpenStackAuthURLEnvVar)
	} else {
		common.Log.WriteDebug("AuthEndpoint: --auth-endpoint")
	}

	// endpoint = --endpoint -> OS_ENDPOINT -> service catalog endpoint
	if cxt.Endpoint == "" {
		cxt.Endpoint = os.Getenv(OpenStackEndpointEnvVar)
		if cxt.Endpoint == "" {
			common.Log.WriteDebug("Endpoint: default")
		} else {
			common.Log.WriteDebug("Endpoint: %s", OpenStackEndpointEnvVar)
		}
	} else {
		common.Log.WriteDebug("Endpoint: --endpoint")
	}

	// username = --username -> OS_USERNAME
	if cxt.Username == "" {
		cxt.Username = os.Getenv(OpenStackUserNameEnvVar)
		if cxt.Username == "" {
			return fmt.Errorf("UserName was not specified via --username or %s", OpenStackUserNameEnvVar)
		}
		common.Log.WriteDebug("UserName: %s", OpenStackUserNameEnvVar)
	} else {
		common.Log.WriteDebug("UserName: --username")
	}

	// password = --password -> OS_PASSWORD
	if cxt.Password == "" {
		cxt.Password = os.Getenv(OpenStackPasswordEnvVar)
		if cxt.Password == "" {
			return fmt.Errorf("Password was not specified via --password or %s", OpenStackPasswordEnvVar)
		}
		common.Log.WriteDebug("Password: %s", OpenStackPasswordEnvVar)
	} else {
		common.Log.WriteDebug("Password: --password")
	}

	// project = --project -> OS_PROJECT_NAME
	if cxt.Project == "" {
		cxt.Project = os.Getenv(OpenStackProjectEnvVar)
		if cxt.Project == "" {
			common.Log.WriteDebug("Project was not specified. Either use --project or set %s.", OpenStackProjectEnvVar)
		} else {
			common.Log.WriteDebug("Project: %s", OpenStackProjectEnvVar)
		}
	} else {
		common.Log.WriteDebug("Project: --project")
	}

	// domain = --domain -> OS_DOMAIN_NAME -> "default"
	if cxt.Domain == "" {
		cxt.Domain = os.Getenv(OpenStackDomainEnvVar)
		if cxt.Domain == "" {
			cxt.Domain = "default"
			common.Log.WriteDebug("Domain: default. Either use --domain or set %s.", OpenStackDomainEnvVar)
		} else {
			common.Log.WriteDebug("Domain: %s", OpenStackDomainEnvVar)
		}
	} else {
		common.Log.WriteDebug("Domain: --domain")
	}

	// region = --region -> OS_REGION_NAME -> "RegionOne"
	if cxt.Region == "" {
		cxt.Region = os.Getenv(OpenStackRegionEnvVar)
		if cxt.Region == "" {
			cxt.Region = "RegionOne"
			common.Log.WriteDebug("Region: RegionOne. Either use --region or set %s.", OpenStackRegionEnvVar)
		} else {
			common.Log.WriteDebug("Region: %s", OpenStackRegionEnvVar)
		}
	} else {
		common.Log.WriteDebug("Region: --region")
	}

	return nil
}
