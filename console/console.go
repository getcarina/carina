package console

import (
	"github.com/getcarina/carina/common"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

var Err error

// WriteRow writes a row of tabular data to the console
func WriteRow(fields []string) {
	if Err != nil {
		return
	}

	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 20, 1, 3, ' ', 0)

	s := strings.Join(fields, "\t")
	b := []byte(s + "\n")
	_, Err = output.Write(b)

	if Err == nil {
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
func WriteCluster(cluster common.Cluster) {
	fields := []string{
		cluster.GetName(),
		cluster.GetFlavor(),
		strconv.Itoa(cluster.GetNodes()),
		cluster.GetStatus(),
	}
	WriteRow(fields)
}
