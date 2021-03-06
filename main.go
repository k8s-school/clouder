package main

import (
	"flag"
	"log"
)

func main() {

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
	defaultImage := "ubuntu-1804-bionic-v20200108"

	imageProject := flag.String("image-project", defaultImageProject, "VM image project")
	image := flag.String("image", defaultImage, "VM image")

	flag.Parse()

	regionzones := make([]RegionZone, 0)

	prefix := "asia-east"
	idxs := []int{1, 2}
	zones := []rune{'c', 'c'}
	regionzones = appendRegionZones(regionzones, prefix, idxs, zones)

	prefix = "asia-northeast"
	idxs = []int{1, 2}
	zones = []rune{'c', 'c'}
	regionzones = appendRegionZones(regionzones, prefix, idxs, zones)

	prefix = "asia-southeast"
	idxs = []int{1}
	zones = []rune{'c'}
	regionzones = appendRegionZones(regionzones, prefix, idxs, zones)

	prefix = "europe-north"
	idxs = []int{1}
	zones = []rune{'c'}
	regionzones = appendRegionZones(regionzones, prefix, idxs, zones)

	prefix = "europe-west"
	idxs = []int{1, 2, 3, 4, 6}
	zones = []rune{'c', 'c', 'c', 'c', 'c'}
	regionzones = appendRegionZones(regionzones, prefix, idxs, zones)

	prefix = "northamerica-northeast"
	idxs = []int{1}
	zones = []rune{'c'}
	regionzones = appendRegionZones(regionzones, prefix, idxs, zones)

	prefix = "us-east"
	idxs = []int{1, 4}
	zones = []rune{'c', 'c'}
	regionzones = appendRegionZones(regionzones, prefix, idxs, zones)

	prefix = "us-west"
	idxs = []int{1, 2}
	zones = []rune{'a', 'a'}
	regionzones = appendRegionZones(regionzones, prefix, idxs, zones)

	k8sClusters := BuildClusterList(*clusterVersion, *k8sPsp, *numK8sClusterPtr, *numNodePtr, *machineTypePtr, *projectPtr, regionzones)
	vmClusters := BuildInstanceClusterList(*image, *imageProject, *numVirtualClusterPtr, *numVirtualPtr, *machineTypePtr, *projectPtr,
		regionzones)

	err := CreateAllClusters(vmClusters, k8sClusters)
	if err != nil {
		log.Printf("Error(s) while creating cluster(s): %v", err)

	}
}
