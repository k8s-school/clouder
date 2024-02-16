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

	machineTypePtr := flag.String("machine-type", "n1-standard-1", "GCE machine type")

	// Number of instances cluster to create
	numVirtualClusterPtr := flag.Int("vm", 1, "Number of virtual machine clusters")
	numVirtualPtr := flag.Int("num-vms", 2, "Number of instances in each virtual machine cluster")

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
	ids = []string{"1-a", "1-b", "2-a", "2-b"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "europe-west"
	ids = []string{"1-b", "1-c", "2-b", "2-c"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "europe-central"
	ids = []string{"2-a", "2-b", "2-c"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "europe-north"
	ids = []string{"1-a", "1-b", "1-c"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "us-west"
	ids = []string{"1-a", "1-b"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "us-east"
	ids = []string{"1-a", "2-a"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "asia-east"
	ids = []string{"1-a", "2-a"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "asia-northeast"
	ids = []string{"1-a"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	prefix = "asia-southeast"
	ids = []string{"1-a"}
	regionzones = appendRegionZones(regionzones, prefix, ids)

	fmt.Printf("regionzones: %v %d\n", regionzones, len(regionzones))

	vmClusters := BuildInstanceClusterList(*image, *imageProject, *numVirtualClusterPtr, *numVirtualPtr, *machineTypePtr, *projectPtr)

	hasError := true
	zones := regionzones
	for hasError && len(regionzones) > 0 {

		zones = UpdateZones(vmClusters, zones)
		vmClusters = CreateClusters(vmClusters)
		if len(vmClusters) != 0 {
			log.Println("Error creating clusters, retrying")

		}
	}
}
