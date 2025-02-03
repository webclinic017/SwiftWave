package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/coreos/go-iptables/iptables"
	"github.com/docker/docker/api/types/network"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
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

const FilterInputChainName = "SWIFTWAVE_FILTER_INPUT"
const FilterOutputChainName = "SWIFTWAVE_FILTER_OUTPUT"
const FilterForwardChainName = "SWIFTWAVE_FILTER_FORWARD"
const NatPreroutingChainName = "SWIFTWAVE_NAT_PREROUTING"
const NatPostroutingChainName = "SWIFTWAVE_NAT_POSTROUTING"
const NatInputChainName = "SWIFTWAVE_NAT_INPUT"
const NatOutputChainName = "SWIFTWAVE_NAT_OUTPUT"

var IPTablesClient *iptables.IPTables

func init() {
	iptablesInstance, err := iptables.New()
	if err != nil {
		panic(err)
	}
	IPTablesClient = iptablesInstance
}

// ------------- Docker Network -------------

func (c *AgentConfig) CreateDockerNetwork(remove bool) error {
	_ = dockerClient.NetworkRemove(context.TODO(), DockerNetworkName)
	// Check if already exists
	_, err := dockerClient.NetworkInspect(context.TODO(), DockerNetworkName, network.InspectOptions{})
	if err == nil {
		return fmt.Errorf("%s already exists", DockerNetworkName)
	}
	_, err = dockerClient.NetworkCreate(context.TODO(), DockerNetworkName, network.CreateOptions{
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
	err = SetConfig(c)
	return err
}

func (c *AgentConfig) SyncDockerBridge() error {
	network, err := dockerClient.NetworkInspect(context.TODO(), DockerNetworkName, network.InspectOptions{})
	if err != nil {
		return err
	}
	oldBridgeId := c.DockerNetwork.BridgeId
	newBridgeId := fmt.Sprintf("br-%s", network.ID[:12])
	c.DockerNetwork.BridgeId = newBridgeId
	err = SetConfig(c)
	if err != nil {
		return err
	}
	/*
		- iptables -A FORWARD -i br-9664fa2b25e5 -o swiftwave_wg -j ACCEPT
		- iptables -A FORWARD -i swiftwave_wg -o br-9664fa2b25e5 -j ACCEPT
	*/

	if oldBridgeId != "" {
		err = IPTablesClient.DeleteIfExists("filter", FilterForwardChainName, "-i", oldBridgeId, "-o", WireguardInterfaceName, "-j", "ACCEPT")
		if err != nil {
			return err
		}
		err = IPTablesClient.DeleteIfExists("filter", FilterForwardChainName, "-i", oldBridgeId, "-o", newBridgeId, "-j", "ACCEPT")
		if err != nil {
			return err
		}
	}

	// Add new iptable rules
	// It's important to insert the rule at the top of the chain
	// Else, other rules might be matched first and the connection will not be established
	err = IPTablesClient.InsertUnique("filter", FilterForwardChainName, 1, "-i", WireguardInterfaceName, "-o", newBridgeId, "-j", "ACCEPT")
	if err != nil {
		return err
	}
	return IPTablesClient.InsertUnique("filter", FilterForwardChainName, 1, "-i", newBridgeId, "-o", WireguardInterfaceName, "-j", "ACCEPT")

}

// ------------- Wireguard -------------

func (c *AgentConfig) RemoveWireguard() error {
	// Check if wireguard interface already exists
	link, err := netlink.LinkByName(WireguardInterfaceName)
	if err != nil {
		return nil
	}
	err = netlink.LinkDel(link)
	if err != nil {
		return fmt.Errorf("failed to delete wireguard interface: %v", err)
	}
	return nil
}

func (c *AgentConfig) SetupWireguard() error {
	// Check if wireguard interface already exists
	_, err := netlink.LinkByName(WireguardInterfaceName)
	if err == nil {
		if err := ConfigureWireguardPeers(); err != nil {
			return err
		}
		return nil
	}
	// If it already exists, return
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to create wireguard client: %v", err)
	}
	defer client.Close()
	// Create wireguard interface
	wg := &netlink.Wireguard{
		LinkAttrs: netlink.LinkAttrs{
			Name: WireguardInterfaceName,
			MTU:  1420,
		},
	}
	err = netlink.LinkAdd(wg)
	if err != nil {
		return fmt.Errorf("failed to add wireguard interface: %v", err)
	}
	// Add IP address to wireguard interface
	addr, err := netlink.ParseAddr(c.WireguardConfig.Address)
	if err != nil {
		return fmt.Errorf("failed to parse wireguard address: %v", err)
	}
	err = netlink.AddrAdd(wg, addr)
	if err != nil {
		return fmt.Errorf("failed to add wireguard address: %v", err)
	}

	// Set interface up
	err = netlink.LinkSetUp(wg)
	if err != nil {
		return fmt.Errorf("failed to set wireguard interface up: %v", err)
	}

	err = ConfigureWireguardPeers()
	if err != nil {
		return fmt.Errorf("failed to configure wireguard peers: %v", err)
	}
	return nil
}

// ------------- Static Routes -------------

func SetupStaticRoutes() {
	records, err := FetchAllStaticRoutes()
	if err != nil {
		fmt.Printf("failed to fetch static routes: %s\n", err.Error())
	}
	for _, record := range records {
		if err := record.AddRoute(); err != nil {
			fmt.Println(err.Error())
		}
	}
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

// ------------- NF Rules -------------

func SetupIptablesChains() error {
	filterChains := []string{FilterInputChainName, FilterOutputChainName, FilterForwardChainName}
	natChains := []string{NatPreroutingChainName, NatPostroutingChainName, NatInputChainName, NatOutputChainName}
	// Create filter chains
	for _, chain := range filterChains {
		exists, err := IPTablesClient.ChainExists("filter", chain)
		if err != nil {
			return fmt.Errorf("failed to check if chain exists: %v", err)
		}
		if !exists {
			err = IPTablesClient.NewChain("filter", chain)
			if err != nil {
				return fmt.Errorf("failed to create chain: %v", err)
			}
		}
	}
	// Create nat chains
	for _, chain := range natChains {
		exists, err := IPTablesClient.ChainExists("nat", chain)
		if err != nil {
			return fmt.Errorf("failed to check if chain exists: %v", err)
		}
		if !exists {
			err = IPTablesClient.NewChain("nat", chain)
			if err != nil {
				return fmt.Errorf("failed to create chain: %v", err)
			}
		}
		// Create default RETURN for the chain
		err = IPTablesClient.AppendUnique("nat", chain, "-j", "RETURN")
		if err != nil {
			return fmt.Errorf("failed to create default RETURN rule: %v", err)
		}
	}
	// Hook the chains to main chain
	if err := IPTablesClient.InsertUnique("nat", "OUTPUT", 1, "-j", NatOutputChainName); err != nil {
		return err
	}
	if err := IPTablesClient.InsertUnique("nat", "POSTROUTING", 1, "-j", NatPostroutingChainName); err != nil {
		return err
	}
	if err := IPTablesClient.InsertUnique("nat", "PREROUTING", 1, "-j", NatPreroutingChainName); err != nil {
		return err
	}
	if err := IPTablesClient.InsertUnique("nat", "INPUT", 1, "-j", NatInputChainName); err != nil {
		return err
	}
	if err := IPTablesClient.InsertUnique("filter", "FORWARD", 1, "-j", FilterForwardChainName); err != nil {
		return err
	}
	if err := IPTablesClient.InsertUnique("filter", "INPUT", 1, "-j", FilterInputChainName); err != nil {
		return err
	}
	if err := IPTablesClient.InsertUnique("filter", "OUTPUT", 1, "-j", FilterOutputChainName); err != nil {
		return err
	}
	return nil
}

func SetupIptables() error {
	rules, err := FetchAllNFRules()
	if err != nil {
		return err
	}
	for _, rule := range rules {
		err = rule.AddRule()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *NFRule) AddRule() error {
	return IPTablesClient.AppendUnique(r.Table, r.Chain, r.ArgList()...)
}

func (r *NFRule) ArgList() []string {
	var args []string
	err := json.Unmarshal([]byte(r.Args), &args)
	if err != nil {
		return []string{}
	}
	return args
}

func (r *NFRule) IsExist() (bool, error) {
	return IPTablesClient.Exists(r.Table, r.Chain, r.ArgList()...)
}

func (r *NFRule) RemoveRule() error {
	return IPTablesClient.DeleteIfExists(r.Table, r.Chain, r.ArgList()...)
}
