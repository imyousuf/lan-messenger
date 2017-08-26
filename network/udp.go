package network

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/imyousuf/lan-messenger/packet"
	"github.com/imyousuf/lan-messenger/profile"
)

const (
	sessionTimeout = 5 * time.Minute
	pingInterval   = 2 * time.Minute
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

func getHostPortFromNetAddr(port int, address *net.Addr) string {
	addrStr := (*address).String()
	endIndex := strings.Index(addrStr, "/")
	if endIndex < 1 {
		endIndex = len(addrStr)
	}
	listeningStr := addrStr[0:endIndex] + ":" + strconv.Itoa(port)
	return listeningStr
}

func listenForMessage(port int, address *net.Addr, channel chan string) {
	serverListeningStr := getHostPortFromNetAddr(port, address)
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

type _ListenerConfig struct {
	port       int
	unicasts   []net.Addr
	multicasts []net.Addr
}

func (lc _ListenerConfig) GetResolvedUnicastAddr() *net.UDPAddr {
	if len(lc.unicasts) < 1 {
		return nil
	}
	udpAddr, err := net.ResolveUDPAddr("udp", getHostPortFromNetAddr(lc.port, &lc.unicasts[0]))
	if err == nil {
		return udpAddr
	}
	return nil
}

func (lc _ListenerConfig) getResolvedBroadcastReceiverAddr() *net.UDPAddr {
	if len(lc.unicasts) < 1 {
		return nil
	}
	udpAddr, err := net.ResolveUDPAddr("udp", getHostPortFromNetAddr(lc.port+2, &lc.unicasts[0]))
	if err == nil {
		return udpAddr
	}
	return nil
}

func (lc _ListenerConfig) GetMultiCastConnections() []*net.UDPConn {
	receiver := lc.getResolvedBroadcastReceiverAddr()
	if receiver == nil {
		return make([]*net.UDPConn, 0)
	}
	connections := make([]*net.UDPConn, 0, len(lc.multicasts))
	for _, mAddress := range lc.multicasts {
		udpAddr, err := net.ResolveUDPAddr("udp", getHostPortFromNetAddr(lc.port+1, &mAddress))
		if err == nil {
			conn, err := net.DialUDP("udp", receiver, udpAddr)
			if err != nil {
				log.Println("2: ", err)
				continue
			}
			connections = connections[:len(connections)+1]
			connections[len(connections)-1] = conn
		} else {
			log.Println("3: ", err)
		}
	}
	return connections
}

// UDPCommunication is a concrete implementation of Communication interface
type _UDPCommunication struct {
	listeners          map[string]_ListenerConfig
	messageChannel     chan string
	broadcastChannel   chan string
	messageListeners   []MessageListener
	broadcastListeners []BroadcastListener
	pingQuit           chan int
	selfProfile        profile.UserProfile
}

func (comm *_UDPCommunication) handleRawMessages() {
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

func (comm *_UDPCommunication) handleRawBroadcasts() {
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

func (comm *_UDPCommunication) AddMessageListener(listener MessageListener) bool {
	oldLen := len(comm.messageListeners)
	itemIndex := containsMessageListener(comm.messageListeners, listener)
	if itemIndex < 0 {
		comm.messageListeners = append(comm.messageListeners, listener)
	}
	return oldLen < len(comm.messageListeners)
}

func (comm *_UDPCommunication) RemoveMessageListener(listener MessageListener) bool {
	oldLen := len(comm.messageListeners)
	itemIndex := containsMessageListener(comm.messageListeners, listener)
	if itemIndex >= 0 {
		if itemIndex < oldLen-1 {
			copy(comm.messageListeners[itemIndex:], comm.messageListeners[itemIndex+1:])
			comm.messageListeners[oldLen-1] = nil
		}
		comm.messageListeners = comm.messageListeners[:oldLen-1]
	}
	return oldLen > len(comm.messageListeners)
}

func (comm *_UDPCommunication) AddBroadcastListener(listener BroadcastListener) bool {
	oldLen := len(comm.broadcastListeners)
	itemIndex := containsBroadcastListener(comm.broadcastListeners, listener)
	if itemIndex < 0 {
		comm.broadcastListeners = append(comm.broadcastListeners, listener)
	}
	return oldLen == len(comm.broadcastListeners)
}

func (comm *_UDPCommunication) RemoveBroadcastListener(listener BroadcastListener) bool {
	oldLen := len(comm.broadcastListeners)
	itemIndex := containsBroadcastListener(comm.broadcastListeners, listener)
	if itemIndex >= 0 {
		if itemIndex < oldLen-1 {
			copy(comm.broadcastListeners[itemIndex:], comm.broadcastListeners[itemIndex+1:])
			comm.broadcastListeners[oldLen-1] = nil
		}
		comm.broadcastListeners = comm.broadcastListeners[:oldLen-1]
	}
	return oldLen > len(comm.broadcastListeners)
}

func isListenable(netInterface net.Interface, config Config) bool {
	if len(config.GetInterfaces()) > 0 {
		isListenable := true
		for _, interfaceName := range config.GetInterfaces() {
			if netInterface.Name != interfaceName {
				isListenable = false
			}
		}
		return isListenable
	}
	return true
}

func (comm *_UDPCommunication) listen(config Config) error {
	port := config.GetPort()
	listeners := make(map[string]_ListenerConfig)
	interfaces, err := net.Interfaces()
	messageChannel := make(chan string)
	broadcastChannel := make(chan string)
	if err == nil {
		for _, netInterface := range interfaces {
			if isInterfaceIgnorable(netInterface) {
				continue
			}
			if !isListenable(netInterface, config) {
				continue
			}
			// Loop for message interfaces
			addresses := getUpIPV4Addresses(netInterface, true)
			for _, address := range addresses {
				go listenForMessage(port, &address, messageChannel)
			}
			// Loop for broadcast interfaces but listen to the next port from
			// the port requested for
			mAddresses := getUpIPV4Addresses(netInterface, false)
			for _, address := range mAddresses {
				go listenForMessage(port+1, &address, broadcastChannel)
			}
			// Add _ListenerConfig
			listeners[netInterface.Name] = _ListenerConfig{
				unicasts:   addresses,
				multicasts: mAddresses,
				port:       port}
		}
	} else {
		// Since nothing will be listened to just close them
		close(messageChannel)
		close(broadcastChannel)
	}
	comm.listeners = listeners
	comm.messageChannel = messageChannel
	comm.broadcastChannel = broadcastChannel
	go comm.handleRawMessages()
	go comm.handleRawBroadcasts()
	return err
}

func (comm _UDPCommunication) broadcastMessage(listener _ListenerConfig, message packet.BasePacket) bool {
	connections := listener.GetMultiCastConnections()
	anyError := len(connections) == 0
	for _, connection := range connections {
		buf := convertPacketToEventData(message)
		_, err := connection.Write(buf)
		if err != nil {
			anyError = true
			log.Println("1: ", err)
		}
		connection.Close()
	}
	return anyError
}

func (comm _UDPCommunication) broadcastJoin() {
	for _, listener := range comm.listeners {
		regPacket := packet.NewBuilderFactory().CreateNewSession().CreateSession(sessionTimeout).CreateUserProfile(comm.selfProfile.GetUsername(), comm.selfProfile.GetDisplayName(), comm.selfProfile.GetEmail()).RegisterDevice(listener.GetResolvedUnicastAddr().String(), 1).BuildRegisterPacket()
		anyError := true
		for anyError {
			anyError = comm.broadcastMessage(listener, regPacket)
		}
	}
}

func (comm _UDPCommunication) broadcastPing() {
	for _, listener := range comm.listeners {
		pingPacket := packet.NewBuilderFactory().Ping().RenewSession(sessionTimeout).BuildPingPacket()
		comm.broadcastMessage(listener, pingPacket)
	}
}

func (comm _UDPCommunication) setupPingBroadcast() {
	ticker := time.NewTicker(pingInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				comm.broadcastPing()
			case <-comm.pingQuit:
				ticker.Stop()
				comm.pingQuit <- 1
				return
			}
		}
	}()
}

func (comm _UDPCommunication) broadcast() error {
	log.Println("Sending initial broadcasts")
	var err error
	comm.broadcastJoin()
	comm.setupPingBroadcast()
	return err
}

func (comm *_UDPCommunication) InitCommunication(profile profile.UserProfile) error {
	comm.selfProfile = profile
	comm.pingQuit = make(chan int)
	return comm.broadcast()
}

// SetupCommunication will multicast the existence of this client to the world in an orderly
// fashion
func (comm *_UDPCommunication) SetupCommunication(config Config) {
	err := comm.listen(config)
	if err != nil {
		log.Fatal(err)
	}
}

func (comm _UDPCommunication) CloseCommunication() {
	log.Println("Closing listener channels")
	comm.pingQuit <- 1
	close(comm.messageChannel)
	close(comm.broadcastChannel)
	<-comm.pingQuit
}

// NewUDPCommunication returns UDP implementation of communication for the application
func NewUDPCommunication() Communication {
	return &_UDPCommunication{}
}
