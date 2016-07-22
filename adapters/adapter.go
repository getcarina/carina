package adapters

import (
	"github.com/pkg/errors"
	"strings"
	"text/tabwriter"
)

// Adapter maps between a container service API and the cli
type Adapter interface {
	// CreateCluster creates a new cluster and prints the cluster information
	CreateCluster(name string, nodes int, waitUntilActive bool) error

	// ListClusters prints out a list of the user's clusters to the console
	ListClusters() error

	// ShowCluster prints out a cluster's information to the console
	ShowCluster(name string, waitUntilActive bool) error

	// DeleteCluster permanently deletes a cluster
	DeleteCluster(name string) error

	// GrowCluster adds nodes to a cluster
	GrowCluster(name string, nodes int) error

	// SetAutoScale enables or disables autoscaling on a cluster
	SetAutoScale(name string, value bool) error
}

// WriteRow prints a row of tabular data to the console
func WriteRow(output *tabwriter.Writer, fields []string) error {
	s := strings.Join(fields, "\t")
	_, err := output.Write([]byte(s + "\n"))
	if err != nil {
		return errors.Wrap(err, "Unable to write tabular data to the console")
	}
	return output.Flush()
}
