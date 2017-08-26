package network

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"
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

func getHostPortFromNetAddr(port int, address *net.Addr) string {
	addrStr := (*address).String()
	endIndex := strings.Index(addrStr, "/")
	if endIndex < 1 {
		endIndex = len(addrStr)
	}
	listeningStr := addrStr[0:endIndex] + ":" + strconv.Itoa(port)
	return listeningStr
}

func listenForMessage(port int, address *net.Addr, channel chan []byte) {
	serverListeningStr := getHostPortFromNetAddr(port, address)
	// Copied from https://varshneyabhi.wordpress.com/2014/12/23/simple-udp-clientserver-in-golang/
	ServerAddr, err := net.ResolveUDPAddr("udp", serverListeningStr)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	checkError(err)
	defer ServerConn.Close()
	// FIXME: We will need to track for packets larger than 10KB
	buf := make([]byte, 1024*10)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		message := buf[0:n]
		log.Println("Received ", string(message), " from ", addr)
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

func (lc _ListenerConfig) isCompatible(connectionStr string) bool {
	endIndex := strings.Index(connectionStr, ":")
	if endIndex < 1 {
		endIndex = len(connectionStr)
	}
	ip := net.ParseIP(connectionStr[0:endIndex])
	for _, unicast := range lc.unicasts {
		_, thisIPNet, _ := net.ParseCIDR(unicast.String())
		if thisIPNet.Contains(ip) {
			return true
		}
	}
	return false
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

type _RegistryEntry struct {
	expiryTime     time.Time
	packetRegistry map[uint64]uint8
}

func newRegistryEntry(event RegisterEvent) _RegistryEntry {
	entry := _RegistryEntry{}
	entry.expiryTime = event.GetRegisterPacket().GetExpiryTime()
	entry.packetRegistry = make(map[uint64]uint8)
	entry.packetRegistry[event.GetRegisterPacket().GetPacketID()] = 1
	return entry
}

type _InnerListener struct {
	HandleRegisterEventMethod   func(event RegisterEvent)
	HandlePingEventMethod       func(event PingEvent)
	HandleSignOffEventMethod    func(event SignOffEvent)
	HandleEndOfBroadcastsMethod func()
}

func (il _InnerListener) HandleRegisterEvent(event RegisterEvent) {
	il.HandleRegisterEventMethod(event)
}
func (il _InnerListener) HandlePingEvent(event PingEvent) {
	il.HandlePingEventMethod(event)
}
func (il _InnerListener) HandleSignOffEvent(event SignOffEvent) {
	il.HandleSignOffEventMethod(event)
}
func (il _InnerListener) HandleEndOfBroadcasts() {
	il.HandleEndOfBroadcastsMethod()
}

func containsBroadcastListener(bList []BroadcastListener, bItem BroadcastListener) int {
	oldLen := len(bList)
	listeners := make([]iListener, oldLen, oldLen)
	for index, bListener := range bList {
		listeners[index] = bListener
	}
	return contains(listeners, bItem)
}

func containsMessageListener(mList []MessageListener, mItem MessageListener) int {
	oldLen := len(mList)
	listeners := make([]iListener, oldLen, oldLen)
	for index, mListener := range mList {
		listeners[index] = mListener
	}
	return contains(listeners, mItem)
}

func contains(list []iListener, item iListener) int {
	itemIndex := -1
	for index, iItem := range list {
		if iItem == item {
			itemIndex = index
			break
		}
	}
	return itemIndex
}
