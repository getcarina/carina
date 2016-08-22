package common

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// CredentialsBundle is a set of certificates and environment information necessary to connect to a cluster
type CredentialsBundle struct {
	Files map[string][]byte
}

// NewCredentialsBundle initializes an empty credentials bundle
func NewCredentialsBundle() *CredentialsBundle {
	return &CredentialsBundle{
		Files: make(map[string][]byte),
	}
}

// LoadCredentialsBundle loads a credentials bundle from the filesystem
func LoadCredentialsBundle(credentialsPath string) (CredentialsBundle, error) {
	var creds CredentialsBundle

	files, err := ioutil.ReadDir(credentialsPath)
	if err != nil {
		return creds, errors.Wrap(err, "Invalid credentials bundle. Cannot list files in "+credentialsPath)
	}

	creds.Files = make(map[string][]byte)
	for _, file := range files {
		filePath := filepath.Join(credentialsPath, file.Name())
		fileContents, err := ioutil.ReadFile(filePath)
		if err != nil {
			return creds, errors.Wrap(err, "Invalid credentials bundle. Cannot read "+filePath)
		}
		creds.Files[file.Name()] = fileContents
	}

	return creds, nil
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

// GetDockerEnv returns the contents of docker.env
func (creds CredentialsBundle) GetDockerEnv() []byte {
	return creds.Files["docker.env"]
}

// Verify validates that we can connect to the Docker host specified in the credentials bundle
func (creds CredentialsBundle) Verify() error {
	tlsConfig, err := creds.getTLSConfig()
	if err != nil {
		return err
	}

	// Lookup the Docker host from docker.env
	dockerEnv := string(creds.GetDockerEnv()[:])
	var dockerHost string
	sourceLines := strings.Split(dockerEnv, "\n")
	for _, line := range sourceLines {
		if strings.Index(line, "export ") == 0 {
			varDecl := strings.TrimRight(line[7:], "\n")
			eqLocation := strings.Index(varDecl, "=")

			varName := varDecl[:eqLocation]
			varValue := varDecl[eqLocation+1:]

			switch varName {
			case "DOCKER_HOST":
				dockerHost = varValue
			}

		}
	}

	dockerHostURL, err := url.Parse(dockerHost)
	if err != nil {
		return errors.Wrap(err, "Invalid credentials bundle. Bad DOCKER_HOST URL.")
	}

	conn, err := tls.Dial("tcp", dockerHostURL.Host, tlsConfig)
	if err != nil {
		return errors.Wrap(err, "Invalid credentials bundle. Unable to connect to the Docker host.")
	}
	conn.Close()

	return nil
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
