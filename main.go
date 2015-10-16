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

	"github.com/rackerlabs/carina/version"
	"github.com/rackerlabs/libcarina"
)

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

	// Options passed along to Carina's API
	Nodes     int
	AutoScale bool

	// Whether to wait until the cluster is created (or errored)
	Wait bool

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

	app := kingpin.New("carina", "command line interface to launch and work with Docker Swarm clusters")
	app.Version(VersionString())

	cap := new(Application)
	ctx := new(Context)

	cap.Application = app

	cap.Context = ctx

	cap.Flag("username", "Rackspace username - can also set env var RACKSPACE_USERNAME").OverrideDefaultFromEnvar("RACKSPACE_USERNAME").StringVar(&ctx.Username)
	cap.Flag("api-key", "Rackspace API Key - can also set env var RACKSPACE_APIKEY").OverrideDefaultFromEnvar("RACKSPACE_APIKEY").PlaceHolder("RACKSPACE_APIKEY").StringVar(&ctx.APIKey)
	cap.Flag("endpoint", "Carina API endpoint").Default(libcarina.BetaEndpoint).StringVar(&ctx.Endpoint)

	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 0, 8, 1, '\t', 0)

	// Make sure the tabwriter gets flushed at the end
	app.Terminate(func(code int) {
		ctx.TabWriter.Flush()
		os.Exit(code)
	})

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

// VersionString returns the current version and commit of this binary (if set)
func VersionString() string {
	s := ""
	s += fmt.Sprintf("Version: %s\n", version.Version)
	s += fmt.Sprintf("Commit:  %s", version.Commit)
	return s
}

// NewCommand creates a command wrapped with carina.Context
func (app *Application) NewCommand(ctx *Context, name, help string) *Command {
	carina := new(Command)
	carina.Context = ctx
	carina.CmdClause = app.Command(name, help)
	carina.PreAction(carina.Auth)
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
func (carina *Command) Auth(pc *kingpin.ParseContext) (err error) {
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

	_, err = carina.TabWriter.Write([]byte(s + "\n"))
	if err != nil {
		return err
	}

	for _, cluster := range clusterList {
		err = writeCluster(carina.TabWriter, &cluster)
		if err != nil {
			return err
		}
	}
	err = carina.TabWriter.Flush()
	return err
}

type clusterOp func(clusterName string) (*libcarina.Cluster, error)

// Does an func against a cluster then returns the new cluster representation
func (carina *ClusterCommand) clusterApply(op clusterOp) (err error) {
	cluster, err := op(carina.ClusterName)
	if err != nil {
		return err
	}

	err = writeCluster(carina.TabWriter, cluster)
	if err != nil {
		return err
	}
	return carina.TabWriter.Flush()
}

// Get an individual cluster
func (carina *ClusterCommand) Get(pc *kingpin.ParseContext) (err error) {
	return carina.clusterApply(carina.ClusterClient.Get)
}

// Delete a cluster
func (carina *ClusterCommand) Delete(pc *kingpin.ParseContext) (err error) {
	return carina.clusterApply(carina.ClusterClient.Delete)
}

// Grow increases the size of the given cluster
func (carina *GrowCommand) Grow(pc *kingpin.ParseContext) (err error) {
	return carina.clusterApply(func(clusterName string) (*libcarina.Cluster, error) {
		return carina.ClusterClient.Grow(clusterName, carina.Nodes)
	})
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

	if err != nil {
		return err
	}

	err = writeCluster(carina.TabWriter, cluster)

	if err != nil {
		return err
	}
	return carina.TabWriter.Flush()
}

// Download credentials for a cluster
func (carina *CredentialsCommand) Download(pc *kingpin.ParseContext) (err error) {
	credentials, err := carina.ClusterClient.GetCredentials(carina.ClusterName)
	if err != nil {
		return err
	}

	p := path.Clean(carina.Path)

	if p != "." {
		err = os.MkdirAll(p, 0777)
	}

	if err != nil {
		return err
	}

	err = writeCredentials(carina.TabWriter, credentials, p)
	if err != nil {
		return err
	}

	fmt.Println(sourceHelpString(p, os.Args[0]))

	err = carina.TabWriter.Flush()
	return err
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

func writeCluster(w *tabwriter.Writer, cluster *libcarina.Cluster) (err error) {
	s := strings.Join([]string{cluster.ClusterName,
		cluster.Username,
		cluster.Flavor,
		cluster.Image,
		fmt.Sprintf("%v", cluster.Nodes),
		cluster.Status}, "\t")
	_, err = w.Write([]byte(s + "\n"))
	return
}

func main() {
	app := New()
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
