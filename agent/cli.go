package main

import (
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(setConfigCmd)
	rootCmd.AddCommand(getConfig)

	setConfigCmd.Flags().String("wireguard-private-key", "", "Wireguard private key")
	setConfigCmd.Flags().String("wireguard-address", "", "Wireguard address")
	setConfigCmd.Flags().Int("wireguard-cidr", 0, "Wireguard CIDR")
	setConfigCmd.Flags().String("docker-network-gateway-address", "", "Docker network gateway address")
	setConfigCmd.Flags().String("docker-network-subnet", "", "Docker network subnet")
	setConfigCmd.Flags().Int("docker-network-cidr", 0, "Docker network CIDR")
	setConfigCmd.MarkFlagRequired("wireguard-private-key")
	setConfigCmd.MarkFlagRequired("wireguard-address")
	setConfigCmd.MarkFlagRequired("wireguard-cidr")
	setConfigCmd.MarkFlagRequired("docker-network-gateway-address")
	setConfigCmd.MarkFlagRequired("docker-network-subnet")
	setConfigCmd.MarkFlagRequired("docker-network-cidr")

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
		go startHttpServer()
		go startDnsServer()
		<-make(chan struct{})
	},
}

var setConfigCmd = &cobra.Command{
	Use: "set-config",
	Run: func(cmd *cobra.Command, args []string) {
		wireguard_cidr, err := strconv.Atoi(cmd.Flag("wireguard-cidr").Value.String())
		if err != nil {
			cmd.PrintErr("Invalid wireguard cidr")
			return
		}
		docker_cidr, err := strconv.Atoi(cmd.Flag("docker-cidr").Value.String())
		if err != nil {
			cmd.PrintErr("Invalid docker cidr")
			return
		}
		config := AgentConfig{
			WireguardConfig: WireguardConfig{
				PrivateKey: cmd.Flag("wireguard-private-key").Value.String(),
				Address:    cmd.Flag("wireguard-address").Value.String(),
				CIDR:       wireguard_cidr,
			},
			DockerNetwork: DockerNetworkConfig{
				GatewayAddress: cmd.Flag("docker-gateway-address").Value.String(),
				Subnet:         cmd.Flag("docker-subnet").Value.String(),
				CIDR:           docker_cidr,
			},
		}
		err = SetConfig(&config)
		if err != nil {
			cmd.PrintErr(err.Error())
			return
		} else {
			cmd.Println("Config updated")
		}
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
		cmd.Println("Wireguard Configuration:")
		cmd.Printf("  • Private Key -------- %s\n", config.WireguardConfig.PrivateKey)
		cmd.Printf("  • Address ------------ %s\n", config.WireguardConfig.Address)
		cmd.Printf("  • CIDR -------------- %d\n", config.WireguardConfig.CIDR)
		cmd.Println()
		cmd.Println("Docker Network Configuration:")
		cmd.Printf("  • Gateway Address ---- %s\n", config.DockerNetwork.GatewayAddress)
		cmd.Printf("  • Subnet ------------- %s\n", config.DockerNetwork.Subnet)
		cmd.Printf("  • CIDR -------------- %d\n", config.DockerNetwork.CIDR)
	},
}
