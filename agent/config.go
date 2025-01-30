package main

import (
	"errors"

	"gorm.io/gorm"
)

type AgentConfig struct {
	ID              uint                `gorm:"primaryKey"`
	WireguardConfig WireguardConfig     `json:"wireguard_config" gorm:"embedded;embeddedPrefix:wireguard_"`
	DockerNetwork   DockerNetworkConfig `json:"docker_network" gorm:"embedded;embeddedPrefix:docker_network_"`
}

type WireguardConfig struct {
	PrivateKey string `json:"private_key" gorm:"column:private_key"`
	Address    string `json:"address" gorm:"column:address"`
	CIDR       int    `json:"cidr" gorm:"column:cidr"`
}

type DockerNetworkConfig struct {
	BridgeId       string `json:"bridge_id" gorm:"column:bridge_id"`
	GatewayAddress string `json:"gateway_address" gorm:"column:gateway_address"`
	Subnet         string `json:"subnet" gorm:"column:subnet"`
	CIDR           int    `json:"cidr" gorm:"column:cidr"`
}

func GetConfig() (*AgentConfig, error) {
	var config AgentConfig
	if err := rDB.First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &AgentConfig{
				WireguardConfig: WireguardConfig{},
				DockerNetwork:   DockerNetworkConfig{},
			}, nil
		} else {
			return nil, errors.New("error getting config")
		}
	}
	return &config, nil
}

func SetConfig(db *gorm.DB, config *AgentConfig) error {
	// Check if the config already exists
	var count int64
	if err := db.Model(&AgentConfig{}).Count(&count).Error; err != nil {
		return err
	}
	config.ID = 1 // set so that it can be updated
	if count > 0 {
		// Update the existing config
		return db.Save(config).Error
	}
	return db.Create(config).Error
}
