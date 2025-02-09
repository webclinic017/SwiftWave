package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
)

var configDirectory = "/root/docker-configs"

func init() {
	// Create the default volume directory if it doesn't exist
	if _, err := os.Stat(configDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(configDirectory, 0700)
		if err != nil {
			fmt.Printf("Failed to create volume binds directory: %v", err)
			os.Exit(1)
		}
	}
}

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

func (c *Container) GetStaticConfigs() []StaticConfig {
	var staticConfigs []StaticConfig
	err := json.Unmarshal([]byte(c.StaticConfigs), &staticConfigs)
	if err != nil {
		return []StaticConfig{}
	}
	return staticConfigs
}

func (c *Container) Run() {
	var config *ContainerConfigWrapper
	err := json.Unmarshal([]byte(c.Data), &config)
	if err != nil {
		return
	}
	staticConfigs := c.GetStaticConfigs()
	// Write the static configs to the config directory
	for _, staticConfig := range staticConfigs {
		err = os.WriteFile(filepath.Join(configDirectory, staticConfig.Name), []byte(staticConfig.Content), 0777)
		if err != nil {
			fmt.Printf("Failed to write static config: %v", err)
		}
	}
	_, err = dockerClient.ContainerCreate(context.Background(), config.ContainerConfig, config.HostConfig, config.NetworkingConfig, nil, c.UUID)
	if err != nil {
		_ = c.UpdateStatus(ContainerStatusCreationFailed)
		return
	}
	_ = c.UpdateStatus(ContainerStatusCreated)
}
