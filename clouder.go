package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
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
	cmd    string
	err    error
	out    string
	errout string
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
		cmd:    cmd,
		err:    errOut,
		out:    out,
		errout: stderr}

	c <- outmsg
}

func CreateCluster(cluster Cluster, c chan OutMsg) {

	var options string
	if cluster.psp == true {
		options = "--enable-pod-security-policy"
	}

	cluster_version_opt := fmt.Sprintf(`--cluster-version "%v"`, cluster.clusterVersion)

	var err_out error
	cmd_tpl := `gcloud container --project "%v" clusters create "%v" --zone "%v" \
    --no-enable-basic-auth %v --machine-type "%v" \
    --image-type "COS" --disk-type "pd-standard" --disk-size "100" \
    --scopes "https://www.googleapis.com/auth/devstorage.read_only","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly","https://www.googleapis.com/auth/trace.append" \
    --preemptible --num-nodes "%v" --no-enable-cloud-logging --no-enable-cloud-monitoring \
    --no-enable-ip-alias --network "%v" --subnetwork "%v" \
	--enable-autoscaling --min-nodes "%v" --max-nodes "%v" %v \
    --addons HorizontalPodAutoscaling,HttpLoadBalancing --enable-autoupgrade --enable-autorepair`

	cmd := fmt.Sprintf(cmd_tpl, cluster.project, cluster.name, cluster.zone, cluster_version_opt, cluster.machineType,
		cluster.nbInstance, cluster.network, cluster.subnetwork, cluster.minNodes,
		cluster.maxNodes, options)

	err, out, errout := Shellout(cmd)
	if err != nil {
		err_msg := fmt.Sprintf("error creating %v: %v\n", cluster.name, err)
		err_out = errors.New(err_msg)
	}

	outmsg := OutMsg{
		cmd:    cmd,
		err:    err_out,
		out:    out,
		errout: errout}

	c <- outmsg
}

func BuildClusterList(clusterVersion string, psp bool, nbCluster int, nbInstance int, machineType string, project string,
	regionzones []RegionZone) []Cluster {

	clusters := make([]Cluster, 0)
	for i, rz := range regionzones[0:nbCluster] {
		name := fmt.Sprintf("gke%v", i)
		region := rz.region
		zone := rz.zone
		network := fmt.Sprintf("projects/%v/global/networks/default", project)
		subnetwork := fmt.Sprintf("projects/%v/regions/%v/subnetworks/default", project, region)
		minNodes := 2
		maxNodes := 4

		c := Cluster{
			clusterVersion: clusterVersion,
			project:        project,
			name:           name,
			region:         region,
			zone:           zone,
			machineType:    machineType,
			nbInstance:     nbInstance,
			network:        network,
			subnetwork:     subnetwork,
			minNodes:       minNodes,
			maxNodes:       maxNodes,
			psp:            psp}
		clusters = append(clusters, c)
	}
	return clusters
}

func BuildInstanceClusterList(image string, imageProject string, nbInstanceCluster int, nbInstance int, machineType string, project string,
	regionzones []RegionZone) []InstanceCluster {
	// Create a list of InstanceCluster,
	// where InstanceCluster represents of group of GCE instance in the same RegioZone
	instanceClusters := make([]InstanceCluster, 0, 30)
	for i, rz := range regionzones[0:nbInstanceCluster] {
		name := fmt.Sprintf("clus%v", i)
		region := rz.region
		zone := rz.zone
		image := image

		is := InstanceCluster{
			project:      project,
			name:         name,
			nbInstance:   nbInstance,
			region:       region,
			zone:         zone,
			machineType:  machineType,
			image:        image,
			imageProject: imageProject}
		instanceClusters = append(instanceClusters, is)
	}
	return instanceClusters
}

func CreateAllClusters(instanceClusters []InstanceCluster, clusters []Cluster) error {
	// var err_msgs string
	var errOut error

	log.Printf("Create %v vm clusters", len(instanceClusters))
	chans := make([]chan OutMsg, 0)
	for _, instanceCluster := range instanceClusters {
		c := make(chan OutMsg)
		chans = append(chans, c)
		go CreateInstanceCluster(instanceCluster, c)
	}
	log.Printf("Create %v k8s clusters", len(clusters))
	for _, cluster := range clusters {
		c := make(chan OutMsg)
		chans = append(chans, c)
		go CreateCluster(cluster, c)
	}
	for _, c := range chans {
		outmsg := <-c
		log.Println(outmsg.cmd)
		log.Println(outmsg.out)
		if outmsg.err != nil {
			log.Println(outmsg.err)
			log.Println(outmsg.errout)
		}
	}
	return errOut
}

type Cluster struct {
	clusterVersion string
	project        string
	name           string
	region         string
	zone           string
	machineType    string
	nbInstance     int
	network        string
	subnetwork     string
	minNodes       int
	maxNodes       int
	psp            bool
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
}

type RegionZone struct {
	region string
	zone   string
}

func appendRegionZones(rzs []RegionZone, prefix string, idxs []int, zones []rune) []RegionZone {
	for _, idx := range idxs {
		region := fmt.Sprintf("%v%v", prefix, idx)
		for _, zone := range zones {
			zone := fmt.Sprintf("%v-%v", region, string(zone))
			r := RegionZone{
				region: region,
				zone:   zone}
			rzs = append(rzs, r)
		}
	}
	return rzs
}
