package adapters

import (
	"github.com/pkg/errors"
	"strings"
	"text/tabwriter"
)

// Maps between a container service API and the command line client
type Adapter interface {
	LoadCredentials(credentials UserCredentials) error
	ListClusters() error
	ShowCluster(name string) error
}

// The credentials supplied by the user to the command line client
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
