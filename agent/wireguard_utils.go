package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const WireguardPersistentKeepalive int = 10
const WireguardListenPort int = 51820
const WireguardPeerPort int = 51820
const WireguardInterfaceName = "swiftwave_wg"

func ConfigureWireguardPeers() error {
	// Create WireGuard wireguardClient
	wireguardClient, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to create wireguard client: %v", err)
	}
	defer wireguardClient.Close()

	config, err := GetConfig()
	if err != nil {
		return err
	}

	privateKey, err := wgtypes.ParseKey(config.WireguardConfig.PrivateKey)
	if err != nil {
		return err
	}

	peerConfig := []wgtypes.PeerConfig{}

	wireguardPeers, err := FetchAllWireguardPeers()
	if err != nil {
		return err
	}

	if config.NodeType != MasterNode {
		wireguardPeers = append(wireguardPeers, WireguardPeer{
			PublicKey:  config.MasterNodeConnectConfig.PublicKey,
			EndpointIP: config.MasterNodeConnectConfig.Endpoint,
			AllowedIPs: config.MasterNodeConnectConfig.AllowedIPs,
		})
	}

	for _, peer := range wireguardPeers {
		publicKey, err := wgtypes.ParseKey(peer.PublicKey)
		if err != nil {
			return err
		}

		allowedIPs := []net.IPNet{}
		for _, ip := range strings.Split(peer.AllowedIPs, ",") {
			_, ipNet, err := net.ParseCIDR(strings.TrimSpace(ip))
			if err != nil {
				return err
			}
			allowedIPs = append(allowedIPs, *ipNet)
		}

		persistentKeepaliveDuration := time.Second * time.Duration(WireguardPersistentKeepalive)
		var endpoint *net.UDPAddr
		if peer.EndpointIP != "" {
			endpoint = &net.UDPAddr{IP: net.ParseIP(peer.EndpointIP), Port: WireguardPeerPort}
		}

		peerConfig = append(peerConfig, wgtypes.PeerConfig{
			PublicKey:                   publicKey,
			AllowedIPs:                  allowedIPs,
			Endpoint:                    endpoint,
			PersistentKeepaliveInterval: &persistentKeepaliveDuration,
			Remove:                      false,
		})
	}

	var WireguardListenPortVar = WireguardListenPort
	return wireguardClient.ConfigureDevice(WireguardInterfaceName, wgtypes.Config{
		PrivateKey:   &privateKey,
		ListenPort:   &WireguardListenPortVar,
		Peers:        peerConfig,
		ReplacePeers: true,
	})
}
