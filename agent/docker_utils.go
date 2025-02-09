package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

func (c *AgentConfig) UpdateDockerDaemonConfig() error {
	// Get ip from swiftwave address
	ip, _, err := net.SplitHostPort(c.SwiftwaveServiceAddress)
	if err != nil {
		return fmt.Errorf("failed to parse swiftwave address: %w", err)
	}
	config := map[string]interface{}{
		"live-restore":        true,
		"iptables":            true,
		"insecure-registries": []string{fmt.Sprintf("%s:3331", ip)},
	}
	jsonString, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal docker daemon config: %w", err)
	}
	// Write in /etc/docker/daemon.json
	err = os.WriteFile("/etc/docker/daemon.json", []byte(jsonString), 0644)
	if err != nil {
		return fmt.Errorf("failed to write docker daemon config: %w", err)
	}
	return nil
}
