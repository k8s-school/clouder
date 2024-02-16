package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const ShellToUse = "bash"

func Shellout(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

type OutMsg struct {
	cluster InstanceCluster
	cmd     string
	err     error
	out     string
	errout  string
}

func CreateInstanceCluster(instanceCluster InstanceCluster, c chan OutMsg) {

	var errOut error

	cmdTpl := `gcloud compute --project="%v" instances create %v \
        --zone="%v" --machine-type="%v" --subnet=default \
        --scopes=https://www.googleapis.com/auth/cloud-platform \
        --image="%v" \
        --image-project="%v" --boot-disk-size=10GB \
        --boot-disk-type=pd-standard \
        --boot-disk-device-name=instance-1`

	instanceNames := ""
	for j := 0; j < instanceCluster.nbInstance; j++ {
		instanceName := fmt.Sprintf("%v-%v", instanceCluster.name, j)
		instanceNames = fmt.Sprintf("%v %v", instanceNames, instanceName)
	}

	cmd := fmt.Sprintf(cmdTpl, instanceCluster.project, instanceNames, instanceCluster.zone,
		instanceCluster.machineType, instanceCluster.image, instanceCluster.imageProject)

	err, out, stderr := Shellout(cmd)
	if err != nil {
		errMsg := fmt.Sprintf("error creating %v: %v\n", instanceCluster.name, err)
		errOut = errors.New(errMsg)
	}

	outmsg := OutMsg{
		cluster: instanceCluster,
		cmd:     cmd,
		err:     errOut,
		out:     out,
		errout:  stderr}

	c <- outmsg
}

func BuildInstanceClusterList(image string, imageProject string, nbInstanceCluster int, nbInstance int, machineType string, project string) []InstanceCluster {
	// Create a list of InstanceCluster,
	// where InstanceCluster represents of group of GCE instance in the same RegioZone
	instanceClusters := make([]InstanceCluster, 0)
	for i := 0; i < nbInstanceCluster; i++ {
		name := fmt.Sprintf("clus%v", i)
		image := image

		is := InstanceCluster{
			project:      project,
			name:         name,
			nbInstance:   nbInstance,
			machineType:  machineType,
			image:        image,
			imageProject: imageProject,
			created:      false}
		instanceClusters = append(instanceClusters, is)
	}
	return instanceClusters
}

func UpdateZones(clusters []InstanceCluster, regionzones []RegionZone) []RegionZone {
	if len(regionzones) < len(clusters) {
		log.Fatalf("not enough regionzones: %v %v", len(clusters), len(regionzones))
	}
	for i := range clusters {
		rz := regionzones[i]
		clusters[i].region = rz.region
		clusters[i].zone = rz.zone
	}
	return regionzones[len(clusters):]
}

func CreateClusters(instanceClusters []InstanceCluster) []InstanceCluster {
	// var err_msgs string
	clustersInError := make([]InstanceCluster, 0)

	log.Printf("Create %d vm clusters", len(instanceClusters))
	chans := make([]chan OutMsg, 0)
	for _, instanceCluster := range instanceClusters {
		c := make(chan OutMsg)
		chans = append(chans, c)
		log.Printf("Creating cluster %v", instanceCluster.name)
		go CreateInstanceCluster(instanceCluster, c)
	}
	for _, c := range chans {
		outmsg := <-c
		log.Println(outmsg.cmd)
		log.Println(outmsg.out)
		if outmsg.err != nil {
			clustersInError = append(clustersInError, outmsg.cluster)
			log.Println(outmsg.err)
			log.Println(outmsg.errout)
		}
	}
	return clustersInError
}

type InstanceCluster struct {
	project      string
	name         string
	nbInstance   int
	region       string
	zone         string
	machineType  string
	image        string
	imageProject string
	created      bool
}

type RegionZone struct {
	region string
	zone   string
}

func appendRegionZones(rzs []RegionZone, prefix string, ids []string) []RegionZone {
	for _, id := range ids {
		// split id on -
		s := strings.Split(id, "-")
		if len(s) != 2 {
			log.Fatalf("invalid id: %v", id)
		}
		region := fmt.Sprintf("%v%v", prefix, s[0])
		zone := fmt.Sprintf("%v-%v", region, s[1])
		r := RegionZone{
			region: region,
			zone:   zone}
		rzs = append(rzs, r)
	}
	return rzs
}
