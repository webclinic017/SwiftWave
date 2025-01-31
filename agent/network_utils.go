package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/docker/docker/api/types/network"
	"github.com/vishvananda/netlink"
	"gorm.io/gorm"
)

/*
* Every time we restart this swiftwave agent,
* There might be 5~10 seconds of downtime.
* As we are cleaning up all the interfaces, rules created by this manager
* and creating them again.
*
* It will happen also in case of any changes in the network configuration.
 */

const DockerNetworkName = "swiftwave_container_network"
const PreroutingChainName = "SWIFTWAVE_PREROUTING"
const PostroutingChainName = "SWIFTWAVE_POSTROUTING"
const ForwardChainName = "SWIFTWAVE_FORWARD"

func (c *AgentConfig) CreateDockerNetwork(db *gorm.DB) error {
	// Check if already exists
	_, err := dockerClient.NetworkInspect(context.TODO(), DockerNetworkName, network.InspectOptions{})
	if err == nil {
		return fmt.Errorf("%s already exists", DockerNetworkName)
	}
	res, err := dockerClient.NetworkCreate(context.TODO(), DockerNetworkName, network.CreateOptions{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{
				{
					Subnet:  c.DockerNetwork.Subnet,
					Gateway: c.DockerNetwork.GatewayAddress,
				},
			},
		},
		Attachable: true,
	})
	if err != nil {
		return err
	}
	c.DockerNetwork.BridgeId = fmt.Sprintf("br-%s", res.ID[:12])
	err = SetConfig(db, c)
	return err
}

func (c *AgentConfig) SetupWireguardInterface() error {
	// If it already exists, return

	return nil
}

func SetupStaticRoutes() error {
	return nil
}

func SetupIptables() error {
	return nil
}

func (s *StaticRoute) AddRoute() error {
	_, ipnet, err := net.ParseCIDR(s.Destination)
	if err != nil {
		return fmt.Errorf("invalid ip address: %s", s.Destination)
	}
	gateway := net.ParseIP(s.Gateway)
	if gateway == nil {
		return fmt.Errorf("invalid gateway ip address: %s", s.Gateway)
	}
	err = netlink.RouteAdd(&netlink.Route{
		Dst: ipnet,
		Gw:  gateway,
	})
	if err != nil {
		if strings.Contains(err.Error(), "file exists") {
			return nil
		}
		return fmt.Errorf("failed to add route: %v", err)
	}
	return nil
}

func (s *StaticRoute) RemoveRoute() error {
	_, ipnet, err := net.ParseCIDR(s.Destination)
	if err != nil {
		return fmt.Errorf("invalid ip address: %s", s.Destination)
	}
	err = netlink.RouteDel(&netlink.Route{
		Dst: ipnet,
	})
	if err != nil {
		if strings.Contains(err.Error(), "no such process") {
			return nil
		}
		return fmt.Errorf("failed to remove route: %v", err)
	}
	return nil
}

// type NetworkManager struct {
// 	DockerNetworkName       string
// 	DockerNetworkSubnet     string
// 	DockerNetworkGateway    string
// 	DockerNetworkBridgeName string

// 	WireguardInterfaceName       string
// 	WireguardListenerPort        int
// 	WireguardInterfaceAddress    string
// 	WireguardInterfaceMTU        int
// 	WireguardInterfacePrivateKey string
// 	WireguardPeers               []WireguardPeer

// 	PostRoutingChainName string
// 	PreRoutingChainName  string
// 	ForwardChainName     string
// }

// func NewNetworkManager() (*NetworkManager, error) {
// 	config, err := GetConfig()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &NetworkManager{
// 		DockerNetworkName:    "swiftwave_overlay",
// 		DockerNetworkSubnet:  fmt.Sprintf("%s/%d", config.DockerNetwork.Subnet, config.DockerNetwork.CIDR),
// 		DockerNetworkGateway: config.DockerNetwork.GatewayAddress,

// 		WireguardInterfaceName:       "swiftwave_wireguard",
// 		WireguardListenerPort:        51820,
// 		WireguardInterfaceAddress:    fmt.Sprintf("%s/%d", config.WireguardConfig.Address, config.WireguardConfig.CIDR),
// 		WireguardInterfaceMTU:        1420,
// 		WireguardInterfacePrivateKey: config.WireguardConfig.PrivateKey,
// 		WireguardPeers:               []WireguardPeer{},
// 		PreRoutingChainName:          "SWIFTWAVE_PREROUTING",
// 		PostRoutingChainName:         "SWIFTWAVE_POSTROUTING",
// 		ForwardChainName:             "SWIFTWAVE_FORWARD",
// 	}, nil
// }

// func (n *NetworkManager) Init() error {
// 	n.Cleanup() // Clean up all interface , rules created by this manager
// 	// Create docker network if not exists
// 	if err := n.CreateDockerNetwork(); err != nil {
// 		return err
// 	}
// 	// Create required iptable chains
// 	if err := n.CreateRequiredChains(); err != nil {
// 		return err
// 	}

// 	// Fetch and set docker network bridge name
// 	if err := n.FetchAndSetDockerNetworkBridgeName(); err != nil {
// 		return err
// 	}
// 	// Create wireguard interface
// 	if err := n.AddWireguardInterface(); err != nil {
// 		return err
// 	}
// 	//	Configure wireguard interface
// 	if err := n.ConfigureWireguard(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (n *NetworkManager) CreateDockerNetwork() error {
// 	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
// 	if err != nil {
// 		return err
// 	}
// 	networks, err := dockerClient.NetworkList(context.TODO(), network.ListOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	for _, network := range networks {
// 		if network.Name == n.DockerNetworkName {
// 			return nil
// 		}
// 	}
// 	_, err = dockerClient.NetworkCreate(context.TODO(), n.DockerNetworkName, network.CreateOptions{
// 		Driver: "bridge",
// 		IPAM: &network.IPAM{
// 			Driver: "default",
// 			Config: []network.IPAMConfig{
// 				{
// 					Subnet:  n.DockerNetworkSubnet,
// 					Gateway: n.DockerNetworkGateway,
// 				},
// 			},
// 		},
// 		Attachable: true,
// 	})
// 	return err
// }

// func (n *NetworkManager) FetchAndSetDockerNetworkBridgeName() error {
// 	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
// 	if err != nil {
// 		return err
// 	}
// 	network, err := dockerClient.NetworkInspect(context.TODO(), n.DockerNetworkName, network.InspectOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	n.DockerNetworkBridgeName = fmt.Sprintf("br-%s", network.ID[:12])
// 	return nil
// }

// func (n *NetworkManager) Cleanup() error {
// 	// Delete wireguard interface if exists
// 	link, err := netlink.LinkByName(n.WireguardInterfaceName)
// 	if err == nil {
// 		netlink.LinkDel(link)
// 	}

// 	// Delete docker network if no containers are using it
// 	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
// 	if err != nil {
// 		return err
// 	}
// 	networks, err := dockerClient.NetworkList(context.TODO(), network.ListOptions{})
// 	if err != nil {
// 		return err
// 	}
// 	for _, network := range networks {
// 		if network.Name == n.DockerNetworkName {
// 			if len(network.Containers) == 0 {
// 				err = dockerClient.NetworkRemove(context.TODO(), n.DockerNetworkName)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	}

// 	// Delete iptables chains
// 	ipt, err := iptables.New()
// 	if err != nil {
// 		return err
// 	}
// 	if err = ipt.ClearAndDeleteChain("nat", n.PostRoutingChainName); err != nil {
// 		return err
// 	}
// 	if err = ipt.ClearAndDeleteChain("nat", n.PreRoutingChainName); err != nil {
// 		return err
// 	}
// 	if err = ipt.ClearAndDeleteChain("filter", n.ForwardChainName); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (n *NetworkManager) CreateRequiredChains() error {
// 	ipt, err := iptables.New()
// 	if err != nil {
// 		return err
// 	}

// 	// Create POSTROUTING chain if it doesn't exist
// 	if err = ipt.NewChain("nat", n.PostRoutingChainName); err != nil {
// 		return err
// 	}
// 	if err = ipt.Append("nat", n.PostRoutingChainName, "-j", "RETURN"); err != nil {
// 		return err
// 	}
// 	// Append to main POSTROUTING chain
// 	if err = ipt.Append("nat", "POSTROUTING", "-j", n.PostRoutingChainName); err != nil {
// 		return err
// 	}

// 	// Create PREROUTING chain if it doesn't exist
// 	if err = ipt.NewChain("nat", n.PreRoutingChainName); err != nil {
// 		return err
// 	}
// 	if err = ipt.Append("nat", n.PreRoutingChainName, "-j", "RETURN"); err != nil {
// 		return err
// 	}
// 	// Append to main PREROUTING chain
// 	if err = ipt.Append("nat", "PREROUTING", "-j", n.PreRoutingChainName); err != nil {
// 		return err
// 	}

// 	// Create FORWARD chain if it doesn't exist
// 	if err = ipt.NewChain("filter", n.ForwardChainName); err != nil {
// 		return err
// 	}
// 	if err = ipt.Append("filter", n.ForwardChainName, "-j", "RETURN"); err != nil {
// 		return err
// 	}
// 	// Append to main FORWARD chain
// 	if err = ipt.Append("filter", "FORWARD", "-j", n.ForwardChainName); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (n *NetworkManager) ConfigureBridgeAndWireguardNAT() error {
// 	// Enable IP forwarding
// 	exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()

// 	// Enable bridge masquerading
// 	ipt, err := iptables.New()
// 	if err != nil {
// 		return err
// 	}
// 	if err = ipt.Append("nat", n.PostRoutingChainName, "-o", n.WireguardInterfaceName, "-j", "MASQUERADE"); err != nil {
// 		return err
// 	}
// 	if err = ipt.Append("filter", n.ForwardChainName, "-i", n.DockerNetworkBridgeName, "-o", n.WireguardInterfaceName, "-j", "ACCEPT"); err != nil {
// 		return err
// 	}
// 	if err = ipt.Append("filter", n.ForwardChainName, "-i", n.WireguardInterfaceName, "-o", n.DockerNetworkBridgeName, "-j", "ACCEPT"); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (n *NetworkManager) AddWireguardInterface() error {
// 	// Create wireguard interface
// 	wg := &netlink.Wireguard{
// 		LinkAttrs: netlink.LinkAttrs{
// 			Name: n.WireguardInterfaceName,
// 			MTU:  n.WireguardInterfaceMTU,
// 		},
// 	}
// 	err := netlink.LinkAdd(wg)
// 	if err != nil {
// 		return err
// 	}

// 	// Set interface up
// 	err = netlink.LinkSetUp(wg)
// 	if err != nil {
// 		return err
// 	}

// 	// Add IP address to wireguard interface
// 	addr, err := netlink.ParseAddr(n.WireguardInterfaceAddress)
// 	if err != nil {
// 		return err
// 	}
// 	return netlink.AddrAdd(wg, addr)
// }

// func (n *NetworkManager) ConfigureWireguard() error {
// // Create WireGuard wireguardClient
// wireguardClient, err := wgctrl.New()
// if err != nil {
// 	return fmt.Errorf("failed to create wireguard client: %v", err)
// }
// defer wireguardClient.Close()

// privateKey, err := wgtypes.ParseKey(n.WireguardInterfacePrivateKey)
// if err != nil {
// 	return err
// }

// peerConfig := []wgtypes.PeerConfig{}

// for _, peer := range n.WireguardPeers {
// 	publicKey, err := wgtypes.ParseKey(peer.PublicKey)
// 	if err != nil {
// 		return err
// 	}

// 	allowedIPs := []net.IPNet{}
// 	for _, ip := range peer.AllowedIPs {
// 		_, ipNet, err := net.ParseCIDR(ip)
// 		if err != nil {
// 			return err
// 		}
// 		allowedIPs = append(allowedIPs, *ipNet)
// 	}

// 	persistentKeepaliveDuration := time.Second * time.Duration(peer.PersistentKeepalive)

// 	peerConfig = append(peerConfig, wgtypes.PeerConfig{
// 		PublicKey:                   publicKey,
// 		AllowedIPs:                  allowedIPs,
// 		Endpoint:                    &net.UDPAddr{IP: net.ParseIP(peer.EndpointIP), Port: peer.EndpointPort},
// 		PersistentKeepaliveInterval: &persistentKeepaliveDuration,
// 		Remove:                      false,
// 	})
// }

// err = wireguardClient.ConfigureDevice(n.WireguardInterfaceName, wgtypes.Config{
// 	PrivateKey:   &privateKey,
// 	ListenPort:   &n.WireguardListenerPort,
// 	Peers:        peerConfig,
// 	ReplacePeers: true,
// })
// if err != nil {
// 	return err
// }

// // Parse wireguard interface address CIDR
// _, wgNet, err := net.ParseCIDR(n.WireguardInterfaceAddress)
// if err != nil {
// 	return err
// }

// for _, peer := range n.WireguardPeers {
// 	for _, ipStr := range peer.AllowedIPs {
// 		ip, dst, err := net.ParseCIDR(ipStr)
// 		if err != nil {
// 			return err
// 		}

// 		// Skip if IP is in same subnet as wireguard interface
// 		if wgNet.Contains(ip) {
// 			fmt.Println("Skipping IP in same subnet as wireguard interface")
// 			continue
// 		}

// 		fmt.Println("Adding route for allowed IP:", ip)

// 		// Add route for this allowed IP through wireguard interface
// 		link, err := netlink.LinkByName(n.WireguardInterfaceName)
// 		if err != nil {
// 			return err
// 		}

// 		err = netlink.RouteAdd(&netlink.Route{
// 			LinkIndex: link.Attrs().Index,
// 			Dst:       dst,
// 		})
// 		if err != nil {
// 			return err
// 		}
// 	}
// }

// 	return nil
// }
