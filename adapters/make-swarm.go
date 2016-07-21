package adapters

import (
	"fmt"
	"text/tabwriter"
)

type Carina struct {
	Credentials UserCredentials
	Output *tabwriter.Writer
}

func (carina *Carina) LoadCredentials(credentials UserCredentials) error {
	carina.Credentials = credentials
	return nil
}

func (carina *Carina) ListClusters() error {
	fmt.Println("[DEBUG] Listing Carina clusters...")
	/*
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
	   	return err*/
	return nil
}
