package console

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/getcarina/carina/common"
	"github.com/pkg/errors"
)

type tuple struct {
	key   string
	value interface{}
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
func WriteTable(fields [][]string) {
	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 0, 8, 1, '\t', 0)

	for _, row := range fields {
		writeInColumns(output, row)
	}
	output.Flush()
}

// WriteRow prints a row of tabular data to the console
func WriteRow(fields []string) {
	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 0, 8, 1, '\t', 0)

	writeInColumns(output, fields)
	output.Flush()
}

// WriteCluster prints the cluster data to the console
func WriteCluster(cluster common.Cluster) {
	output := new(tabwriter.Writer)
	output.Init(os.Stdout, 0, 8, 2, '\t', 0)

	fields := []tuple{
		tuple{"ID", cluster.GetID()},
		tuple{"Name", cluster.GetName()},
		tuple{"Status", cluster.GetStatus()},
		tuple{"Type", cluster.GetType()},
		tuple{"Nodes", cluster.GetNodes()},
		tuple{"Details", cluster.GetStatusDetails()},
	}
	writeInRows(output, fields)
	output.Flush()
}

// WriteClusters prints the clusters data to the console
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
	writeInColumns(output, headerFields)

	for _, cluster := range clusters {
		fields := []string{
			cluster.GetID(),
			cluster.GetName(),
			cluster.GetStatus(),
			cluster.GetType(),
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

func writeInRows(output *tabwriter.Writer, rows []tuple) {
	for _, row := range rows {
		// Use the default string conversion when displaying the value
		val := fmt.Sprint(row.value)

		// Indent multi-line values
		val = strings.Replace(val, "\n", "\n\t", -1)

		_, err := fmt.Fprintf(output, "%s\t%s\n", row.key, val)
		if err != nil {
			err = errors.Wrap(err, "Unable to write to console.")
			fmt.Println(err.Error())
		}
	}
}
