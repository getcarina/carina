package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	"unicode"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/boltdb/bolt"
	"github.com/getcarina/carina/version"
	"github.com/getcarina/libcarina"
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

// WaitClusterCommand is simply a ClusterCommand that waits for cluster creation
type WaitClusterCommand struct {
	*ClusterCommand
	// Whether to wait until the cluster is created (or errored)
	Wait bool
}

// CreateCommand keeps context about the create command
type CreateCommand struct {
	*WaitClusterCommand

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

// UserNameEnvKey is the name of the env var accepted for the username
const UserNameEnvVar = "CARINA_USERNAME"

// APIKeyEnvVar is the name of the env var accepted for the API key
const APIKeyEnvVar = "CARINA_APIKEY"

// New creates a new Application
func New() *Application {

	app := kingpin.New("carina", "command line interface to launch and work with Docker Swarm clusters")
	app.Version(VersionString())

	cap := new(Application)
	ctx := new(Context)

	cap.Application = app

	cap.Context = ctx

	cap.Flag("username", "Carina username - can also set env var "+UserNameEnvVar).OverrideDefaultFromEnvar(UserNameEnvVar).StringVar(&ctx.Username)
	cap.Flag("api-key", "Carina API Key - can also set env var "+APIKeyEnvVar).OverrideDefaultFromEnvar(APIKeyEnvVar).PlaceHolder(APIKeyEnvVar).StringVar(&ctx.APIKey)
	cap.Flag("endpoint", "Carina API endpoint").Default(libcarina.BetaEndpoint).StringVar(&ctx.Endpoint)

	writer := new(tabwriter.Writer)
	writer.Init(os.Stdout, 20, 1, 3, ' ', 0)

	// Make sure the tabwriter gets flushed at the end
	app.Terminate(func(code int) {
		// Squish any errors from flush, since we're terminating the app anyway
		_ = ctx.TabWriter.Flush()
		os.Exit(code)
	})

	cap.Flag("bash-completion", "Generate bash completion").Action(cap.generateBashCompletion).Hidden().Bool()

	ctx.TabWriter = writer

	createCommand := new(CreateCommand)
	createCommand.WaitClusterCommand = cap.NewWaitClusterCommand(ctx, "create", "Create a swarm cluster")
	createCommand.Flag("nodes", "number of nodes for the initial cluster").Default("1").IntVar(&createCommand.Nodes)
	createCommand.Flag("autoscale", "whether autoscale is on or off").BoolVar(&createCommand.AutoScale)
	createCommand.Action(createCommand.Create)

	getCommand := cap.NewClusterCommand(ctx, "get", "Get information about a swarm cluster")
	getCommand.Action(getCommand.Get)

	listCommand := cap.NewCommand(ctx, "list", "List swarm clusters")
	listCommand.Action(listCommand.List)

	lsCommand := cap.NewCommand(ctx, "ls", "List swarm clusters")
	lsCommand.Action(lsCommand.List).Hidden()

	growCommand := new(GrowCommand)
	growCommand.ClusterCommand = cap.NewClusterCommand(ctx, "grow", "Grow a cluster by the requested number of nodes")
	growCommand.Flag("nodes", "number of nodes to increase the cluster by").Required().IntVar(&growCommand.Nodes)
	growCommand.Action(growCommand.Grow)

	credentialsCommand := cap.NewCredentialsCommand(ctx, "credentials", "download credentials")
	credentialsCommand.Action(credentialsCommand.Download)

	// Hidden shorthand
	credsCommand := cap.NewCredentialsCommand(ctx, "creds", "download credentials")
	credsCommand.Action(credsCommand.Download).Hidden()

	rebuildCommand := cap.NewWaitClusterCommand(ctx, "rebuild", "Rebuild a swarm cluster")
	rebuildCommand.Action(rebuildCommand.Rebuild)

	deleteCommand := cap.NewClusterCommand(ctx, "delete", "Delete a swarm cluster")
	deleteCommand.Action(deleteCommand.Delete)

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

// NewCredentialsCommand is a command that dumps out credentials to a path
func (app *Application) NewCredentialsCommand(ctx *Context, name, help string) *CredentialsCommand {
	credentialsCommand := new(CredentialsCommand)
	credentialsCommand.ClusterCommand = app.NewClusterCommand(ctx, name, help)
	credentialsCommand.Flag("path", "path to write credentials out to").PlaceHolder("<cluster-name>").StringVar(&credentialsCommand.Path)
	return credentialsCommand
}

// NewWaitClusterCommand is a command that uses a cluster name and allows the
// user to wait for a cluster state
func (app *Application) NewWaitClusterCommand(ctx *Context, name, help string) *WaitClusterCommand {
	wcc := new(WaitClusterCommand)
	wcc.ClusterCommand = app.NewClusterCommand(ctx, name, help)
	wcc.Flag("wait", "wait for swarm cluster to come online (or error)").BoolVar(&wcc.Wait)
	return wcc
}

const rackspaceUserNameEnvVar = "RACKSPACE_USERNAME"
const rackspaceAPIKeyEnvVar = "RACKSPACE_APIKEY"

type semver struct {
	Major    int
	Minor    int
	Patch    int
	Leftover string
}

func extractSemver(semi string) (*semver, error) {
	var err error

	if len(semi) < 5 { // 1.3.5
		return nil, errors.New("Invalid semver")
	}
	// Allow a v in front
	if semi[0] == 'v' {
		semi = semi[1:]
	}
	semVerStrings := strings.SplitN(semi, ".", 3)

	if len(semVerStrings) < 3 {
		return nil, errors.New("Could not parse semver")
	}

	parsedSemver := new(semver)

	digitError := errors.New("Could not parse digits of semver")
	if parsedSemver.Major, err = strconv.Atoi(semVerStrings[0]); err != nil {
		return nil, digitError
	}
	if parsedSemver.Minor, err = strconv.Atoi(semVerStrings[1]); err != nil {
		return nil, digitError
	}

	var ps []rune

	// Now to extract the patch and any follow on
	for i, char := range semVerStrings[2] {
		if !unicode.IsDigit(char) {
			parsedSemver.Leftover = semVerStrings[2][i:]
			break
		}
		ps = append(ps, char)
	}

	if parsedSemver.Patch, err = strconv.Atoi(string(ps)); err != nil {
		return nil, digitError
	}

	return parsedSemver, nil

}

func (s *semver) Greater(s2 *semver) bool {
	switch {
	case s.Major == s2.Major && s.Minor == s2.Minor:
		return s.Patch > s2.Patch
	case s.Major == s2.Major:
		return s.Minor > s2.Minor
	}

	return s.Major > s2.Major
}

func (s *semver) String() string {
	return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch)
}

var carinaBoltBucket = []byte("carina")
var lastCheckKey = []byte("last-check")

func carinaDB() (*bolt.DB, error) {
	bd, err := CarinaCredentialsBaseDir()
	if err != nil {
		return nil, err
	}
	boltDB := path.Join(bd, "carina.db")
	db, err := bolt.Open(boltDB, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(carinaBoltBucket)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return db, err
}

func informLatest(pc *kingpin.ParseContext) error {
	db, err := carinaDB()
	if err != nil {
		return err
	}
	defer db.Close()

	var lastCheck time.Time

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(carinaBoltBucket)
		return lastCheck.UnmarshalText(bucket.Get(lastCheckKey))
	})
	if err != nil {
		// TODO: Delete the old database, clean it up?
		return err
	}

	// If we last checked < 12 hours ago, don't check again
	if lastCheck.Add(12 * time.Hour).After(time.Now()) {
		return nil
	}

	lastCheck = time.Now()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(carinaBoltBucket)
		byteTime, err := lastCheck.MarshalText()
		if err != nil {
			return err
		}
		err = bucket.Put(lastCheckKey, byteTime)
		return err
	})
	if err != nil {
		return err
	}

	if strings.Contains(version.Version, "-dev") || version.Version == "" {
		fmt.Fprintln(os.Stderr, "# [WARN] In dev mode, not checking for latest release")
		return nil
	}

	rel, err := version.LatestRelease()
	if err != nil {
		fmt.Fprintf(os.Stderr, "# [WARN] Unable to fetch information about the latest release of %s. %s\n.", os.Args[0], err)
		return nil
	}

	latest, err := extractSemver(rel.TagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "# [WARN] Trouble parsing latest tag (%v): %s\n", rel.TagName, err)
		return nil
	}
	current, err := extractSemver(version.Version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "# [WARN] Trouble parsing current tag (%v): %s\n", version.Version, err)
		return nil
	}

	if latest.Greater(current) {
		fmt.Fprintf(os.Stderr, "# A new version of the Carina client is out, go get it\n")
		fmt.Fprintf(os.Stderr, "# You're on %v and the latest is %v\n", current, latest)
		fmt.Fprintf(os.Stderr, "# https://github.com/getcarina/carina/releases\n")
	}

	return nil
}

// Auth does the authentication
func (carina *Command) Auth(pc *kingpin.ParseContext) (err error) {

	// Check for the latest release.
	informLatest(pc)

	if carina.Username == "" || carina.APIKey == "" {
		// Backwards compatibility for prior releases, to be deprecated
		// Check on RACKSPACE_USERNAME
		if os.Getenv(rackspaceUserNameEnvVar) != "" && os.Getenv(rackspaceAPIKeyEnvVar) != "" {
			fmt.Fprintf(os.Stderr, "Warning: use of %s and %s environment variables is deprecated.\n", rackspaceUserNameEnvVar, rackspaceAPIKeyEnvVar)
			fmt.Fprintf(os.Stderr, "Please use %s and %s instead.\n", UserNameEnvVar, APIKeyEnvVar)
			carina.Username = os.Getenv(rackspaceUserNameEnvVar)
			carina.APIKey = os.Getenv(rackspaceAPIKeyEnvVar)
		}
	}

	carina.ClusterClient, err = libcarina.NewClusterClient(carina.Endpoint, carina.Username, carina.APIKey)
	return err
}

// List the current swarm clusters
func (carina *Command) List(pc *kingpin.ParseContext) (err error) {
	clusterList, err := carina.ClusterClient.List()
	if err != nil {
		return err
	}

	err = writeClusterHeader(carina.TabWriter)
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

// Rebuild nukes your cluster and builds it over again
func (carina *WaitClusterCommand) Rebuild(pc *kingpin.ParseContext) (err error) {
	return carina.clusterApplyWait(carina.ClusterClient.Rebuild)
}

const startupFudgeFactor = 40 * time.Second
const waitBetween = 10 * time.Second

// Cluster status when new
const StatusNew = "new"

// Cluster status when building
const StatusBuilding = "building"

// Cluster status when rebuilding swarm
const StatusRebuildingSwarm = "rebuilding-swarm"

// Does an func against a cluster then returns the new cluster representation
func (carina *WaitClusterCommand) clusterApplyWait(op clusterOp) (err error) {
	cluster, err := op(carina.ClusterName)

	if carina.Wait {
		time.Sleep(startupFudgeFactor)
		// Transitions past point of "new" or "building" are assumed to be states we
		// can stop on.
		for cluster.Status == StatusNew || cluster.Status == StatusBuilding || cluster.Status == StatusRebuildingSwarm {
			time.Sleep(waitBetween)
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

// CredentialsBaseDirEnvVar environment variable name for where credentials are downloaded to by default
const CredentialsBaseDirEnvVar = "CARINA_CREDENTIALS_DIR"

// Create a cluster
func (carina *CreateCommand) Create(pc *kingpin.ParseContext) (err error) {
	return carina.clusterApplyWait(func(clusterName string) (*libcarina.Cluster, error) {
		if carina.Nodes < 1 {
			return nil, errors.New("nodes must be >= 1")
		}
		nodes := libcarina.Number(carina.Nodes)

		c := libcarina.Cluster{
			ClusterName: carina.ClusterName,
			Nodes:       nodes,
			AutoScale:   carina.AutoScale,
		}
		return carina.ClusterClient.Create(c)
	})
}

const clusterDirName = "clusters"

// Download credentials for a cluster
func (carina *CredentialsCommand) Download(pc *kingpin.ParseContext) (err error) {
	credentials, err := carina.ClusterClient.GetCredentials(carina.ClusterName)
	if err != nil {
		return err
	}

	if carina.Path == "" {
		baseDir, err := CarinaCredentialsBaseDir()
		if err != nil {
			return err
		}
		carina.Path = path.Join(baseDir, clusterDirName, carina.Username, carina.ClusterName)
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

	fmt.Fprintln(os.Stdout, sourceHelpString(p))

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
	s := strings.Join([]string{
		cluster.ClusterName,
		cluster.Flavor,
		strconv.FormatInt(cluster.Nodes.Int64(), 10),
		strconv.FormatBool(cluster.AutoScale),
		cluster.Status,
	}, "\t")
	_, err = w.Write([]byte(s + "\n"))
	return
}

func writeClusterHeader(w *tabwriter.Writer) (err error) {
	headerFields := []string{
		"ClusterName",
		"Flavor",
		"Nodes",
		"AutoScale",
		"Status",
	}
	s := strings.Join(headerFields, "\t")

	_, err = w.Write([]byte(s + "\n"))
	return err
}

func (app *Application) generateBashCompletion(c *kingpin.ParseContext) error {
	app.Writer(os.Stdout)
	if err := app.UsageForContextWithTemplate(c, 2, BashCompletionTemplate); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}

func main() {
	app := New()
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
