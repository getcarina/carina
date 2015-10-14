package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/tabwriter"
	"time"

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

	return nil
}

// Application is, our, well, application
type Application struct {
	*Context
	*kingpin.Application
}

// Command is a command needing a ClusterClient
type Command struct {
	*Context
	*kingpin.CmdClause
}

// Context context for the  App
type Context struct {
	ClusterClient *libcarina.ClusterClient
	TabWriter     *tabwriter.Writer
	Username      string
	APIKey        string
	Endpoint      string
}

// ClusterCommand is a Command with a ClusterName set
type ClusterCommand struct {
	*Command
	ClusterName string
}

// CredentialsCommand keeps context about the download command
type CredentialsCommand struct {
	*ClusterCommand
	Path string
}

// CreateCommand keeps context about the create command
type CreateCommand struct {
	*ClusterCommand

	Wait bool

	// Options passed along to Carina's API
	Nodes     int
	AutoScale bool

	// TODO: See if setting flavor or image makes sense, even if the API takes it
	// Flavor    string
	// Image     string
}

// GrowCommand keeps context about the number of nodes to scale by
type GrowCommand struct {
	*ClusterCommand
	Nodes int
}

// New creates a new Application
func New() *Application {

	app := kingpin.New("carina", "command line interface to work with Docker Swarm clusters")

	cap := new(Application)
	ctx := new(Context)

	cap.Application = app
	cap.Context = ctx

	cap.PreAction(cap.Auth)

	cap.Flag("username", "Rackspace username - can also set env var RACKSPACE_USERNAME").OverrideDefaultFromEnvar("RACKSPACE_USERNAME").StringVar(&ctx.Username)
	cap.Flag("api-key", "Rackspace API Key - can also set env var RACKSPACE_APIKEY").OverrideDefaultFromEnvar("RACKSPACE_APIKEY").PlaceHolder("RACKSPACE_APIKEY").StringVar(&ctx.APIKey)
	cap.Flag("endpoint", "Carina API endpoint").Default(libcarina.BetaEndpoint).StringVar(&ctx.Endpoint)

	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 8, 1, '\t', 0)

	ctx.TabWriter = writer

	listCommand := cap.NewCommand(ctx, "list", "list swarm clusters")
	listCommand.Action(listCommand.List)

	getCommand := cap.NewClusterCommand(ctx, "get", "get information about a swarm cluster")
	getCommand.Action(getCommand.Get)

	deleteCommand := cap.NewClusterCommand(ctx, "delete", "delete a swarm cluster")
	deleteCommand.Action(deleteCommand.Delete)

	createCommand := new(CreateCommand)
	createCommand.ClusterCommand = cap.NewClusterCommand(ctx, "create", "create a swarm cluster")
	createCommand.Flag("wait", "wait for swarm cluster completion").BoolVar(&createCommand.Wait)
	createCommand.Flag("nodes", "number of nodes for the initial cluster").Default("1").IntVar(&createCommand.Nodes)
	createCommand.Flag("autoscale", "whether autoscale is on or off").BoolVar(&createCommand.AutoScale)
	createCommand.Action(createCommand.Create)

	credentialsCommand := new(CredentialsCommand)
	credentialsCommand.ClusterCommand = cap.NewClusterCommand(ctx, "credentials", "download credentials")
	credentialsCommand.Flag("path", "path to write credentials out to").StringVar(&credentialsCommand.Path)
	credentialsCommand.Action(credentialsCommand.Download)

	growCommand := new(GrowCommand)
	growCommand.ClusterCommand = cap.NewClusterCommand(ctx, "grow", "Grow a cluster by the requested number of nodes")
	growCommand.Flag("nodes", "number of nodes to increase the cluster by").Required().IntVar(&growCommand.Nodes)
	growCommand.Action(growCommand.Grow)

	return cap
}

// NewCommand creates a command that relies on Auth
func (app *Application) NewCommand(ctx *Context, name, help string) *Command {
	carina := new(Command)
	carina.Context = ctx
	carina.CmdClause = app.Command(name, help)
	return carina
}

// NewClusterCommand is a command that uses a cluster name
func (app *Application) NewClusterCommand(ctx *Context, name, help string) *ClusterCommand {
	cc := new(ClusterCommand)
	cc.Command = app.NewCommand(ctx, name, help)
	cc.Arg("cluster-name", "name of the cluster").Required().StringVar(&cc.ClusterName)
	return cc
}

// Auth does the authentication
func (app *Application) Auth(pc *kingpin.ParseContext) (err error) {
	carina := app.Context
	carina.ClusterClient, err = libcarina.NewClusterClient(carina.Endpoint, carina.Username, carina.APIKey)
	return err
}

// List the current swarm clusters
func (carina *Command) List(pc *kingpin.ParseContext) (err error) {
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
func (carina *ClusterCommand) Get(pc *kingpin.ParseContext) (err error) {
	cluster, err := carina.ClusterClient.Get(carina.ClusterName)
	if err == nil {
		writeCluster(carina.TabWriter, cluster)
	}
	carina.TabWriter.Flush()
	return err
}

// Delete a cluster
func (carina *ClusterCommand) Delete(pc *kingpin.ParseContext) (err error) {
	cluster, err := carina.ClusterClient.Delete(carina.ClusterName)
	if err == nil {
		writeCluster(carina.TabWriter, cluster)
	}
	carina.TabWriter.Flush()
	return err
}

// Create a cluster
func (carina *CreateCommand) Create(pc *kingpin.ParseContext) (err error) {
	if carina.Nodes < 1 {
		return errors.New("nodes must be >= 1")
	}

	nodes := libcarina.Number(carina.Nodes)

	c := libcarina.Cluster{
		ClusterName: carina.ClusterName,
		Nodes:       nodes,
		AutoScale:   carina.AutoScale,
	}

	cluster, err := carina.ClusterClient.Create(c)

	// Transitions past point of "new" or "building" are assumed to be states we
	// can stop on.
	if carina.Wait {
		for cluster.Status == "new" || cluster.Status == "building" {
			time.Sleep(13 * time.Second)
			cluster, err = carina.ClusterClient.Get(carina.ClusterName)
			if err != nil {
				break
			}
		}
	}

	if err == nil {
		writeCluster(carina.TabWriter, cluster)
	}
	carina.TabWriter.Flush()
	return err
}

// Grow increase the size of the given cluster
func (carina *GrowCommand) Grow(pc *kingpin.ParseContext) (err error) {
	cluster, err := carina.ClusterClient.Grow(carina.ClusterName, carina.Nodes)
	if err == nil {
		writeCluster(carina.TabWriter, cluster)
	}
	carina.TabWriter.Flush()
	return err
}

// Download credentials for a cluster
func (carina *CredentialsCommand) Download(pc *kingpin.ParseContext) (err error) {
	credentials, err := carina.ClusterClient.GetCredentials(carina.ClusterName)

	p := path.Clean(carina.Path)

	if p != "." {
		os.MkdirAll(p, 0777)
	}

	if err != nil {
		return err
	}

	writeCredentials(carina.TabWriter, credentials, p)
	// TODO: Handle Windows conditionally
	fmt.Fprintf(os.Stdout, "source \"%v\"\n", path.Join(p, "docker.env"))
	fmt.Fprintf(os.Stdout, "# Run the above or use a subshell with your arguments to %v\n", os.Args[0])
	fmt.Fprintf(os.Stdout, "# $( %v command... ) \n", os.Args[0])

	carina.TabWriter.Flush()
	return err
}

func main() {
	app := New()
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
