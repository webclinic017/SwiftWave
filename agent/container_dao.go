package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type ContainerConfigWrapper struct {
	ContainerConfig  *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

func (c *Container) Validate() error {
	if c == nil {
		return fmt.Errorf("provided record is nil")
	}
	if c.UUID == "" {
		return fmt.Errorf("UUID is required for container")
	}
	if c.ImageURI == "" {
		return fmt.Errorf("image URI is required for container")
	}
	// check if payload is valid json
	if c.Data == "" {
		return fmt.Errorf("data is required for container")
	}
	if !IsValidJSON(c.Data) {
		return fmt.Errorf("data is not valid json")
	}
	// Try to unmarshal data
	var config *ContainerConfigWrapper
	err := json.Unmarshal([]byte(c.Data), &config)
	if err != nil {
		return fmt.Errorf("data is not valid container config")
	}
	var staticConfigs []StaticConfig
	err = json.Unmarshal([]byte(c.StaticConfigs), &staticConfigs)
	if err != nil {
		return fmt.Errorf("static config is not valid json")
	}
	// static configs name shouldn't be empty or have spaces
	for _, staticConfig := range staticConfigs {
		if staticConfig.Name == "" || strings.Contains(staticConfig.Name, " ") {
			return fmt.Errorf("static config name is not valid")
		}
	}
	return nil
}

func (c *Container) Create() error {
	if err := c.Validate(); err != nil {
		return err
	}
	c.Status = ContainerStatusImagePullPending
	// Create the record
	if err := rwDB.Create(c).Error; err != nil {
		return err
	}
	containersToRun <- c.UUID
	return nil
}

func (c *Container) Remove() error {
	_ = dockerClient.ContainerRemove(context.Background(), c.UUID, container.RemoveOptions{
		RemoveLinks: true,
	})
	err := rwDB.Model(&Container{}).Where("uuid = ?", c.UUID).Delete(&Container{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete from db : %v", err)
	}
	// Try to delete the static configs
	staticConfigs := c.GetStaticConfigs()
	for _, staticConfig := range staticConfigs {
		_ = os.Remove(filepath.Join(configDirectory, staticConfig.Name))
	}
	return nil
}

func (c *Container) GetStatus() ContainerStatus {
	// If it's still on image stage, return status
	if strings.HasPrefix(string(c.Status), "image_") || c.Status == ContainerStatusCreated {
		return c.Status
	}
	// Refresh Status
	container, err := dockerClient.ContainerInspect(context.Background(), c.UUID)

	var status ContainerStatus
	if err != nil {
		status = ContainerStatusNotFound
	}

	if container.State.Running {
		status = ContainerStatusRunning
	} else if container.State.Paused {
		status = ContainerStatusPaused
	} else if container.State.Restarting {
		status = ContainerStatusRestarting
	} else if container.State.ExitCode != 0 {
		status = ContainerStatusExited
	} else {
		status = ContainerStatusNotFound
	}

	if status != c.Status {
		c.UpdateStatus(status)
	}

	return status
}

func (c *Container) UpdateStatus(status ContainerStatus) error {
	c.Status = status
	if rwDB.Model(&Container{}).Where("uuid = ?", c.UUID).Update("status", status).Error != nil {
		return fmt.Errorf("failed to update container status")
	}
	return nil
}

func FetchContainerByUUID(uuid string) (*Container, error) {
	var container Container
	if err := rDB.Where("uuid = ?", uuid).First(&container).Error; err != nil {
		return nil, err
	}
	return &container, nil
}
