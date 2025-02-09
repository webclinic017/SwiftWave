package main

import (
	"context"
	"encoding/json"
	"runtime"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
)

func (c *Container) PullImage() {
	_ = c.UpdateStatus(ContainerStatusImagePulling)
	_, err := dockerClient.ImagePull(context.Background(), c.ImageURI, image.PullOptions{
		RegistryAuth: c.ImageAuthHeader,
	})
	if err != nil {
		if errdefs.IsUnauthorized(err) {
			_ = c.UpdateStatus(ContainerStatusImagePullAuthError)
		} else {
			_ = c.UpdateStatus(ContainerStatusImagePullFailed)
		}
	}
	// Now, check the arch of the image
	imageInfo, _, err := dockerClient.ImageInspectWithRaw(context.Background(), c.ImageURI)
	if err != nil {
		_ = c.UpdateStatus(ContainerStatusImagePullFailed)
		return
	}
	// Validate the image arch
	if imageInfo.Architecture != runtime.GOARCH {
		_ = c.UpdateStatus(ContainerStatusImagePullFailed)
		return
	}
	_ = c.UpdateStatus(ContainerStatusImagePulled)
}

func (c *Container) Run() {
	var config *ContainerConfigWrapper
	err := json.Unmarshal([]byte(c.Data), &config)
	if err != nil {
		return
	}
	_, err = dockerClient.ContainerCreate(context.Background(), config.ContainerConfig, config.HostConfig, config.NetworkingConfig, nil, c.UUID)
	if err != nil {
		_ = c.UpdateStatus(ContainerStatusCreationFailed)
		return
	}
	_ = c.UpdateStatus(ContainerStatusCreated)
}
