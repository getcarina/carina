package console

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/getcarina/carina/common"
	"github.com/pkg/errors"
)

type Tuple struct {
	Key   string
	Value interface{}
}

// Write prints text to the console
func Write(format string, a ...interface{}) {
	if common.Log.IsSilent {
		return
	}

	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	fmt.Printf(format, a...)
}

// WriteTable prints rows of tabular data to the console
func WriteTable(rows [][]string) {
	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 5, 8, 2, ' ', 0)

	for _, row := range rows {
		writeInColumns(output, row)
	}
	output.Flush()
}

// WriteMap prints the cluster data to the console
func WriteMap(items []Tuple) {
	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 5, 8, 2, ' ', 0)

	writeInRows(output, items)
	output.Flush()
}

// WriteCluster prints the cluster data to the console
func WriteCluster(cluster common.Cluster) {
	items := []Tuple{
		{"ID", cluster.GetID()},
		{"Name", cluster.GetName()},
		{"Status", cluster.GetStatus()},
		{"Template", cluster.GetTemplate().GetName()},
		{"Nodes", cluster.GetNodes()},
		{"Details", cluster.GetStatusDetails()},
	}
	WriteMap(items)
}

// WriteClusters prints the clusters data to the console
func WriteClusters(clusters []common.Cluster) {
	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 5, 8, 2, ' ', 0)

	headerFields := []string{
		"ID",
		"Name",
		"Status",
		"Template",
		"Nodes",
	}
	writeInColumns(output, headerFields)

	for _, cluster := range clusters {
		fields := []string{
			cluster.GetID(),
			cluster.GetName(),
			cluster.GetStatus(),
			cluster.GetTemplate().GetName(),
			cluster.GetNodes(),
		}
		writeInColumns(output, fields)
	}

	output.Flush()
}

func writeInColumns(output *tabwriter.Writer, columns []string) {
	s := strings.Join(columns, "\t")
	b := []byte(s + "\n")
	_, err := output.Write(b)
	if err != nil {
		err = errors.Wrap(err, "Unable to write to console.")
		fmt.Println(err.Error())
	}
}

func writeInRows(output *tabwriter.Writer, items []Tuple) {
	for _, item := range items {
		// Use the default string conversion when displaying the value
		val := fmt.Sprint(item.Value)

		// Indent multi-line values
		val = strings.Replace(val, "\n", "\n\t", -1)

		_, err := fmt.Fprintf(output, "%s\t%s\n", item.Key, val)
		if err != nil {
			err = errors.Wrap(err, "Unable to write to console.")
			fmt.Println(err.Error())
		}
	}
}
