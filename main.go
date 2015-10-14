package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/rackerlabs/libcarina"
)

func writeCluster(w *tabwriter.Writer, cluster *libcarina.Cluster) {
	s := strings.Join([]string{cluster.ClusterName,
		cluster.Username,
		cluster.Flavor,
		cluster.Image,
		fmt.Sprintf("%v", cluster.Nodes),
		cluster.Status}, "\t")
	w.Write([]byte(s + "\n"))
}

func writeCredentials(w *tabwriter.Writer, creds *libcarina.Credentials, pth string) (err error) {
	// TODO: Prompt when file already exists?
	for fname, b := range creds.Files {
		p := path.Join(pth, fname)
		err = ioutil.WriteFile(p, b, 0600)
		if err != nil {
			return err
		}
	}

	// TODO: Handle Windows conditionally
	fmt.Printf("source \"%v\"\n", path.Join(pth, "docker.env"))
	fmt.Printf("# Run the above or use a subshell with your arguments to %v\n", os.Args[0])
	fmt.Printf("# $( %v command... ) \n", os.Args[0])
	return nil
}

// CarinaApplication is, our, well, application
type CarinaApplication struct {
	*kingpin.Application
	TabWriter *tabwriter.Writer
}

// CarinaCommand is a command is a command needing a ClusterClient
type CarinaCommand struct {
	*kingpin.CmdClause
	ClusterClient *libcarina.ClusterClient
	TabWriter     *tabwriter.Writer
	Username      string
	APIKey        string
	Endpoint      string
}

// CarinaClusterCommand is a CarinaCommand with a ClusterName set
type CarinaClusterCommand struct {
	*CarinaCommand
	ClusterName string
}

// CarinaDownloadCommand keeps context about the download command
type CarinaDownloadCommand struct {
	*CarinaClusterCommand
	Path string
}

// NewCarina creates a new CarinaApplication
func NewCarina() *CarinaApplication {

	app := kingpin.New("carina", "command line interface to work with Docker Swarm clusters")

	cap := new(CarinaApplication)

	cap.Application = app

	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 8, 1, '\t', 0)

	cap.TabWriter = writer

	listCommand := cap.NewCarinaCommand(writer, "list", "list swarm clusters")
	listCommand.Action(listCommand.List)

	getCommand := cap.NewCarinaClusterCommand(writer, "get", "get information about a swarm cluster")
	getCommand.Action(getCommand.Get)

	deleteCommand := cap.NewCarinaClusterCommand(writer, "delete", "delete a swarm cluster")
	deleteCommand.Action(deleteCommand.Delete)

	createCommand := cap.NewCarinaClusterCommand(writer, "create", "create a swarm cluster")
	createCommand.Action(createCommand.Create)

	downloadCommand := new(CarinaDownloadCommand)
	downloadCommand.CarinaClusterCommand = cap.NewCarinaClusterCommand(writer, "download", "download credentials")
	downloadCommand.Flag("path", "path to write credentials out to").StringVar(&downloadCommand.Path)
	downloadCommand.Action(downloadCommand.Download)

	return cap
}

// NewCarinaCommand creates a command that relies on Auth
func (app *CarinaApplication) NewCarinaCommand(writer *tabwriter.Writer, name, help string) *CarinaCommand {
	carina := new(CarinaCommand)

	carina.CmdClause = app.Command(name, help)
	carina.Flag("username", "Rackspace username").Default("").OverrideDefaultFromEnvar("RACKSPACE_USERNAME").StringVar(&carina.Username)
	carina.Flag("api-key", "Rackspace API Key").Default("").OverrideDefaultFromEnvar("RACKSPACE_APIKEY").StringVar(&carina.APIKey)
	carina.Flag("endpoint", "Carina API endpoint").Default(libcarina.BetaEndpoint).StringVar(&carina.Endpoint)

	carina.PreAction(carina.Auth)

	carina.TabWriter = new(tabwriter.Writer)
	carina.TabWriter.Init(os.Stdout, 0, 8, 1, '\t', 0)

	return carina
}

// NewCarinaClusterCommand is a command that uses a cluster name
func (app *CarinaApplication) NewCarinaClusterCommand(writer *tabwriter.Writer, name, help string) *CarinaClusterCommand {
	cc := new(CarinaClusterCommand)
	cc.CarinaCommand = app.NewCarinaCommand(writer, name, help)
	cc.Arg("cluster-name", "name of the cluster").Required().StringVar(&cc.ClusterName)
	return cc
}

// Auth does the authentication
func (carina *CarinaCommand) Auth(pc *kingpin.ParseContext) (err error) {
	carina.ClusterClient, err = libcarina.NewClusterClient(carina.Endpoint, carina.Username, carina.APIKey)
	return err
}

// List the current swarm clusters
func (carina *CarinaCommand) List(pc *kingpin.ParseContext) (err error) {
	clusterList, err := carina.ClusterClient.List()
	if err != nil {
		return err
	}

	headerFields := []string{
		"ClusterName",
		"Username",
		"Flavor",
		"Image",
		"Nodes",
		"Status",
	}
	s := strings.Join(headerFields, "\t")

	carina.TabWriter.Write([]byte(s + "\n"))

	for _, cluster := range clusterList {
		writeCluster(carina.TabWriter, &cluster)
	}
	carina.TabWriter.Flush()

	return nil
}

// Get an individual cluster
func (carina *CarinaClusterCommand) Get(pc *kingpin.ParseContext) (err error) {
	cluster, err := carina.ClusterClient.Get(carina.ClusterName)
	if err == nil {
		writeCluster(carina.TabWriter, cluster)
	}
	carina.TabWriter.Flush()
	return err
}

// Delete a cluster
func (carina *CarinaClusterCommand) Delete(pc *kingpin.ParseContext) (err error) {
	cluster, err := carina.ClusterClient.Delete(carina.ClusterName)
	if err == nil {
		writeCluster(carina.TabWriter, cluster)
	}
	carina.TabWriter.Flush()
	return err
}

// Create a cluster
func (carina *CarinaClusterCommand) Create(pc *kingpin.ParseContext) (err error) {
	c := libcarina.Cluster{
		ClusterName: carina.ClusterName,
	}
	cluster, err := carina.ClusterClient.Create(c)
	if err == nil {
		writeCluster(carina.TabWriter, cluster)
	}
	carina.TabWriter.Flush()
	return err
}

// Download credentials for a cluster
func (carina *CarinaDownloadCommand) Download(pc *kingpin.ParseContext) (err error) {
	credentials, err := carina.ClusterClient.GetCredentials(carina.ClusterName)

	p := path.Clean(carina.Path)

	if p != "." {
		os.MkdirAll(p, 0777)
	}

	if err == nil {
		writeCredentials(carina.TabWriter, credentials, p)
	}
	carina.TabWriter.Flush()
	return err
}

func main() {
	app := NewCarina()
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
