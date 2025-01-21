package main

import (
	"github.com/docker/docker/client"
)

var dockerClient *client.Client

func init() {
	dockerClientInstance, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	dockerClient = dockerClientInstance
}
