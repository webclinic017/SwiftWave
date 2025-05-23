package cmd

import (
	_ "embed"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

//go:embed swiftwave-updater.service
var swiftwaveUpdaterService string

//go:embed swiftwave-updater.timer
var swiftwaveUpdaterTimer string

func init() {
	autoUpdateCmd.AddCommand(enableUpdateServiceCmd)
	autoUpdateCmd.AddCommand(disableUpdateServiceCmd)
}

var autoUpdateCmd = &cobra.Command{
	Use:   "auto-update",
	Short: "Check for updates each 15 min and update Swiftwave automatically",
	Run: func(cmd *cobra.Command, args []string) {
		// print help
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

var enableUpdateServiceCmd = &cobra.Command{
	Use: "enable",
	Run: func(cmd *cobra.Command, args []string) {
		// Move swiftwave-updater.service to /etc/systemd/system/
		err := os.WriteFile("/etc/systemd/system/swiftwave-updater.service", []byte(swiftwaveUpdaterService), 0644)
		if err != nil {
			printError("Failed to write swiftwave-updater.service file")
			return
		}
		// Move swiftwave-updater.timer to /etc/systemd/system/
		err = os.WriteFile("/etc/systemd/system/swiftwave-updater.timer", []byte(swiftwaveUpdaterTimer), 0644)
		if err != nil {
			printError("Failed to write swiftwave-updater.timer file")
			return
		}
		// Reload systemd daemon
		runCommand := exec.Command("systemctl", "daemon-reload")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to reload systemd daemon")
		} else {
			printSuccess("Reloaded systemd daemon")
		}
		// Enable swiftwave updater timer
		runCommand = exec.Command("systemctl", "enable", "swiftwave-updater.timer")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to enable swiftwave updater service")
		} else {
			printSuccess("Enabled swiftwave updater service")
		}
		// Start swiftwave service
		runCommand = exec.Command("systemctl", "start", "swiftwave-updater.timer")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to start swiftwave updater service")
		} else {
			printSuccess("Started swiftwave updater service")
		}
	},
}

var disableUpdateServiceCmd = &cobra.Command{
	Use: "disable",
	Run: func(cmd *cobra.Command, args []string) {
		// Stop swiftwave service
		runCommand := exec.Command("systemctl", "stop", "swiftwave-updater.timer")
		err := runCommand.Run()
		if err != nil {
			printError("Failed to stop swiftwave auto updater service")
		} else {
			printSuccess("Stopped swiftwave auto updater service")
		}
		// Disable swiftwave service
		runCommand = exec.Command("systemctl", "disable", "swiftwave-updater.timer")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to disable swiftwave auto updater service")
		} else {
			printSuccess("Disabled swiftwave auto updater service")
		}
		// Remove swiftwave-updater.service from /etc/systemd/system/
		err = os.Remove("/etc/systemd/system/swiftwave-updater.service")
		if err != nil {
			printError("Failed to remove swiftwave-updater.service file")
		}
		// Remove swiftwave-updater.timer from /etc/systemd/system/
		err = os.Remove("/etc/systemd/system/swiftwave-updater.timer")
		if err != nil {
			printError("Failed to remove swiftwave-updater.timer file")
		}
		// Reload systemd daemon
		runCommand = exec.Command("systemctl", "daemon-reload")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to reload systemd daemon")
		} else {
			printSuccess("Reloaded systemd daemon")
		}
	},
}
