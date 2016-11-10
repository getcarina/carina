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
	"github.com/spf13/viper"
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
const OpenStackPasswordEnvVar = "OS_PASSWORD"

// RackspaceAuthURLEnvVar is the Rackspace Identity URL
const RackspaceAuthURLEnvVar = "RS_AUTH_URL"

// OpenStackAuthURLEnvVar is the OpenStack Identity URL (v2 and v3 supported)
const OpenStackAuthURLEnvVar = "OS_AUTH_URL"

// CarinaEndpointEnvVar overrides the default Carina endpoint
const CarinaEndpointEnvVar = "CARINA_ENDPOINT"

// OpenStackEndpointEnvVar overrides the default endpoint from the service catalog
const OpenStackEndpointEnvVar = "OS_ENDPOINT"

// OpenStackProjectEnvVar is the OpenStack project name, required for identity v3
const OpenStackProjectEnvVar = "OS_PROJECT_NAME"

// OpenStackProjectDomainEnvVar is the OpenStack _project_ domain name, optional for identity v3
const OpenStackProjectDomainEnvVar = "OS_PROJECT_DOMAIN_NAME"

// OpenStackUserDomainEnvVar is the OpenStack _user_ domain name, optional for identity v3
const OpenStackUserDomainEnvVar = "OS_USER_DOMAIN_NAME"

// OpenStackDomainEnvVar is the OpenStack domain name, optional for identity v3
const OpenStackDomainEnvVar = "OS_DOMAIN_NAME"

// CarinaRegionEnvVar is the Carina region name
const CarinaRegionEnvVar = "CARINA_REGION"

// RackspaceRegionEnvVar is the Rackspace region name
const RackspaceRegionEnvVar = "RS_REGION_NAME"

// OpenStackRegionEnvVar is the OpenStack region name
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
	Profile          string
	ProfileDisabled  bool
	CloudType        string
	Username         string
	APIKey           string
	Password         string
	Project          string
	Domain           string
	Region           string
	AuthEndpoint     string
	Endpoint        string
}

func (cxt *context) shouldTryProfile() bool {
	if cxt.ProfileDisabled {
		return false
	}

	if cxt.userSpecifiedAuthFlagsExist() {
		return false
	}

	configFile := viper.ConfigFileUsed()
	return configFile != ""
}

func (cxt *context) userSpecifiedAuthFlagsExist() bool {
	return cxt.CloudType != "" ||
		cxt.Username != "" ||
		cxt.Password != "" ||
		cxt.APIKey != "" ||
		cxt.Domain != "" ||
		cxt.Project != "" ||
		cxt.Region != ""
}

func (cxt *context) buildAccount() client.Account {
	switch cxt.CloudType {
	case client.CloudMakeCOE:
		return &makecoe.Account{
			Endpoint: cxt.Endpoint,
			UserName:         cxt.Username,
			APIKey:           cxt.APIKey,
			Region:           cxt.Region,
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

	var profileLoaded bool
	var err error
	if cxt.shouldTryProfile() {
		profileLoaded, err = cxt.loadProfile()
		if err != nil {
			return err
		}
	}

	// Build-up to the authentication information from flags and environment variables
	if !profileLoaded {
		// Detect the cloud provider
		err := cxt.detectCloud()
		if err != nil {
			return err
		}

		// Initialize the remaining flags based on the cloud provider
		switch cxt.CloudType {
		case client.CloudMakeSwarm, client.CloudMakeCOE:
			err = cxt.initCarinaFlags()
		case client.CloudMagnum:
			err = cxt.initMagnumFlags()
		}
		if err != nil {
			return err
		}
	}

	cxt.Client = client.NewClient(cxt.CacheEnabled)
	cxt.Account = cxt.buildAccount()

	return nil
}

func (cxt *context) loadProfile() (ok bool, err error) {
	configFile := viper.ConfigFileUsed()

	// Try to use the default profile
	if cxt.Profile == "" && viper.InConfig("default") {
		cxt.Profile = "default"
		return cxt.loadProfile()
	}

	profile := viper.GetStringMapString(cxt.Profile)
	if len(profile) == 0 {
		return false, fmt.Errorf("Profile, %s, not found in %s", cxt.Profile, configFile)
	}
	common.Log.WriteDebug("Reading %s profile from %s", cxt.Profile, configFile)

	cxt.CloudType = profile["cloud"]
	switch cxt.CloudType {
	case client.CloudMakeSwarm, client.CloudMakeCOE:
		err = cxt.loadCarinaProfile(profile)
	case client.CloudMagnum:
		err = cxt.loadMagnumProfile(profile)
	case "":
		err = fmt.Errorf("Invalid profile: cloud is missing")
	default:
		err = fmt.Errorf("Invalid profile: %s is not a valid cloud type", cxt.CloudType)
	}

	return err == nil, err
}

func (cxt *context) detectCloud() error {
	// Verify that we have enough information: apikey or password
	apikeyFound := cxt.APIKey != "" || os.Getenv(CarinaAPIKeyEnvVar) != "" || os.Getenv(RackspaceAPIKeyEnvVar) != ""
	passwordFound := cxt.Password != "" || os.Getenv(OpenStackPasswordEnvVar) != ""
	if !apikeyFound && !passwordFound {
		return errors.New("No credentials provided. A --profile, --apikey or --password must be specified or the equivalent environment variables set. Run carina --help for more information.")
	}

	switch cxt.CloudType {
	case client.CloudMakeCOE, client.CloudMagnum, client.CloudMakeSwarm:
		break
	case "":
		common.Log.WriteDebug("No cloud type specified, detecting with the provided credentials. Use --cloud or --profile to skip detection.")
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

	return nil
}

func (cxt *context) initCarinaFlags() error {
	// auth-endpoint = --auth-endpoint -> RS_AUTH_URL -> rackspace identity endpoint
	if cxt.AuthEndpoint == "" {
		cxt.AuthEndpoint = os.Getenv(RackspaceAuthURLEnvVar)
		if cxt.AuthEndpoint == "" {
			common.Log.WriteDebug("AuthEndpoint: %s", RackspaceAuthURLEnvVar)
		} else {
			common.Log.WriteDebug("AuthEndpoint: default")
		}
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

	// apikey = --apikey -> CARINA_APIKEY -> RS_API_KEY
	if cxt.APIKey == "" {
		cxt.APIKey = os.Getenv(CarinaAPIKeyEnvVar)
		if cxt.APIKey == "" {
			cxt.APIKey = os.Getenv(RackspaceAPIKeyEnvVar)
			if cxt.APIKey == "" {
				return fmt.Errorf("API Key was not specified. Either use --apikey or set %s or %s", CarinaAPIKeyEnvVar, RackspaceAPIKeyEnvVar)
			}
			common.Log.WriteDebug("API Key: %s", RackspaceAPIKeyEnvVar)
		} else {
			common.Log.WriteDebug("API Key: %s", CarinaAPIKeyEnvVar)
		}
	} else {
		common.Log.WriteDebug("API Key: --apikey")
	}

	// region = --region -> CARINA_REGION -> RS_REGION_NAME
	if cxt.Region == "" {
		cxt.Region = os.Getenv(CarinaRegionEnvVar)
		if cxt.Region == "" {
			cxt.Region = os.Getenv(RackspaceRegionEnvVar)
			if cxt.Region == "" {
				common.Log.WriteDebug("Region: not specified. Either use --region, or set %s or %s.", CarinaRegionEnvVar, RackspaceRegionEnvVar)
			} else {
				common.Log.WriteDebug("Region: %s", RackspaceRegionEnvVar)
			}
		} else {
			common.Log.WriteDebug("Region: %s", CarinaRegionEnvVar)
		}
	} else {
		common.Log.WriteDebug("Region: --region")
	}

	return nil
}

func (cxt *context) initMagnumFlags() error {
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

	// domain = --domain -> OS_PROJECT_DOMAIN_NAME -> OS_USER_DOMAIN_NAME -> OS_DOMAIN_NAME -> "default"
	if cxt.Domain == "" {
		domainVar := OpenStackProjectDomainEnvVar
		cxt.Domain = os.Getenv(OpenStackProjectDomainEnvVar)
		if cxt.Domain == "" {
			domainVar = OpenStackUserDomainEnvVar
			cxt.Domain = os.Getenv(OpenStackUserDomainEnvVar)
		}
		if cxt.Domain == "" {
			domainVar = OpenStackDomainEnvVar
			cxt.Domain = os.Getenv(OpenStackDomainEnvVar)
		}

		if cxt.Domain == "" {
			cxt.Domain = "default"
			common.Log.WriteDebug("Domain: default. Either use --domain or set %s/%s/%s.", OpenStackProjectDomainEnvVar, OpenStackUserDomainEnvVar, OpenStackDomainEnvVar)
		} else {
			common.Log.WriteDebug("Domain: %s", domainVar)
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

func (cxt *context) loadCarinaProfile(profile map[string]string) (err error) {
	cxt.AuthEndpoint, err = cxt.getProfileSetting(profile, "auth-endpoint", "", false)
	if err != nil {
		return err
	}

	cxt.Endpoint, err = cxt.getProfileSetting(profile, "endpoint", "", false)
	if err != nil {
		return err
	}

	cxt.Username, err = cxt.getProfileSetting(profile, "username", "", true)
	if err != nil {
		return err
	}

	cxt.APIKey, err = cxt.getProfileSetting(profile, "apikey", "", true)
	if err != nil {
		return err
	}

	cxt.Region, err = cxt.getProfileSetting(profile, "region", "", false)
	if err != nil {
		return err
	}

	return nil
}

func (cxt *context) loadMagnumProfile(profile map[string]string) (err error) {
	cxt.AuthEndpoint, err = cxt.getProfileSetting(profile, "auth-endpoint", "", true)
	if err != nil {
		return err
	}

	cxt.Endpoint, err = cxt.getProfileSetting(profile, "endpoint", "", false)
	if err != nil {
		return err
	}

	cxt.Username, err = cxt.getProfileSetting(profile, "username", "", true)
	if err != nil {
		return err
	}

	cxt.Password, err = cxt.getProfileSetting(profile, "password", "", true)
	if err != nil {
		return err
	}

	cxt.Project, err = cxt.getProfileSetting(profile, "project", "", true)
	if err != nil {
		return err
	}

	cxt.Domain, err = cxt.getProfileSetting(profile, "domain", "default", false)
	if err != nil {
		return err
	}

	cxt.Region, err = cxt.getProfileSetting(profile, "region", "RegionOne", false)
	if err != nil {
		return err
	}

	return nil
}

func (cxt *context) getProfileSetting(profile map[string]string, key string, defaultValue string, required bool) (string, error) {
	envVar := profile[key+"-var"]
	value := profile[key]

	if envVar != "" {
		value = os.Getenv(envVar)
		common.Log.WriteSetting(key, envVar, value)
	} else if value != "" {
		common.Log.WriteSetting(key, "profile", value)
	} else if defaultValue != "" {
		value = defaultValue
		common.Log.WriteSetting(key, "using default value", value)
	}

	if required && value == "" {
		return "", fmt.Errorf("Invalid Profile: %s is missing", key)
	}

	return value, nil
}
