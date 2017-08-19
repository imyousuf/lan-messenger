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
	// Copied from https://varshneyabhi.wordpress.com/2014/12/23/simple-udp-clientserver-in-golang/
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
			log.Println("Error: ", err)
		}
	}
}

// UDPCommunication is a concrete implementation of Communication interface
type _UDPCommunication struct {
	listeners          map[string][]net.Addr
	messageChannel     chan string
	broadcastChannel   chan string
	listenError        error
	messageListeners   []MessageListener
	broadcastListeners []BroadcastListener
}

func (comm _UDPCommunication) handleRawMessages() {
	for message := range comm.messageChannel {
		event := _MessageEvent{message: message}
		for _, listener := range comm.messageListeners {
			listener.HandleMessageReceived(event)
		}
	}
	for _, listener := range comm.messageListeners {
		listener.HandleEndOfMessages()
	}
}

func (comm _UDPCommunication) handleRawBroadcasts() {
	for message := range comm.broadcastChannel {
		event := _BroadcastEvent{broadcastMessage: message}
		for _, listener := range comm.broadcastListeners {
			listener.HandleBroadcastReceived(event)
		}
	}
	for _, listener := range comm.broadcastListeners {
		listener.HandleEndOfBroadcasts()
	}
}

func (comm _UDPCommunication) AddMessageListener(listener MessageListener) bool {
	oldLen := len(comm.messageListeners)
	comm.messageListeners = append(comm.messageListeners, listener)
	return oldLen == len(comm.messageListeners)
}

func (comm _UDPCommunication) RemoveMessageListener(listener MessageListener) bool {
	oldLen := len(comm.messageListeners)
	itemIndex := -1
	for index, mListener := range comm.messageListeners {
		if mListener == listener {
			itemIndex = index
			break
		}
	}
	if itemIndex >= 0 {
		if itemIndex < oldLen-1 {
			copy(comm.messageListeners[itemIndex:], comm.messageListeners[itemIndex+1:])
			comm.messageListeners[oldLen-1] = nil
		}
		comm.messageListeners = comm.messageListeners[:oldLen-1]
	}
	return oldLen > len(comm.messageListeners)
}

func (comm _UDPCommunication) AddBroadcastListener(listener BroadcastListener) bool {
	oldLen := len(comm.broadcastListeners)
	comm.broadcastListeners = append(comm.broadcastListeners, listener)
	return oldLen == len(comm.broadcastListeners)
}

func (comm _UDPCommunication) RemoveBroadcastListener(listener BroadcastListener) bool {
	oldLen := len(comm.broadcastListeners)
	itemIndex := -1
	for index, mListener := range comm.broadcastListeners {
		if mListener == listener {
			itemIndex = index
			break
		}
	}
	if itemIndex >= 0 {
		if itemIndex < oldLen-1 {
			copy(comm.broadcastListeners[itemIndex:], comm.broadcastListeners[itemIndex+1:])
			comm.broadcastListeners[oldLen-1] = nil
		}
		comm.broadcastListeners = comm.broadcastListeners[:oldLen-1]
	}
	return oldLen > len(comm.broadcastListeners)
}

func (comm _UDPCommunication) listen(config Config) {
	port := config.GetPort()
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
			// Loop for message interfaces
			addresses := getUpIPV4Addresses(netInterface, true)
			for _, address := range addresses {
				go listenForMessage(port, &address, messageChannel)
				listeningToAddresses = append(listeningToAddresses, address)
			}
			// Loop for broadcast interfaces but listen to the next port from
			// the port requested for
			mAddresses := getUpIPV4Addresses(netInterface, false)
			for _, address := range mAddresses {
				go listenForMessage(port+1, &address, broadcastChannel)
				listeningToAddresses = append(listeningToAddresses, address)
			}
			//Add if listening to any interface
			if len(listeningToAddresses) > 0 {
				listeners[netInterface.Name] = listeningToAddresses
			}
		}
	} else {
		// Since nothing will be listened to just close them
		close(messageChannel)
		close(broadcastChannel)
	}
	comm.listeners = listeners
	comm.messageChannel = messageChannel
	comm.broadcastChannel = broadcastChannel
	comm.listenError = err
	go comm.handleRawMessages()
	go comm.handleRawBroadcasts()
}

// SetupCommunication will multicast the existence of this client to the world in an orderly
// fashion
func (comm _UDPCommunication) SetupCommunication(config Config) {
	comm.listen(config)
}

// NewUDPCommunication returns UDP implementation of communication for the application
func NewUDPCommunication() Communication {
	return _UDPCommunication{}
}
