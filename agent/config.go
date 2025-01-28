package main

import (
	"errors"

	"gorm.io/gorm"
)

type AgentConfig struct {
	WireguardConfig WireguardConfig     `json:"wireguard_config" gorm:"embedded;embeddedPrefix:wireguard_"`
	DockerNetwork   DockerNetworkConfig `json:"docker_network" gorm:"embedded;embeddedPrefix:docker_network_"`
}

type WireguardConfig struct {
	PrivateKey string `json:"private_key" gorm:"column:private_key"`
	Address    string `json:"address" gorm:"column:address"`
	CIDR       int    `json:"cidr" gorm:"column:cidr"`
}

type DockerNetworkConfig struct {
	GatewayAddress string `json:"gateway_address" gorm:"column:gateway_address"`
	Subnet         string `json:"subnet" gorm:"column:subnet"`
	CIDR           int    `json:"cidr" gorm:"column:cidr"`
}

func GetConfig() (*AgentConfig, error) {
	var config AgentConfig
	if err := rDB.First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AgentConfig{}, nil
		}
	}
	return &config, nil
}

func SetConfig(config *AgentConfig) error {
	// Check if the config already exists
	var count int64
	if err := rDB.Model(&AgentConfig{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		// Update the existing config
		return rDB.Save(config).Error
	}
	// Create a new config
	return rDB.Create(config).Error
}
