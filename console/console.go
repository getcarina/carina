package console

import (
	"text/tabwriter"
	"strings"
	"os"

	carinaclient "github.com/getcarina/carina/client"
)

var output *tabwriter.Writer
var Err error

func init() {
	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 20, 1, 3, ' ', 0)
}

// WriteRow writes a row of tabular data to the console
func WriteRow(fields []string) {
	if Err != nil {
		return
	}

	s := strings.Join(fields, "\t")
	_, Err = output.Write([]byte(s + "\n"))
	
	if Err != nil {
		Err = output.Flush()
	}
}

// WriteClusterHeader writes the cluster table header to the console
func WriteClusterHeader() {
	headerFields := []string{
		"ClusterName",
		"Flavor",
		"Nodes",
		"Status",
	}
	WriteRow(headerFields)
}

// WriteCluster writes the cluster data to the console
func WriteCluster(cluster carinaclient.Cluster) {
	fields := []string{
		cluster.GetName(),
		cluster.GetFlavor(),
		cluster.GetNodes(),
		cluster.GetStatus(),
	}
	return WriteRow(fields)
}