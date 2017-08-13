package network

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func checkError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func isIPv4Address(addr net.Addr) bool {
	return strings.Contains(addr.String(), ":")
}

func filterIPv4(addresses []net.Addr, filter func(addr net.Addr) bool) []net.Addr {
	var newAddresses []net.Addr
	for _, address := range addresses {
		if !isIPv4Address(address) {
			newAddresses = append(newAddresses, address)
		}
	}
	return newAddresses
}

func getUpIPV4Addresses(netInterface net.Interface, unicast bool) []net.Addr {
	var addresses []net.Addr
	var addrsErr error
	if unicast {
		addresses, addrsErr = netInterface.Addrs()
	} else {
		addresses, addrsErr = netInterface.MulticastAddrs()
	}
	if addrsErr != nil {
		log.Fatal(addrsErr)
	} else {
		filteredAddresses := filterIPv4(addresses, isIPv4Address)
		return filteredAddresses
	}
	return make([]net.Addr, 0, 0)
}

func isInterfaceIgnorable(netInterface net.Interface) bool {
	ifaceFlags := netInterface.Flags.String()
	if !strings.Contains(ifaceFlags, "up") || strings.Contains(ifaceFlags, "loopback") {
		return true
	}
	return false
}

// PrintIPv4Addresses is a demo function
// FIXME: Remove this function
func PrintIPv4Addresses(netInterface net.Interface, unicast bool) {
	netType := "Multicast"
	if unicast {
		netType = "Unicast"
	}
	if isInterfaceIgnorable(netInterface) {
		log.Println("Flags for ", netInterface.Name, " ", netInterface.Flags.String(), ", ", netInterface.HardwareAddr)
		return
	}
	log.Println(netInterface.Name, fmt.Sprintf(" has %s addresses - ", netType), getUpIPV4Addresses(netInterface, unicast))

}

func listenForMessage(port int, address *net.Addr, channel chan string) {
	addrStr := (*address).String()
	endIndex := strings.Index(addrStr, "/")
	if endIndex < 1 {
		endIndex = len(addrStr)
	}
	serverListeningStr := addrStr[0:endIndex] + ":" + strconv.Itoa(port)
	log.Println(serverListeningStr)
	ServerAddr, err := net.ResolveUDPAddr("udp", serverListeningStr)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	checkError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		message := string(buf[0:n])
		log.Println("Received ", message, " from ", addr)
		channel <- message
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

// ListenForMessages will bind to the port in UDP and listen for messages from peers
func Listen(port int) (map[string][]net.Addr, chan string, chan string, error) {
	listeners := make(map[string][]net.Addr)
	interfaces, err := net.Interfaces()
	messageChannel := make(chan string)
	broadcastChannel := make(chan string)
	if err == nil {
		for _, netInterface := range interfaces {
			if isInterfaceIgnorable(netInterface) {
				continue
			}
			var listeningToAddresses []net.Addr
			addresses := getUpIPV4Addresses(netInterface, true)
			if len(addresses) > 0 {
				for _, address := range addresses {
					go listenForMessage(port, &address, messageChannel)
					listeningToAddresses = append(listeningToAddresses, address)
				}
			}
			mAddresses := getUpIPV4Addresses(netInterface, false)
			if len(mAddresses) > 0 {
				for _, address := range mAddresses {
					go listenForMessage(port+1, &address, broadcastChannel)
					listeningToAddresses = append(listeningToAddresses, address)
				}
			}
			listeners[netInterface.Name] = listeningToAddresses
		}
	} else {
		close(messageChannel)
		close(broadcastChannel)
		return listeners, messageChannel, broadcastChannel, err
	}
	return listeners, messageChannel, broadcastChannel, nil
}

// Broadcast will multicast the existence of this client to the world
func Broadcast() {

}
