package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {

	Version := "0.0.1-rc1"
	fmt.Printf("Version: %v\n", Version)

	projectPtr := flag.String("project", "coastal-sunspot-206412", "Identifier of GCP project")

	machineTypePtr := flag.String("machine-type", "n1-standard-2", "GCE machine type")
	clusterVersion := flag.String("cluster-version", "", "GKE cluster version")

	numK8sClusterPtr := flag.Int("k8s", 0, "Number of GKE/K8S clusters")
	numNodePtr := flag.Int("num-nodes", 2, "Number of nodes in each GKE/K8S cluster")
	k8sPsp := flag.Bool("psp", false, "Enable pod security policies on GKE/K8S cluster")

	// Number of instances cluster to create
	numVirtualClusterPtr := flag.Int("vm", 1, "Number of virtual machine clusters")
	numVirtualPtr := flag.Int("num-vms", 3, "Number of instances in each virtual machine cluster")

	// defaultImageProject := "centos-cloud"
	// defaultImage := "centos-8-v20191210"

	defaultImageProject := "ubuntu-os-cloud"
	defaultImage := "ubuntu-minimal-2004-focal-v20240209a"

	imageProject := flag.String("image-project", defaultImageProject, "VM image project")
	image := flag.String("image", defaultImage, "VM image")

	flag.Parse()

	// Get zone with
	// gcloud compute machine-types list --filter="name=n1-standard-2" --format="value(zone)"
	// gcloud compute machine-types list --filter="name=n1-standard-2" --format="value(zone)"
	regionzones := make([]RegionZone, 0)

	var prefix string
	var ids []string

	prefix = "us-central"
	ids = []string{"1-a", "1-b", "1-c", "1-f", "1-d", "2-a", "2-b", "2-c"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "europe-west"
	ids = []string{"1-b", "1-c", "1-d"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "us-west"
	ids = []string{"1-a", "1-b", "1-c"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "us-east"
	ids = []string{"1-a", "1-b", "1-c"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	fmt.Printf("regionzones: %v %d\n", regionzones, len(regionzones))

	k8sClusters := BuildClusterList(*clusterVersion, *k8sPsp, *numK8sClusterPtr, *numNodePtr, *machineTypePtr, *projectPtr, regionzones)
	vmClusters := BuildInstanceClusterList(*image, *imageProject, *numVirtualClusterPtr, *numVirtualPtr, *machineTypePtr, *projectPtr,
		regionzones)

	err := CreateAllClusters(vmClusters, k8sClusters)
	if err != nil {
		log.Printf("Error(s) while creating cluster(s): %v", err)

	}
}
