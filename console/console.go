package console

import (
	"fmt"
	"github.com/getcarina/carina/common"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

// WriteRow writes a row of tabular data to the console
func WriteRow(fields []string) {
	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 20, 1, 3, ' ', 0)

	s := strings.Join(fields, "\t")
	b := []byte(s + "\n")
	_, err := output.Write(b)
	if err == nil {
		_ = output.Flush()
	} else {
		err = errors.Wrap(err, "Unable to write to console.")
		fmt.Println(err.Error())
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
