package network

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/imyousuf/lan-messenger/packet"
	"github.com/imyousuf/lan-messenger/profile"
	"github.com/imyousuf/lan-messenger/utils"
)

const (
	sessionTimeout = 5 * time.Minute
	pingInterval   = 2 * time.Minute
)

// UDPCommunication is a concrete implementation of Communication interface
type _UDPCommunication struct {
	listeners            map[string]_ListenerConfig
	messageChannel       chan []byte
	broadcastChannel     chan []byte
	messageListeners     []MessageListener
	broadcastListeners   []BroadcastListener
	pingQuit             chan int
	selfProfile          profile.UserProfile
	sessionRegistry      map[string]_RegistryEntry
	sessionRegistryMutex *sync.Mutex
}

func (comm *_UDPCommunication) isNotDuplicate(event Event) bool {
	sessionID, packetID := event.GetEventIdentifier()
	if utils.IsStringBlank(sessionID) {
		return false
	}
	comm.sessionRegistryMutex.Lock()
	defer comm.sessionRegistryMutex.Unlock()
	if registryEntry, ok := comm.sessionRegistry[sessionID]; ok {
		if _, packetExists := registryEntry.packetRegistry[packetID]; packetExists {
			registryEntry.packetRegistry[packetID]++
			return false
		}
		registryEntry.packetRegistry[packetID] = 1
		return true
	} else if registerEvent, eventOk := event.(RegisterEvent); eventOk {
		comm.sessionRegistry[sessionID] = newRegistryEntry(registerEvent)
		return true
	} else {
		return false
	}
}

func (comm *_UDPCommunication) renewRegistryEntry(event PingEvent) {
	sessionID, _ := event.GetEventIdentifier()
	comm.sessionRegistryMutex.Lock()
	defer comm.sessionRegistryMutex.Unlock()
	if registryEntry, ok := comm.sessionRegistry[sessionID]; ok {
		registryEntry.expiryTime = event.GetPingPacket().GetExpiryTime()
	}
}

func (comm *_UDPCommunication) cleanExpiredRegistryEntries() {
	comm.sessionRegistryMutex.Lock()
	defer comm.sessionRegistryMutex.Unlock()
	for sessionID, registryEvent := range comm.sessionRegistry {
		now := time.Now()
		if registryEvent.expiryTime.Before(now) {
			delete(comm.sessionRegistry, sessionID)
		}
	}
}

func (comm *_UDPCommunication) handleRawMessages() {
	for message := range comm.messageChannel {
		event := _MessageEvent{message: string(message)}
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
		event := createEventFromEventData(message)
		if comm.isNotDuplicate(event) {
			for _, listener := range comm.broadcastListeners {
				switch event.(type) {
				case RegisterEvent:
					listener.HandleRegisterEvent(event.(RegisterEvent))
				case PingEvent:
					listener.HandlePingEvent(event.(PingEvent))
				case SignOffEvent:
					listener.HandleSignOffEvent(event.(SignOffEvent))
				default:
					panic("Event type not supported for broadcast consumption")
				}
			}
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

func (comm *_UDPCommunication) listen(config Config) error {
	port := config.GetPort()
	listeners := make(map[string]_ListenerConfig)
	interfaces, err := net.Interfaces()
	messageChannel := make(chan []byte)
	broadcastChannel := make(chan []byte)
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

func (comm _UDPCommunication) getSelfRegisterPacket(listener _ListenerConfig) packet.RegisterPacket {
	return packet.NewBuilderFactory().CreateNewSession().CreateSession(sessionTimeout).CreateUserProfile(comm.selfProfile.GetUsername(), comm.selfProfile.GetDisplayName(), comm.selfProfile.GetEmail()).RegisterDevice(listener.GetResolvedUnicastAddr().String(), 1).BuildRegisterPacket()
}

func (comm _UDPCommunication) broadcastJoin() {
	for _, listener := range comm.listeners {
		regPacket := comm.getSelfRegisterPacket(listener)
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
				comm.cleanExpiredRegistryEntries()
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
	comm.sessionRegistry = make(map[string]_RegistryEntry)
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

func (comm _UDPCommunication) findAppropriateListenerConfig(connectionStr string) _ListenerConfig {
	for _, lc := range comm.listeners {
		if lc.isCompatible(connectionStr) {
			return lc
		}
	}
	panic("No interface found for connection string: " + connectionStr)
}

func (comm _UDPCommunication) SendMessage(toConnectionStr string, payload packet.BasePacket) {
	utils.PanicableInvocation(func() {
		config := comm.findAppropriateListenerConfig(toConnectionStr)
		comm.sendMessage(config, toConnectionStr, payload)
	}, func(panicReason interface{}) {
		log.Println(panicReason)
	})
}

func (comm _UDPCommunication) sendMessage(lc _ListenerConfig, toConnectionStr string, payload packet.BasePacket) bool {
	receiver := lc.getResolvedBroadcastReceiverAddr()
	anyError := false
	udpAddr, err := net.ResolveUDPAddr("udp", toConnectionStr)
	if err == nil {
		connection, err := net.DialUDP("udp", receiver, udpAddr)
		if err != nil {
			log.Println("4: ", err)
			return anyError
		}
		buf := convertPacketToEventData(payload)
		_, err = connection.Write(buf)
		if err != nil {
			anyError = true
			log.Println("6: ", err)
		}
		defer connection.Close()
	} else {
		log.Println("5: ", err)
	}
	return anyError
}

func (comm *_UDPCommunication) addInternalListeners() {
	innerListener := _InnerListener{}
	innerListener.HandleRegisterEventMethod = func(event RegisterEvent) {
		utils.PanicableInvocation(func() {
			replyTo := event.GetRegisterPacket().GetReplyTo()
			config := comm.findAppropriateListenerConfig(replyTo)
			anyError := true
			for anyError {
				anyError = comm.sendMessage(config, replyTo, comm.getSelfRegisterPacket(config))
			}
		}, func(panicReason interface{}) {
			log.Println(panicReason)
		})
	}
	innerListener.HandlePingEventMethod = func(event PingEvent) {
		comm.renewRegistryEntry(event)
	}
	innerListener.HandleSignOffEventMethod = func(event SignOffEvent) {}
	innerListener.HandleEndOfBroadcastsMethod = func() {}
	comm.AddBroadcastListener(innerListener)
}

// NewUDPCommunication returns UDP implementation of communication for the application
func NewUDPCommunication() Communication {
	comm := &_UDPCommunication{}
	comm.sessionRegistryMutex = &sync.Mutex{}
	comm.addInternalListeners()
	return comm
}
