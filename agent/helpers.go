package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"math/rand/v2"

	"github.com/labstack/gommon/random"
	"github.com/tredoe/osutil/user/crypt"
	"github.com/tredoe/osutil/user/crypt/sha256_crypt"
)

func RunCommand(command string) (input string, output string, err error) {
	cmd := exec.Command("bash", "-c", command)
	stdoutBuffer := bytes.NewBuffer([]byte{})
	stderrBuffer := bytes.NewBuffer([]byte{})
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.Stdout = stdoutBuffer
	cmd.Stderr = stderrBuffer
	err = cmd.Run()
	return strings.TrimSpace(stdoutBuffer.String()), strings.TrimSpace(stderrBuffer.String()), err
}

func RunCommandWithoutBuffer(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func InstallToolIfNotExists(toolName string, installCmd string) error {
	_, err := exec.LookPath(toolName)
	if err != nil {
		fmt.Println("Installing " + toolName + "...")
		err = RunCommandWithoutBuffer(installCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetCPUArchitecture() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	case "386":
		return "i686"
	case "arm":
		return "arm"
	case "ppc64":
		return "ppc64"
	case "ppc64le":
		return "ppc64le"
	case "mips":
		return "mips"
	case "mipsle":
		return "mipsle"
	case "mips64":
		return "mips64"
	case "mips64le":
		return "mips64le"
	case "riscv64":
		return "riscv64"
	case "s390x":
		return "s390x"
	default:
		return "unknown"
	}
}

func GenerateBasicAuthPassword(password string) (string, error) {
	c := crypt.New(crypt.SHA256)
	s := sha256_crypt.GetSalt()
	randomSalt := random.String(5)
	saltString := fmt.Sprintf("%s%s", s.MagicPrefix, randomSalt)
	return c.Generate([]byte(password), []byte(saltString))
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

func GetServiceStatus(serviceName string) bool {
	_, _, err := RunCommand("systemctl is-active " + serviceName)
	return err == nil
}

func IsValidJSON(data string) bool {
	var js map[string]interface{}
	err := json.Unmarshal([]byte(data), &js)
	return err == nil
}

func openFileInEditor(filePath string) {
	// Check if the $EDITOR environment variable is set
	editor := os.Getenv("EDITOR")

	if editor != "" {
		// $EDITOR is set, use it to open the file
		cmd := exec.Command(editor, filePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = nil

		if err := cmd.Run(); err != nil {
			fmt.Println("Error opening file with " + editor)
		}
	} else {
		// $EDITOR is not set, try using mimeopen
		cmd := exec.Command("mimeopen", "-d", filePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = nil

		if err := cmd.Run(); err != nil {
			fmt.Println("Error opening file with mimeopen")
			fmt.Println("Set the $EDITOR environment variable to open the file with your preferred editor")
		}
	}
}
