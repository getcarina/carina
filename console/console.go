package console

import (
	"fmt"
	"github.com/getcarina/carina/common"
	"github.com/pkg/errors"
	"os"
	"strings"
	"text/tabwriter"
)

// WriteRow writes a row of tabular data to the console
func WriteRow(fields []string) {
	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 0, 8, 1, '\t', 0)

	write(output, fields)
	output.Flush()
}

// WriteCluster writes teh cluster data to the console
func WriteCluster(cluster common.Cluster) {
	WriteClusters([]common.Cluster{cluster})
}

// WriteClusters writes the clusters data to the console
func WriteClusters(clusters []common.Cluster) {
	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 0, 8, 1, '\t', 0)

	headerFields := []string{
		"ID",
		"Name",
		"Status",
		"Type",
		"Nodes",
	}
	write(output, headerFields)

	for _, cluster := range clusters {
		fields := []string{
			cluster.GetID(),
			cluster.GetName(),
			cluster.GetStatus(),
			cluster.GetType(),
			cluster.GetNodes(),
		}
		write(output, fields)
	}

	output.Flush()
}

func write(output *tabwriter.Writer, fields []string) {
	s := strings.Join(fields, "\t")
	b := []byte(s + "\n")
	_, err := output.Write(b)
	if err != nil {
		err = errors.Wrap(err, "Unable to write to console.")
		fmt.Println(err.Error())
	}
}
