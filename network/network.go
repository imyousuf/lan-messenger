package network

import (
	"github.com/imyousuf/lan-messenger/packet"
	"github.com/imyousuf/lan-messenger/profile"
)

// Config represents configuration for the application to inform which interfaces
// to listen and broadcast to and also which port to bind to.
type Config interface {
	GetInterfaces() []string
	GetPort() int
}

type _Config struct {
	Interfaces []string
	Port       int
}

func (conf _Config) GetInterfaces() []string {
	return conf.Interfaces
}

func (conf _Config) GetPort() int {
	return conf.Port
}

// NewConfig initializes and returns a network configuration to be used for listening and
// broadcasting
func NewConfig(port int, interfaceName string) Config {
	return _Config{Port: port, Interfaces: []string{interfaceName}}
}

// MessageEvent is the event interface that MessageListener should expect
type MessageEvent interface {
	GetMessage() string
}

type iListener interface {
}

// MessageListener is the Interface that Communication accepts to notify of messages received
type MessageListener interface {
	iListener
	HandleMessageReceived(event MessageEvent)
	HandleEndOfMessages()
}

// BroadcastListener is the Interface that Communication accepts to notify of broadcasts received
type BroadcastListener interface {
	iListener
	HandleRegisterEvent(event RegisterEvent)
	HandlePingEvent(event PingEvent)
	HandleSignOffEvent(event SignOffEvent)
	HandleEndOfBroadcasts()
}

type _MessageEvent struct {
	message string
}

func (msgEvent _MessageEvent) GetMessage() string {
	return msgEvent.message
}

// Communication defines the interface the application uses to communicate between
// nodes
type Communication interface {
	SetupCommunication(config Config)
	InitCommunication(profile profile.UserProfile) error
	AddMessageListener(listener MessageListener) bool
	RemoveMessageListener(listener MessageListener) bool
	AddBroadcastListener(listener BroadcastListener) bool
	RemoveBroadcastListener(listener BroadcastListener) bool
	SendMessage(toConnectionStr string, payload packet.BasePacket)
	CloseCommunication()
}
