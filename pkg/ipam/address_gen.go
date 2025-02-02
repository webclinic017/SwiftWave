package ipam

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/*
IP Allocation Rules

Swiftwave will utilize a large subnet subnet for the IP allocation.
By default, we will use 10.x.x.x/8 subnet

Template format -

00001010xxxyyyyyyyyyzzzzzzzzzzzz

xxx > Reserved bits. This need to be at minimum 3 bits for operations.
- 001 - Wireguard Private Network IP
- 010 - Docker Private Network IP

y > Bits for each server. We will increase the bit count and allocate the IPs for each server.
z > Bits for container in the server. We will increase the bit count and allocate the IPs for each container.

*/

const (
	wireguardNetwork = "001"
	dockerNetwork    = "010"
)

type IPAllocationTemplate struct {
	Template                string
	ServerBitsCount         int
	ServerBitsStartIndex    int
	ServerBitsEndIndex      int
	ServerMinValue          int
	ServerMaxValue          int
	ContainerBitsCount      int
	ContainerBitsStartIndex int
	ContainerBitsEndIndex   int
	ContainerMinValue       int
	ContainerMaxValue       int
}

func parseTemplate(template string) (*IPAllocationTemplate, error) {
	if len(template) != 32 {
		return nil, errors.New("invalid template length")
	}
	// Check for the reserved bits
	if !strings.Contains(template, "xxx") {
		return nil, errors.New("reserved bits not found")
	}
	// Validate the format
	matched, _ := regexp.MatchString(`^([0|1]+)xxx(y+)(z+)$`, template)
	if !matched {
		return nil, errors.New("invalid template format")
	}

	serverBitsCount := strings.Count(template, "y")
	serverBitsStartIndex := strings.Index(template, "y")
	containerBitsCount := strings.Count(template, "z")
	containerBitsStartIndex := strings.Index(template, "z")

	return &IPAllocationTemplate{
		Template:                strings.ReplaceAll(strings.ReplaceAll(template, "y", "0"), "z", "0"),
		ServerBitsCount:         serverBitsCount,
		ServerBitsStartIndex:    serverBitsStartIndex,
		ServerBitsEndIndex:      serverBitsStartIndex + serverBitsCount,
		ServerMinValue:          1,
		ServerMaxValue:          1<<serverBitsCount - 1,
		ContainerBitsCount:      containerBitsCount,
		ContainerBitsStartIndex: containerBitsStartIndex,
		ContainerBitsEndIndex:   containerBitsStartIndex + containerBitsCount,
		ContainerMinValue:       2,
		ContainerMaxValue:       1<<containerBitsCount - 1,
	}, nil
}

func GenerateWireguardIP(template string, serverId int) (string, error) {
	t, err := parseTemplate(template)
	if err != nil {
		return "", err
	}

	// Check the server id
	if serverId > t.ServerMaxValue || serverId < t.ServerMinValue {
		return "", errors.New("server id is not in the valid range")
	}

	// Replace the reserved bits with the wireguard
	templateString := t.Template
	templateString = strings.Replace(templateString, "xxx", wireguardNetwork, 1)

	// Change last bit of the template to 1
	templateString = templateString[:len(templateString)-1] + "1"

	// Convert the server id to binary and replace the bits in the template
	serverIdBinary := fmt.Sprintf("%b", serverId)
	templateString = templateString[:(t.ServerBitsEndIndex-len(serverIdBinary))] + serverIdBinary + templateString[t.ServerBitsEndIndex:]

	return binaryFormatToIP(templateString)
}

func GenerateWireguardSubnetCIDR(template string) (int, error) {
	t, err := parseTemplate(template)
	if err != nil {
		return 0, err
	}
	return t.ServerBitsStartIndex, nil
}

func GenerateWireguardSubnet(template string) (string, error) {
	t, err := parseTemplate(template)
	if err != nil {
		return "", err
	}

	// Replace the reserved bits with the wireguard
	templateString := t.Template
	templateString = strings.Replace(templateString, "xxx", wireguardNetwork, 1)

	ip, err := binaryFormatToIP(templateString)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%d", ip, t.ServerBitsStartIndex), nil
}

func GenerateContainerWildcardSubnet(template string) (string, error) {
	t, err := parseTemplate(template)
	if err != nil {
		return "", err
	}

	// Replace the reserved bits with the wireguard
	templateString := t.Template
	templateString = strings.Replace(templateString, "xxx", dockerNetwork, 1)

	ip, err := binaryFormatToIP(templateString)
	if err != nil {
		return "", err
	}
	/*
	* Wildcard subnet is the subnet across all servers
	* That's why take the first server bits as the start index
	* to decide the container wildcard subnet
	*/
	return fmt.Sprintf("%s/%d", ip, t.ServerBitsStartIndex), nil
}

func GenerateContainerGatewayIP(template string, serverId int) (string, error) {
	t, err := parseTemplate(template)
	if err != nil {
		return "", err
	}

	// Check the server id
	if serverId > t.ServerMaxValue || serverId < t.ServerMinValue {
		return "", errors.New("server id is not in the valid range")
	}

	// Replace the reserved bits with the wireguard
	templateString := t.Template
	templateString = strings.Replace(templateString, "xxx", dockerNetwork, 1)

	// Change last bit of the template to 1
	templateString = templateString[:len(templateString)-1] + "1"

	// Convert the server id to binary and replace the bits in the template
	serverIdBinary := fmt.Sprintf("%b", serverId)
	templateString = templateString[:(t.ServerBitsEndIndex-len(serverIdBinary))] + serverIdBinary + templateString[t.ServerBitsEndIndex:]

	return binaryFormatToIP(templateString)
}

func GenerateContainerSubnet(template string, serverId int) (string, error) {
	t, err := parseTemplate(template)
	if err != nil {
		return "", err
	}

	// Check the server id
	if serverId > t.ServerMaxValue || serverId < t.ServerMinValue {
		return "", errors.New("server id is not in the valid range")
	}

	// Replace the reserved bits with the wireguard
	templateString := t.Template
	templateString = strings.Replace(templateString, "xxx", dockerNetwork, 1)

	// Convert the server id to binary and replace the bits in the template
	serverIdBinary := fmt.Sprintf("%b", serverId)
	templateString = templateString[:(t.ServerBitsEndIndex-len(serverIdBinary))] + serverIdBinary + templateString[t.ServerBitsEndIndex:]

	ip, err := binaryFormatToIP(templateString)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%d", ip, t.ContainerBitsStartIndex), nil
}

func GenerateContainerIP(template string, serverId int, containerId int) (string, error) {
	t, err := parseTemplate(template)
	if err != nil {
		return "", err
	}

	// Validate the server id
	if serverId > t.ServerMaxValue || serverId < t.ServerMinValue {
		return "", errors.New("server id is not in the valid range")
	}

	// Validate the container id
	if containerId > t.ContainerMaxValue || containerId < t.ContainerMinValue {
		return "", errors.New("container id is not in the valid range")
	}

	// Replace the reserved bits with the wireguard
	templateString := t.Template
	templateString = strings.Replace(templateString, "xxx", dockerNetwork, 1)

	// Convert the server id to binary and replace the bits in the template
	serverIdBinary := fmt.Sprintf("%b", serverId)
	templateString = templateString[:(t.ServerBitsEndIndex-len(serverIdBinary))] + serverIdBinary + templateString[t.ServerBitsEndIndex:]

	// Convert the container id to binary and replace the bits in the template
	containerIdBinary := fmt.Sprintf("%b", containerId)
	templateString = templateString[:(t.ContainerBitsEndIndex-len(containerIdBinary))] + containerIdBinary + templateString[t.ContainerBitsEndIndex:]

	return binaryFormatToIP(templateString)
}

func binaryFormatToIP(ip string) (string, error) {
	if len(ip) != 32 {
		return "", errors.New("invalid binary presentation")
	}
	// Split the template into 4 parts, each part is 8 bits
	parts := []string{
		ip[:8],
		ip[8:16],
		ip[16:24],
		ip[24:],
	}
	// Convert each binary string to decimal
	for i, part := range parts {
		decimal, err := strconv.ParseInt(part, 2, 64)
		if err != nil {
			return "", err
		}
		parts[i] = fmt.Sprintf("%d", decimal)
	}
	// Join the parts using "."
	return strings.Join(parts, "."), nil
}
