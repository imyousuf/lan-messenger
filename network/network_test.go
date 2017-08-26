package network

import (
	"fmt"
	"net"
	"testing"
)

func printIPv4Addresses(t *testing.T, netInterface net.Interface, unicast bool) {
	netType := "Multicast"
	if unicast {
		netType = "Unicast"
	}
	if isInterfaceIgnorable(netInterface) {
		fmt.Println("Flags for ", netInterface.Name, " ", netInterface.Flags.String(), ", ", netInterface.HardwareAddr)
		return
	}
	addresses := getUpIPV4Addresses(netInterface, unicast)
	fmt.Println(netInterface.Name, fmt.Sprintf(" has %s addresses - ", netType), addresses)
	staticTestIP := "172.16.2.6"
	thatIP := net.ParseIP(staticTestIP)
	for _, address := range addresses {
		if unicast {
			_, thisIPNet, _ := net.ParseCIDR(address.String())
			fmt.Println(address.String(), " Contains ", staticTestIP, ": ", thisIPNet.Contains(thatIP))
		}
		fmt.Println(netInterface.Name, ": ", address.String(), "\t", address.Network())
	}
}

func TestInterfaces(t *testing.T) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, netInterface := range interfaces {
			// Unicast addresses
			printIPv4Addresses(t, netInterface, true)
			// Multicast addresses
			printIPv4Addresses(t, netInterface, false)
		}
	}
}
