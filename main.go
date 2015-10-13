package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	carina "github.com/rackerlabs/libcarina"
	"github.com/samalba/dockerclient"
)

func dockerInfo(creds *carina.Credentials) (*dockerclient.Info, error) {
	tlsConfig, err := creds.GetTLSConfig()
	if err != nil {
		return nil, err
	}

	docker, err := dockerclient.NewDockerClient(creds.DockerHost, tlsConfig)
	if err != nil {
		return nil, err
	}
	info, err := docker.Info()
	return info, err
}

func writeCredentials(w *tabwriter.Writer, creds *carina.Credentials, pth string) (err error) {
	statusFormat := "%s\t%s\n"
	for fname, b := range creds.Files {
		p := path.Join(pth, fname)
		err = ioutil.WriteFile(p, b, 0600)
		if err != nil {
			fmt.Fprintf(w, statusFormat, fname, "🚫")
			return err
		}
		fmt.Fprintf(w, statusFormat, fname, "✅")
	}
	return nil
}

func main() {
	var username, apiKey, endpoint string

	flag.Usage = usage

	flag.StringVar(&username, "username", "", "Rackspace username")
	flag.StringVar(&apiKey, "api-key", "", "Rackspace API Key")
	flag.StringVar(&endpoint, "endpoint", carina.BetaEndpoint, "carina API Endpoint")
	flag.Parse()

	if username == "" && os.Getenv("RACKSPACE_USERNAME") != "" {
		username = os.Getenv("RACKSPACE_USERNAME")
	}
	if apiKey == "" && os.Getenv("RACKSPACE_APIKEY") != "" {
		apiKey = os.Getenv("RACKSPACE_APIKEY")
	}

	if username == "" || apiKey == "" {
		fmt.Println("Either set -username and -api-key or set the " +
			"RACKSPACE_USERNAME and RACKSPACE_APIKEY environment variables.")
		fmt.Println()
		usage()
		os.Exit(1)
	}

	var command, clusterName string

	command = flag.Arg(0)
	clusterName = flag.Arg(1)

	switch {
	case flag.NArg() < 1 || (command != "list" && flag.NArg() < 2):
		usage()
		os.Exit(2)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, '\t', 0)

	clusterClient, err := carina.NewClusterClient(endpoint, username, apiKey)
	if err != nil {
		simpleErr(w, err)
		w.Flush()
		os.Exit(3)
	}

	switch command {
	case "list":
		var clusters []carina.Cluster
		clusters, err = clusterClient.List()
		if err == nil {
			for _, cluster := range clusters {
				writeCluster(w, &cluster, err)
			}
		}
	case "get":
		cluster, err := clusterClient.Get(clusterName)
		writeCluster(w, cluster, err)
	case "delete":
		cluster, err := clusterClient.Delete(clusterName)
		writeCluster(w, cluster, err)
	case "create":
		c := carina.Cluster{
			ClusterName: clusterName,
		}
		cluster, err := clusterClient.Create(c)
		writeCluster(w, cluster, err)
	case "credentials":
		creds, err := clusterClient.GetCredentials(clusterName)
		if err == nil {
			err = writeCredentials(w, creds, ".")
		}

		// Snuck in as an example
	case "docker-info":
		creds, err := clusterClient.GetCredentials(clusterName)
		if err != nil {
			break
		}
		info, err := dockerInfo(creds)
		fmt.Fprintf(w, "%+v\n", info)
	default:
		usage()
		err = errors.New("command " + command + " not recognized")
	}
	exitCode := 0

	if err != nil {
		simpleErr(w, err)
		exitCode = 4
	}

	w.Flush()
	os.Exit(exitCode)
}

func writeCluster(w *tabwriter.Writer, cluster *carina.Cluster, err error) {
	if err != nil {
		return
	}
	s := strings.Join([]string{cluster.ClusterName,
		cluster.Username,
		cluster.Flavor,
		cluster.Image,
		fmt.Sprintf("%v", cluster.Nodes),
		cluster.Status}, "\t")
	w.Write([]byte(s + "\n"))
}

func simpleErr(w *tabwriter.Writer, err error) {
	fmt.Fprintf(w, "ERROR: %v\n", err)
}

func usage() {
	exe := os.Args[0]

	fmt.Printf("NAME:\n")
	fmt.Printf("  %s - command line interface to manage swarm clusters\n", exe)
	fmt.Printf("USAGE:\n")
	fmt.Printf("  %s <command> [clustername] [--username <username>] [--api-key <apiKey>] [--endpoint <endpoint>]\n", exe)
	fmt.Println()
	fmt.Printf("COMMANDS:\n")
	fmt.Printf("  %s list\n", exe)
	fmt.Printf("  %s create <clustername>      - create a new cluster\n", exe)
	fmt.Printf("  %s get <clustername>         - get a cluster by name\n", exe)
	fmt.Printf("  %s delete <clustername>      - delete a cluster by name\n", exe)
	fmt.Printf("  %s credentials <clustername> - download credentials to the current directory\n", exe)
	fmt.Println()
	fmt.Printf("FLAGS:\n")
	fmt.Printf("  -username string\n")
	fmt.Printf("    Rackspace username\n")
	fmt.Printf("  -api-key string\n")
	fmt.Printf("    Rackspace API key\n")
	fmt.Printf("  -endpoint string\n")
	fmt.Printf("    carina API Endpoint (default \"https://mycluster.rackspacecloud.com\")\n")
	fmt.Println()
	fmt.Printf("ENVIRONMENT VARIABLES:\n")
	fmt.Printf("  RACKSPACE_USERNAME - set instead of --username\n")
	fmt.Printf("  RACKSPACE_APIKEY   - set instead of --api-key\n")
}
