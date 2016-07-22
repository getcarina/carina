package adapters

import (
	"github.com/pkg/errors"
	"strings"
	"text/tabwriter"
)

// Adapter maps between a container service API and the cli
type Adapter interface {
	// LoadCredentials accepts credentials collected by the cli
	LoadCredentials(credentials UserCredentials) error

	// ListClusters prints out a list of the user's clusters to the console
	ListClusters() error

	// ShowCluster prints out a cluster's information to the console
	ShowCluster(name string) error
}

// UserCredentials is the set of authentication credentials discovered by the cli
type UserCredentials struct {
	Endpoint        string
	UserName        string
	Secret          string
	Project         string
	Domain          string
	Region          string
	Token           string
	TokenExpiration string
}

func writeRow(output *tabwriter.Writer, fields []string) error {
	s := strings.Join(fields, "\t")
	_, err := output.Write([]byte(s + "\n"))
	if err != nil {
		return errors.Wrap(err, "Unable to write tabular data to the console")
	}
	return output.Flush()
}
