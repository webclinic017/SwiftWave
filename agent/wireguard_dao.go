package main

import (
	"fmt"
	"net"
	"strings"
)

func (w *WireguardPeer) Validate() error {
	if w == nil {
		return fmt.Errorf("provided record is nil")
	}
	if w.PublicKey == "" {
		return fmt.Errorf("public key is required for wireguard peer")
	}
	if w.EndpointIP == "" {
		return fmt.Errorf("endpoint ip is required for wireguard peer")
	}
	if w.AllowedIPs == "" {
		return fmt.Errorf("allowed ips is required for wireguard peer")
	}
	// split allowed ips
	allowedIPs := strings.Split(w.AllowedIPs, ",")
	// check if valid parsable address
	for _, ip := range allowedIPs {
		_, ipNet, err := net.ParseCIDR(strings.TrimSpace(ip))
		if err != nil {
			return err
		}
		if ipNet == nil {
			return fmt.Errorf("invalid allowed ip: %s", ip)
		}
	}
	return nil
}

func (w *WireguardPeer) Create() error {
	if err := w.Validate(); err != nil {
		return err
	}
	// check if another exists with the same public key
	if exists, err := ExistsWireguardPeer(w.PublicKey); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("wireguard peer already exists")
	}
	tx := rwDB.Begin()
	defer tx.Rollback()
	err := tx.Create(w).Error
	if err != nil {
		return err
	}
	// Try to reconfigure wireguard
	if err := ConfigureWireguardPeers(); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (w *WireguardPeer) Remove() error {
	tx := rwDB.Begin()
	defer tx.Rollback()
	err := tx.Delete(w).Error
	if err != nil {
		return err
	}
	// Try to reconfigure wireguard
	err = ConfigureWireguardPeers()
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func FetchWireguardPeerByPublicKey(publicKey string) (*WireguardPeer, error) {
	var peer WireguardPeer
	if err := rDB.Where("public_key = ?", publicKey).First(&peer).Error; err != nil {
		return nil, err
	}
	return &peer, nil
}

func FetchAllWireguardPeers() ([]WireguardPeer, error) {
	var peers []WireguardPeer
	if err := rDB.Find(&peers).Error; err != nil {
		return nil, err
	}
	return peers, nil
}

func ExistsWireguardPeer(publicKey string) (bool, error) {
	var count int64
	if err := rDB.Model(&WireguardPeer{}).Where("public_key = ?", publicKey).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func UpdateEndpointIP(publicKey string, endpointIP string) error {
	var peer WireguardPeer
	if err := rDB.Where("public_key = ?", publicKey).First(&peer).Error; err != nil {
		return err
	}
	peer.EndpointIP = endpointIP
	if err := peer.Validate(); err != nil {
		return err
	}
	tx := rDB.Begin()
	defer tx.Rollback()
	err := tx.Save(&peer).Error
	if err != nil {
		return err
	}
	err = ConfigureWireguardPeers()
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
