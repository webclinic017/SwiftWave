package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(getConfig)
	rootCmd.AddCommand(syncDockerBridge)
	rootCmd.AddCommand(dbMigrate)
	rootCmd.AddCommand(cleanup)

	setupCmd.Flags().String("wireguard-private-key", "", "Wireguard private key")
	setupCmd.Flags().String("wireguard-address", "", "Wireguard address")
	setupCmd.Flags().String("docker-network-gateway-address", "", "Docker network gateway address")
	setupCmd.Flags().String("docker-network-subnet", "", "Docker network subnet")

	setupCmd.Flags().Bool("master-node", false, "Setup as a master node")
	setupCmd.Flags().String("master-node-endpoint", "", "Master server endpoint")
	setupCmd.Flags().String("master-node-public-key", "", "Master server public key")
	setupCmd.Flags().String("master-node-allowed-ips", "", "Master server allowed ips")

	setupCmd.MarkFlagRequired("wireguard-private-key")
	setupCmd.MarkFlagRequired("wireguard-address")
	setupCmd.MarkFlagRequired("docker-network-gateway-address")
	setupCmd.MarkFlagRequired("docker-network-subnet")
}

var rootCmd = &cobra.Command{
	Use: "swiftwave-agent",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		// Pre-setup before starting the main process
		config, err := GetConfig()
		if err != nil {
			cmd.Println("Failed to fetch config")
			cmd.PrintErr(err.Error())
			return
		}
		config.SetupWireguard()
		SetupStaticRoutes()
		SetupIptables()
		// Start main process
		go startHttpServer()
		go startDnsServer()
		<-make(chan struct{})
	},
}

var setupCmd = &cobra.Command{
	Use: "setup",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := GetConfig()
		if err == nil {
			cmd.Println("Sorry, you can't change any config")
			return
		}

		isMasterNode, err := cmd.Flags().GetBool("master-node")
		if err != nil {
			cmd.PrintErr("Invalid master node flag")
			return
		}

		nodeType := WorkerNode
		if isMasterNode {
			nodeType = MasterNode
		}

		config := AgentConfig{
			NodeType: nodeType,
			WireguardConfig: WireguardConfig{
				PrivateKey: cmd.Flag("wireguard-private-key").Value.String(),
				Address:    cmd.Flag("wireguard-address").Value.String(),
			},
			DockerNetwork: DockerNetworkConfig{
				GatewayAddress: cmd.Flag("docker-network-gateway-address").Value.String(),
				Subnet:         cmd.Flag("docker-network-subnet").Value.String(),
			},
			MasterNodeConnectConfig: MasterNodeConnectConfig{
				Endpoint:   cmd.Flag("master-node-endpoint").Value.String(),
				PublicKey:  cmd.Flag("master-node-public-key").Value.String(),
				AllowedIPs: cmd.Flag("master-node-allowed-ips").Value.String(),
			},
		}

		if !isMasterNode {
			// Check MasterNodeConnectConfig endpoint
			ip := net.ParseIP(config.MasterNodeConnectConfig.Endpoint)
			if ip == nil {
				cmd.PrintErr("Invalid master node endpoint")
				return
			}
			// Check MasterNodeConnectConfig public key
			_, err := wgtypes.ParseKey(config.MasterNodeConnectConfig.PublicKey)
			if err != nil {
				cmd.PrintErr("Invalid master node public key")
				return
			}
			// Check MasterNodeConnectConfig allowed ips
			allowedIPs := strings.Split(config.MasterNodeConnectConfig.AllowedIPs, ",")
			for _, ip := range allowedIPs {
				_, _, err := net.ParseCIDR(strings.TrimSpace(ip))
				if err != nil {
					cmd.PrintErrf("Invalid master node allowed ips: %s", err.Error())
					return
				}
			}
		}

		isSuccess := false

		defer func() {
			if !isSuccess {
				_ = RemoveConfig()
				fmt.Println("Config removed")
			} else {
				fmt.Println("Config updated")
			}
		}()

		err = SetConfig(&config)
		if err != nil {
			cmd.PrintErr(err.Error())
			return
		}
		// Create docker network if it doesn't exist
		err = config.CreateDockerNetwork(true)
		if err != nil {
			cmd.PrintErr(err.Error())
			return
		}
		// Remove wireguard
		_ = config.RemoveWireguard()
		// Setup wireguard
		err = config.SetupWireguard()
		if err != nil {
			cmd.PrintErr(err.Error())
			return
		}
		isSuccess = true
	},
}

var syncDockerBridge = &cobra.Command{
	Use: "sync-docker-bridge",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if docker network bridge exists
		_, err := dockerClient.NetworkInspect(context.TODO(), DockerNetworkName, network.InspectOptions{})
		if err != nil {
			cmd.Println("Docker network bridge doesn't exist")
			cmd.PrintErr(err.Error())
			return
		}
		// Set docker network bridge id
		config, err := GetConfig()
		if err != nil {
			cmd.PrintErr(err.Error())
			return
		}
		config.DockerNetwork.BridgeId = fmt.Sprintf("br-%s", config.DockerNetwork.BridgeId[:12])
		err = SetConfig(config)
		if err != nil {
			cmd.PrintErr(err.Error())
			return
		}
		cmd.Println("Docker network bridge info synced")
	},
}

var getConfig = &cobra.Command{
	Use: "get-config",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := GetConfig()
		if err != nil {
			cmd.PrintErr(err.Error())
			return
		}
		cmd.Println("Agent Configuration:")
		cmd.Printf("  • Node Type ---------- %s\n", config.NodeType)
		if config.NodeType == WorkerNode {
			cmd.Printf("  • Connect Configuration:\n")
			cmd.Printf("    • Endpoint ---------- %s\n", config.MasterNodeConnectConfig.Endpoint)
			cmd.Printf("    • Public Key -------- %s\n", config.MasterNodeConnectConfig.PublicKey)
			cmd.Printf("    • Allowed IPs ------- %s\n", config.MasterNodeConnectConfig.AllowedIPs)
		}
		cmd.Println()
		cmd.Println("Wireguard Configuration:")
		cmd.Printf("  • Private Key -------- %s\n", config.WireguardConfig.PrivateKey)
		cmd.Printf("  • Address ------------ %s\n", config.WireguardConfig.Address)
		cmd.Println()
		cmd.Println("Docker Network Configuration:")
		cmd.Printf("  • Bridge ID ---------- %s\n", config.DockerNetwork.BridgeId)
		cmd.Printf("  • Gateway Address ---- %s\n", config.DockerNetwork.GatewayAddress)
		cmd.Printf("  • Subnet ------------- %s\n", config.DockerNetwork.Subnet)
	},
}

var cleanup = &cobra.Command{
	Use: "cleanup",
	Run: func(cmd *cobra.Command, args []string) {
		// Ask for confirmation
		fmt.Println("This will delete all containers and remove docker and wireguard networks")
		fmt.Println("Are you sure you want to continue? (y/n)")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			cmd.PrintErr(err.Error())
			return
		}
		if response != "y" {
			cmd.Println("Aborting")
			return
		}
		// Delete all containers
		containers, err := dockerClient.ContainerList(context.TODO(), container.ListOptions{})
		if err != nil {
			cmd.PrintErr(err.Error())
			cmd.Println("Failed to fetch containers")
			return
		}
		for _, c := range containers {
			if err := dockerClient.ContainerRemove(context.TODO(), c.ID, container.RemoveOptions{
				Force:         true,
				RemoveVolumes: false,
				RemoveLinks:   true,
			}); err != nil {
				cmd.PrintErr(err.Error())
				cmd.Println("Failed to remove container " + c.ID)
			}
		}
		cmd.Println("Containers removed")

		// Delete docker network
		err = dockerClient.NetworkRemove(context.TODO(), DockerNetworkName)
		if err != nil {
			cmd.PrintErr("Failed to delete docker network\n")
			cmd.PrintErr(err.Error())
			cmd.Println()
		} else {
			cmd.Println("Docker network removed")
		}
		// Delete wireguard network
		link, err := netlink.LinkByName(WireguardInterfaceName)
		if err == nil {
			err = netlink.LinkDel(link)
			if err != nil {
				cmd.PrintErr("Failed to delete wireguard interface")
				cmd.PrintErr(err.Error())
			} else {
				cmd.Println("Wireguard interface removed")
			}
		}
		// Flush all the iptables rules
		_ = IPTablesClient.ClearChain("filter", FilterInputChainName)
		_ = IPTablesClient.ClearChain("filter", FilterOutputChainName)
		_ = IPTablesClient.ClearChain("filter", FilterForwardChainName)
		_ = IPTablesClient.ClearChain("nat", NatPreroutingChainName)
		_ = IPTablesClient.ClearChain("nat", NatPostroutingChainName)
		_ = IPTablesClient.ClearChain("nat", NatInputChainName)
		_ = IPTablesClient.ClearChain("nat", NatOutputChainName)
		// Backup and remove the database file
		moveDBFilesToBackup()
		// Done
		cmd.Println("Cleanup completed")
	},
}

var dbMigrate = &cobra.Command{
	Use: "db-migrate",
	Run: func(cmd *cobra.Command, args []string) {
		// Migrate the database
		err := MigrateDatabase()
		if err != nil {
			cmd.PrintErr(err.Error())
			return
		}
		cmd.Println("Database migrated")
	},
}

func moveDBFilesToBackup() {
	// Move the `dbName` file to `dbName.bak`
	if _, err := os.Stat(dbName); err == nil {
		if err := os.Rename(dbName, dbName+".bak"); err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("Database file moved to " + dbName + ".bak")
	}
	// Also move the db-shm and db-wal files to db-shm.bak and db-wal.bak
	if _, err := os.Stat(dbName + "-shm"); err == nil {
		if err := os.Rename(dbName+"-shm", dbName+"-shm.bak"); err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("Database shm file moved to " + dbName + "-shm.bak")
	}
	if _, err := os.Stat(dbName + "-wal"); err == nil {
		if err := os.Rename(dbName+"-wal", dbName+"-wal.bak"); err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("Database wal file moved to " + dbName + "-wal.bak")
	}
}
