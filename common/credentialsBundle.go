package common

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const verifyCredentialsTimeout = 2 * time.Second

// CredentialsBundle is a set of certificates and environment information necessary to connect to a cluster
type CredentialsBundle struct {
	Files map[string][]byte
	Err   error
}

// NewCredentialsBundle initializes an empty credentials bundle
func NewCredentialsBundle() *CredentialsBundle {
	return &CredentialsBundle{
		Files: make(map[string][]byte),
	}
}

// LoadCredentialsBundle loads a credentials bundle from the filesystem
func LoadCredentialsBundle(credentialsPath string) CredentialsBundle {
	var creds CredentialsBundle

	files, err := ioutil.ReadDir(credentialsPath)
	if err != nil {
		creds.Err = errors.Wrapf(err, "Invalid credentials bundle. Cannot list files in %s", credentialsPath)
		return creds
	}

	creds.Files = make(map[string][]byte)
	for _, file := range files {
		filePath := filepath.Join(credentialsPath, file.Name())
		fileContents, err := ioutil.ReadFile(filePath)
		if err != nil {
			creds.Err = errors.Wrapf(err, "Invalid credentials bundle. Cannot read %s", filePath)
			return creds
		}
		creds.Files[file.Name()] = fileContents
	}

	return creds
}

// GetCA returns the contents of ca.pem
func (creds CredentialsBundle) GetCA() []byte {
	return creds.Files["ca.pem"]
}

// GetCert returns the contents of cert.pem
func (creds CredentialsBundle) GetCert() []byte {
	return creds.Files["cert.pem"]
}

// GetKey returns the contents of key.pem
func (creds CredentialsBundle) GetKey() []byte {
	return creds.Files["key.pem"]
}

// Verify validates that we can connect to the Docker host specified in the credentials bundle
func (creds CredentialsBundle) Verify() error {
	if creds.Err != nil {
		return creds.Err
	}

	Log.Debug("Verifying credentials bundle...")

	tlsConfig, err := creds.getTLSConfig()
	if err != nil {
		return err
	}

	host, err := creds.parseHost()
	if err != nil {
		return err
	}

	telephone := &net.Dialer{Timeout: verifyCredentialsTimeout}
	conn, err := tls.DialWithDialer(telephone, "tcp", host, tlsConfig)
	if err != nil {
		return errors.Wrapf(err, "Invalid credentials bundle. Unable to connect to %s.", host)
	}
	conn.Close()

	return nil
}

func (creds CredentialsBundle) parseHost() (string, error) {
	var host string
	var ok bool

	if config, isDocker := creds.Files["docker.env"]; isDocker {
		host, ok = parseHost(config, "DOCKER_HOST=")
		if !ok {
			return "", errors.New("Invalid credentials bundle. Could not parse DOCKER_HOST from docker.env.")
		}
	} else if config, isKubernetes := creds.Files["kubectl.config"]; isKubernetes {
		host, ok = parseHost(config, "server:")
		if !ok {
			return "", errors.New("Invalid credentials bundle. Could not parse server from kubectl.config.")
		}
	} else {
		return "", errors.New("Invalid credentials bundle. Missing both docker.env and kubectl.config.")
	}

	hostURL, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("Invalid credentials bundle. Bad host URL %s", host)
	}

	return hostURL.Host, nil
}

func parseHost(config []byte, token string) (string, bool) {
	lines := strings.Split(string(config), "\n")
	for _, line := range lines {
		host := strings.Split(line, token)
		if len(host) == 2 {
			return strings.TrimSpace(host[1]), true
		}
	}

	return "", false
}

func (creds CredentialsBundle) getTLSConfig() (*tls.Config, error) {
	var tlsConfig tls.Config
	tlsConfig.InsecureSkipVerify = true
	certPool := x509.NewCertPool()

	certPool.AppendCertsFromPEM(creds.GetCA())
	tlsConfig.RootCAs = certPool
	keypair, err := tls.X509KeyPair(creds.GetCert(), creds.GetKey())
	if err != nil {
		return &tlsConfig, errors.Wrap(err, "Invalid credentials bundle. Keypair mis-match.")
	}
	tlsConfig.Certificates = []tls.Certificate{keypair}
	return &tlsConfig, nil
}
